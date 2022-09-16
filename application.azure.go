package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2020-04-01-preview/authorization" //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-10-01/resources"                         //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/go-uuid"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/mblaschke/devops-console/helper"
	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/formdata"
	"github.com/mblaschke/devops-console/models/response"
)

type (
	ApplicationAzure struct {
		*Server

		authorizer *autorest.Authorizer

		roleDefinitionMap map[string]string
	}

	azureRoleAssignment struct {
		PrincipalId        string
		RoleDefinitionName string
		RoleDefinitionId   string
		Description        string
	}
)

func (c *ApplicationAzure) azureAuthorizer() (*autorest.Authorizer, error) {
	if c.authorizer == nil {
		authorizer, err := auth.NewAuthorizerFromEnvironment()
		if err != nil {
			return nil, err
		}
		c.authorizer = &authorizer
	}

	return c.authorizer, nil
}

func (c *ApplicationAzure) translateRoleDefinitionNameToId(ctx context.Context, subscriptionId, name string) (*string, error) {
	if c.roleDefinitionMap == nil {
		c.roleDefinitionMap = map[string]string{}
	}

	authorizer, err := c.azureAuthorizer()
	if err != nil {
		return nil, err
	}

	roleDefinitionsClient := authorization.NewRoleDefinitionsClient(subscriptionId)
	roleDefinitionsClient.Authorizer = *authorizer

	cacheKey := fmt.Sprintf("%v:%v", subscriptionId, name)
	if _, exists := c.roleDefinitionMap[cacheKey]; !exists {
		// get role definition via API
		filter := fmt.Sprintf("roleName eq '%s'", strings.Replace(name, "'", "\\'", -1))
		result, err := roleDefinitionsClient.List(ctx, fmt.Sprintf("/subscriptions/%s", subscriptionId), filter)
		if err != nil {
			return nil, fmt.Errorf("error fetching Azure RoleDefinition: %w", err)
		}

		roleDefinitions := result.Values()
		if len(roleDefinitions) != 1 {
			return nil, fmt.Errorf("could not find Azure RoleDefinition: %v", name)
		}

		c.roleDefinitionMap[cacheKey] = *roleDefinitions[0].ID
	}

	roleDefinitionId := c.roleDefinitionMap[cacheKey]
	return &roleDefinitionId, nil
}

func (c *ApplicationAzure) removeRoleAssignmentOnScope(subscriptionId string, scopeId string, list []azureRoleAssignment) error {
	ctx := context.Background()

	authorizer, err := c.azureAuthorizer()
	if err != nil {
		return err
	}

	// translate RoleDefinition Id
	for i, roleAssignment := range list {
		roleDefinitionId, err := c.translateRoleDefinitionNameToId(
			ctx,
			subscriptionId,
			roleAssignment.RoleDefinitionName,
		)
		if err != nil {
			return fmt.Errorf("error fetching Azure RoleDefinition: %w", err)
		}
		list[i].RoleDefinitionId = *roleDefinitionId
	}

	// delete RoleAssignments on scope
	roleAssignmentsClient := authorization.NewRoleAssignmentsClient(subscriptionId)
	roleAssignmentsClient.Authorizer = *authorizer
	result, err := roleAssignmentsClient.ListForScopeComplete(ctx, scopeId, "", "")
	if err != nil {
		return fmt.Errorf("error fetching Azure RoleAssignments: %w", err)
	}

	for _, scopeRoleAssignment := range *result.Response().Value {
		for _, roleAssignment := range list {
			if *scopeRoleAssignment.PrincipalID == roleAssignment.PrincipalId && *scopeRoleAssignment.RoleDefinitionID == roleAssignment.RoleDefinitionId {
				if _, err := roleAssignmentsClient.DeleteByID(ctx, *scopeRoleAssignment.ID, ""); err != nil {
					return fmt.Errorf("unable to delete Azure RoleAssignment: %w", err)
				}
			}
		}
	}

	return nil
}

func (c *ApplicationAzure) createRoleAssignmentOnScope(subscriptionId string, scopeId string, list []azureRoleAssignment) error {
	ctx := context.Background()

	authorizer, err := c.azureAuthorizer()
	if err != nil {
		return err
	}

	// translate RoleDefinition Id
	for i, roleAssignment := range list {
		roleDefinitionId, err := c.translateRoleDefinitionNameToId(
			ctx,
			subscriptionId,
			roleAssignment.RoleDefinitionName,
		)
		if err != nil {
			return fmt.Errorf("error fetching Azure RoleDefinition: %w", err)
		}
		list[i].RoleDefinitionId = *roleDefinitionId
	}

	// create RoleAssignments on scope
	roleAssignmentsClient := authorization.NewRoleAssignmentsClient(subscriptionId)
	roleAssignmentsClient.Authorizer = *authorizer
	for _, roleAssignment := range list {
		properties := authorization.RoleAssignmentCreateParameters{
			RoleAssignmentProperties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: &roleAssignment.RoleDefinitionId,
				PrincipalID:      &roleAssignment.PrincipalId,
				Description:      to.StringPtr(roleAssignment.Description),
			},
		}

		// create uuid
		roleAssignmentId, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("unable to build UUID: %w", err)
		}
		_, err = roleAssignmentsClient.Create(ctx, scopeId, roleAssignmentId, properties)
		if err != nil {
			return fmt.Errorf("unable to create Azure RoleAssignment: %w", err)
		}
	}

	return nil
}

func (c *ApplicationAzure) ApiResourceGroupCreate(ctx iris.Context, user *models.User) {
	azureContext := context.Background()
	var group resources.Group
	validationMessages := []string{}

	formData := formdata.AzureResourceGroup{}
	err := ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	subscriptionId := opts.Azure.SubscriptionId

	if formData.Name == "" {
		validationMessages = append(validationMessages, "validation of ResourceGroup name failed (empty)")
	}

	// validate name
	if !c.config.Azure.ResourceGroup.Validation.Validate(formData.Name) {
		validationMessages = append(validationMessages, fmt.Sprintf("validation of ResourceGroup name \"%v\" failed (%v)", formData.Name, c.config.Azure.ResourceGroup.Validation.HumanizeString()))
	}

	// membership check
	if !user.IsMemberOf(formData.Team) {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("access to team \"%s\" denied", formData.Team))
		return
	}

	roleAssignmentList := []azureRoleAssignment{}
	if teamObj, err := user.GetTeam(formData.Team); err == nil {
		for _, teamRoleAssignemnt := range teamObj.AzureRoleAssignments {
			roleAssignmentList = append(roleAssignmentList, azureRoleAssignment{
				RoleDefinitionName: teamRoleAssignemnt.Role,
				PrincipalId:        teamRoleAssignemnt.PrincipalId,
			})
		}
	}

	// create ResourceGroup tagList
	tagList := map[string]*string{}

	// add tags from user
	for _, tagConfig := range c.config.Azure.ResourceGroup.Tags {
		tagValue := ""
		if val, ok := formData.Tags[tagConfig.Name]; ok {
			tagValue = val
		}

		if !tagConfig.Validation.Validate(tagValue) {
			validationMessages = append(validationMessages, fmt.Sprintf("validation of \"%s\" failed (%v)", tagConfig.Label, tagConfig.Validation.HumanizeString()))
		}

		if val := tagConfig.Transformation.Transform(tagValue); val != nil {
			tagValue = *val
		} else {
			validationMessages = append(validationMessages, fmt.Sprintf("parsing of \"%s\" failed", tagConfig.Label))
		}

		if tagValue != "" {
			tagList[tagConfig.Name] = &tagValue
		}
	}

	// fixed tags
	tagList["creator"] = to.StringPtr(user.Username)
	tagList["owner"] = to.StringPtr(formData.Team)
	tagList["updated"] = to.StringPtr(time.Now().Local().Format("2006-01-02"))
	tagList["created-by"] = to.StringPtr("devops-console")

	if len(validationMessages) >= 1 {
		c.respondError(ctx, errors.New(strings.Join(validationMessages, "\n")))
		return
	}

	// azure authorizer
	authorizer, err := c.azureAuthorizer()
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to setup Azure Authorizer: %w", err))
		return
	}

	// setup clients
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = *authorizer

	// check for existing resourcegroup
	group, _ = groupsClient.Get(azureContext, formData.Name)
	if group.ID != nil {
		tagList := []string{}

		for tagName, tagValue := range group.Tags {
			tagList = append(tagList, fmt.Sprintf("%v=%v", tagName, to.String(tagValue)))
		}

		tagLine := ""
		if len(tagList) >= 1 {
			tagLine = fmt.Sprintf(" tags:%v", tagList)
		}

		c.respondError(ctx, fmt.Errorf("failed to create already existing Azure ResourceGroup: \"%s\"%s", formData.Name, tagLine))
		return
	}

	resourceGroup := resources.Group{
		Location: to.StringPtr(formData.Location),
		Tags:     tagList,
	}

	group, err = groupsClient.CreateOrUpdate(azureContext, formData.Name, resourceGroup)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to create Azure ResourceGroup: %w", err))
		return
	}

	err = c.createRoleAssignmentOnScope(subscriptionId, *group.ID, roleAssignmentList)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to create RoleAssignments: %w", err))
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "createResourceGroup"}).Inc()

	resp := struct {
		Message    string `json:"message"`
		ResoruceId string `json:"resourceId"`
	}{
		Message:    fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name),
		ResoruceId: *group.ID,
	}

	c.notificationMessage(ctx, fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name))
	c.auditLog(ctx, fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name), 1)

	c.responseJson(ctx, resp)
}

func (c *ApplicationAzure) ApiRoleAssignmentCreate(ctx iris.Context, user *models.User) {
	c.handleRoleAssignmentAction(ctx, user, "create")
}

func (c *ApplicationAzure) ApiRoleAssignmentDelete(ctx iris.Context, user *models.User) {
	c.handleRoleAssignmentAction(ctx, user, "delete")
}

func (c *ApplicationAzure) handleRoleAssignmentAction(ctx iris.Context, user *models.User, task string) {
	azureContext := context.Background()
	var group resources.Group

	formData := formdata.AzureRoleAssignment{}
	err := ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	// resourceid
	if formData.ResourceId == "" {
		c.respondError(ctx, fmt.Errorf("no ResourceID specified"))
		return
	}

	// roledefinition
	if formData.RoleDefinition == "" {
		c.respondError(ctx, fmt.Errorf("no RoleDefinition specified"))
		return
	}
	if !stringInSlice(formData.RoleDefinition, c.config.Azure.RoleAssignment.RoleDefinitions) {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("invalid RoleDefinition specified"))
		return
	}

	// ttl
	if formData.Ttl == "" {
		c.respondError(ctx, fmt.Errorf("no TTL specified"))
		return
	}
	if !stringInSlice(formData.Ttl, c.config.Azure.RoleAssignment.Ttl) {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("invalid TTL specified"))
		return
	}

	// reason
	formData.Reason = strings.TrimSpace(formData.Reason)
	if formData.Reason == "" {
		c.respondError(ctx, fmt.Errorf("no Reason specified"))
		return
	}

	// azure authorizer
	authorizer, err := c.azureAuthorizer()
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to setup Azure Authorizer: %w", err))
		return
	}

	// parse and validate resourceid
	resourceIdInfo, err := helper.ParseResourceID(formData.ResourceId)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to parse Azure ResourceID: %w", err))
		return
	}

	if resourceIdInfo.Subscription == "" {
		c.respondError(ctx, fmt.Errorf("unable to parse subscription id, please check your resource id"))
		return
	}

	if resourceIdInfo.ResourceGroup == "" {
		c.respondError(ctx, fmt.Errorf("unable to parse resourcegroup, please check your resource id"))
		return
	}

	subscriptionId := resourceIdInfo.Subscription
	resourceGroupName := resourceIdInfo.ResourceGroup

	// setup clients
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = *authorizer

	// check for existing resourcegroup
	group, err = groupsClient.Get(azureContext, resourceGroupName)
	if err != nil {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("unable to fetch Azure ResourceGroup: %w", err))
		return
	}

	resourceGroupTags := to.StringMap(group.Tags)
	if owner, exists := resourceGroupTags["owner"]; exists {
		owner = strings.ToLower(strings.TrimSpace(owner))
		if owner == "" {
			c.respondError(ctx, fmt.Errorf("found empty owner tag in Azure ResourceGroup"))
			return
		}

		// membership check
		if !user.IsMemberOf(owner) {
			c.respondErrorWithPenalty(ctx, fmt.Errorf("access to team \"%s\" denied", owner))
			return
		}
	} else {
		c.respondError(ctx, fmt.Errorf("no owner tag found in Azure ResourceGroup"))
		return
	}

	reason := fmt.Sprintf("[ttl:%v] %v", formData.Ttl, formData.Reason)
	roleAssignmentList := []azureRoleAssignment{
		{
			PrincipalId:        user.Uuid,
			RoleDefinitionName: formData.RoleDefinition,
			Description:        reason,
		},
	}

	resp := response.GeneralMessage{}

	switch task {
	case "create":
		err = c.removeRoleAssignmentOnScope(subscriptionId, formData.ResourceId, roleAssignmentList)
		if err != nil {
			c.respondError(ctx, fmt.Errorf("unable to remove RoleAssignments: %w", err))
			return
		}

		err = c.createRoleAssignmentOnScope(subscriptionId, formData.ResourceId, roleAssignmentList)
		if err != nil {
			c.respondError(ctx, fmt.Errorf("unable to create RoleAssignments: %w", err))
			return
		}
		PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "createRoleAssignment"}).Inc()

		resp.Message = fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason)
		c.notificationMessage(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason))
		c.auditLog(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason), 1)
	case "delete":
		err = c.removeRoleAssignmentOnScope(subscriptionId, formData.ResourceId, roleAssignmentList)
		if err != nil {
			c.respondError(ctx, fmt.Errorf("unable to remove RoleAssignments: %w", err))
			return
		}
		PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "deleteRoleAssignment"}).Inc()

		resp.Message = fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" removed: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason)
		c.notificationMessage(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" removed: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason))
		c.auditLog(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" removed: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason), 1)
	default:
		c.respondError(ctx, fmt.Errorf("unable to handle RoleAssignment change: not defined"))
	}

	c.responseJson(ctx, resp)
}

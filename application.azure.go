package main

import (
	"context"
	"devops-console/models"
	"devops-console/models/formdata"
	"devops-console/models/response"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/go-uuid"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"strings"
	"time"
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

func (c *ApplicationAzure) createRoleAssignmentOnScope(subscriptionId string, scopeId string, list []azureRoleAssignment) error {
	ctx := context.Background()
	if c.roleDefinitionMap == nil {
		c.roleDefinitionMap = map[string]string{}
	}

	authorizer, err := c.azureAuthorizer()
	if err != nil {
		return err
	}

	// get RoleDefinition Id
	roleDefinitionsClient := authorization.NewRoleDefinitionsClient(subscriptionId)
	roleDefinitionsClient.Authorizer = *authorizer
	for i, roleAssignment := range list {
		roleDefintionName := roleAssignment.RoleDefinitionName
		if val, exists := c.roleDefinitionMap[roleDefintionName]; exists {
			// use cached value
			list[i].RoleDefinitionId = val
		} else {
			// get role definition via API
			filter := fmt.Sprintf("roleName eq '%s'", strings.Replace(roleDefintionName, "'", "\\'", -1))
			result, err := roleDefinitionsClient.List(ctx, fmt.Sprintf("/subscriptions/%s", subscriptionId), filter)
			if err != nil {
				return fmt.Errorf("error fetching Azure RoleDefinition: %v", err)
			}

			roleDefinitions := result.Values()
			if len(roleDefinitions) != 1 {
				return fmt.Errorf("could not find Azure RoleDefinition: %v", roleDefintionName)
			}

			list[i].RoleDefinitionId = *roleDefinitions[0].ID
			c.roleDefinitionMap[roleAssignment.RoleDefinitionName] = *roleDefinitions[0].ID
		}
	}

	// create RoleAssignments on scope
	roleAssignmentsClient := authorization.NewRoleAssignmentsClient(subscriptionId)
	roleAssignmentsClient.Authorizer = *authorizer
	for _, roleAssignment := range list {
		properties := authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: &roleAssignment.RoleDefinitionId,
				PrincipalID:      &roleAssignment.PrincipalId,
			},
		}

		// create uuid
		roleAssignmentId, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("unable to build UUID: %v", err)
		}
		_, err = roleAssignmentsClient.Create(ctx, scopeId, roleAssignmentId, properties)
		if err != nil {
			return fmt.Errorf("unable to create Azure RoleAssignment: %v", err)
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

	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")

	if formData.Name == "" {
		validationMessages = append(validationMessages, "validation of ResourceGroup name failed (empty)")
	}

	// validate name
	if !c.config.Azure.ResourceGroup.Validation.Validate(formData.Name) {
		validationMessages = append(validationMessages, fmt.Sprintf("validation of ResourceGroup name \"%v\" failed (%v)", formData.Name, c.config.Azure.ResourceGroup.Validation.HumanizeString()))
	}

	// membership check
	if !user.IsMemberOf(formData.Team) {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("access to team \"%s\" denied", err))
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
		if val, ok := formData.Tag[tagConfig.Name]; ok {
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
		c.respondError(ctx, fmt.Errorf("unable to setup Azure Authorizer: %v", err))
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
		c.respondError(ctx, fmt.Errorf("unable to create Azure ResourceGroup: %v", err))
		return
	}

	err = c.createRoleAssignmentOnScope(subscriptionId, *group.ID, roleAssignmentList)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to create RoleAssignments: %v", err))
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "createResourceGroup"}).Inc()

	resp := response.GeneralMessage{}

	resp.Message = fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name)
	c.notificationMessage(ctx, fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name))
	c.auditLog(ctx, fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name), 1)

	c.responseJson(ctx, resp)
}

func (c *ApplicationAzure) ApiRoleAssignmentCreate(ctx iris.Context, user *models.User) {
	azureContext := context.Background()
	var group resources.Group

	formData := formdata.AzureRoleAssignment{}
	err := ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	// azure authorizer
	authorizer, err := c.azureAuthorizer()
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to setup Azure Authorizer: %v", err))
		return
	}

	resourceIdInfo, err := azure.ParseResourceID(formData.ResourceId)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to parse Azure ResourceID: %v", err))
		return
	}

	if resourceIdInfo.SubscriptionID == "" {
		c.respondError(ctx, fmt.Errorf("unable to parse subscription id, please check your resource id"))
		return
	}

	if resourceIdInfo.ResourceGroup == "" {
		c.respondError(ctx, fmt.Errorf("unable to parse subscription id, please check your resource id"))
		return
	}

	subscriptionId := resourceIdInfo.SubscriptionID
	resourceGroupName := resourceIdInfo.ResourceGroup

	// setup clients
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = *authorizer

	// check for existing resourcegroup
	group, err = groupsClient.Get(azureContext, resourceGroupName)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to fetch Azure ResourceGroup: %v", err))
		return
	}

	if owner, exists := group.Tags["owner"]; exists {
		if owner == nil {
			c.respondError(ctx, fmt.Errorf("found empty owner tag in Azure ResourceGroup"))
			return
		}

		// membership check
		if !user.IsMemberOf(*owner) {
			c.respondErrorWithPenalty(ctx, fmt.Errorf("access to team \"%s\" denied", err))
			return
		}
	} else {
		c.respondError(ctx, fmt.Errorf("no owner tag found in Azure ResourceGroup"))
		return
	}

	roleAssignmentList := []azureRoleAssignment{
		{
			PrincipalId:        user.Uuid,
			RoleDefinitionName: formData.RoleDefinition,
		},
	}

	err = c.createRoleAssignmentOnScope(subscriptionId, formData.ResourceId, roleAssignmentList)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to create RoleAssignments: %v", err))
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "createRoleAssignment"}).Inc()

	resp := response.GeneralMessage{}

	resp.Message = fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason)
	c.notificationMessage(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason))
	c.auditLog(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason), 1)

	c.responseJson(ctx, resp)
}

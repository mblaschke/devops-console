package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	armauthorization "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	armresources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/hashicorp/go-uuid"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/azuresdk/armclient"
	"github.com/webdevops/go-common/utils/to"

	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/formdata"
	"github.com/mblaschke/devops-console/models/response"
)

type (
	ApplicationAzure struct {
		*Server

		armClient *armclient.ArmClient

		roleDefinitionMap map[string]string
	}

	azureRoleAssignment struct {
		PrincipalId        string
		RoleDefinitionName string
		RoleDefinitionId   string
		Description        string
	}
)

func NewApplicationAzure(c *Server) *ApplicationAzure {
	app := ApplicationAzure{Server: c}

	armClient, err := armclient.NewArmClientWithCloudName(opts.Azure.Environment, c.logger)
	if err != nil {
		log.Panic(err.Error())
	}

	armClient.SetUserAgent(UserAgent + gitTag)

	app.armClient = armClient
	return &app
}

func (c *ApplicationAzure) translateRoleDefinitionNameToId(ctx context.Context, subscriptionId, name string) (*string, error) {
	if c.roleDefinitionMap == nil {
		c.roleDefinitionMap = map[string]string{}
	}

	client, err := armauthorization.NewRoleDefinitionsClient(c.armClient.GetCred(), c.armClient.NewArmClientOptions())
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%v:%v", subscriptionId, strings.ToLower(name))

	if _, exists := c.roleDefinitionMap[cacheKey]; !exists {
		// update roledefintionlist
		scope := fmt.Sprintf("/subscriptions/%v", subscriptionId)
		pager := client.NewListPager(scope, nil)
		for pager.More() {
			result, err := pager.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			for _, roleDefinition := range result.Value {
				key := fmt.Sprintf("%v:%v", subscriptionId, to.StringLower(roleDefinition.Properties.RoleName))

				c.roleDefinitionMap[key] = *roleDefinition.ID
			}
		}

	}

	if roleDefinitionId, exists := c.roleDefinitionMap[cacheKey]; exists {
		return &roleDefinitionId, nil
	} else {
		return nil, fmt.Errorf(`unable to find RoleDefintion "%v" for subscription %v`, name, subscriptionId)
	}

}

func (c *ApplicationAzure) removeRoleAssignmentOnScope(scopeId string, list []azureRoleAssignment) error {
	ctx := context.Background()

	scopeInfo, err := armclient.ParseResourceId(scopeId)
	if err != nil {
		return err
	}

	client, err := armauthorization.NewRoleAssignmentsClient(scopeInfo.Subscription, c.armClient.GetCred(), c.armClient.NewArmClientOptions())
	if err != nil {
		return err
	}

	// translate RoleDefinition Id
	for i, roleAssignment := range list {
		roleDefinitionId, err := c.translateRoleDefinitionNameToId(
			ctx,
			scopeInfo.Subscription,
			roleAssignment.RoleDefinitionName,
		)
		if err != nil {
			return fmt.Errorf("error fetching Azure RoleDefinition: %w", err)
		}
		list[i].RoleDefinitionId = *roleDefinitionId
	}

	pager := client.NewListForScopePager(scopeId, nil)
	for pager.More() {
		result, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, existingRoleAssignment := range result.Value {
			for _, roleAssignment := range list {
				if *existingRoleAssignment.Properties.PrincipalID == roleAssignment.PrincipalId && *existingRoleAssignment.Properties.RoleDefinitionID == roleAssignment.RoleDefinitionId {
					_, err := client.DeleteByID(ctx, *existingRoleAssignment.ID, nil)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *ApplicationAzure) createRoleAssignmentOnScope(scopeId string, list []azureRoleAssignment) error {
	ctx := context.Background()

	scopeInfo, err := armclient.ParseResourceId(scopeId)
	if err != nil {
		return err
	}

	client, err := armauthorization.NewRoleAssignmentsClient(scopeInfo.Subscription, c.armClient.GetCred(), c.armClient.NewArmClientOptions())
	if err != nil {
		return err
	}

	// translate RoleDefinition Id
	for i, roleAssignment := range list {
		roleDefinitionId, err := c.translateRoleDefinitionNameToId(ctx, scopeInfo.Subscription, roleAssignment.RoleDefinitionName)
		if err != nil {
			return fmt.Errorf("error fetching Azure RoleDefinition: %w", err)
		}
		list[i].RoleDefinitionId = *roleDefinitionId
	}

	// create RoleAssignments on scope
	for _, row := range list {
		roleAssignment := row

		// create uuid
		roleAssignmentId, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("unable to build UUID: %w", err)
		}

		roleAssignmentProperties := armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				RoleDefinitionID: &roleAssignment.RoleDefinitionId,
				PrincipalID:      &roleAssignment.PrincipalId,
				Description:      to.StringPtr(roleAssignment.Description),
			},
		}

		_, err = client.Create(ctx, scopeId, roleAssignmentId, roleAssignmentProperties, nil)
		if err != nil {
			return fmt.Errorf("unable to create Azure RoleAssignment: %w", err)
		}
	}

	return nil
}

func (c *ApplicationAzure) ApiResourceGroupCreate(ctx iris.Context, user *models.User) {
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
	if !c.config.Azure.ResourceGroup.Filter.Name.Validate(formData.Name) {
		validationMessages = append(validationMessages, fmt.Sprintf("validation of ResourceGroup name \"%v\" failed (%v)", formData.Name, c.config.Azure.ResourceGroup.Filter.Name.String()))
	}

	// membership check
	if !user.IsMemberOf(formData.Team) {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("access to team \"%s\" denied", formData.Team))
		return
	}

	roleAssignmentList := []azureRoleAssignment{}
	if teamObj, err := user.GetTeam(formData.Team); err == nil {
		if teamObj.Azure.Group != nil {
			roleAssignmentList = append(roleAssignmentList, azureRoleAssignment{
				RoleDefinitionName: c.config.Azure.ResourceGroup.RoleDefinitionName,
				PrincipalId:        *teamObj.Azure.Group,
			})
		}

		if teamObj.Azure.ServicePrincipal != nil {
			roleAssignmentList = append(roleAssignmentList, azureRoleAssignment{
				RoleDefinitionName: c.config.Azure.ResourceGroup.RoleDefinitionName,
				PrincipalId:        *teamObj.Azure.ServicePrincipal,
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
	client, err := armresources.NewResourceGroupsClient(subscriptionId, c.armClient.GetCred(), c.armClient.NewArmClientOptions())
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to create Azure Client: %w", err))
		return
	}

	// check for existing resourcegroup
	existingResourceGroup, _ := client.Get(ctx, formData.Name, nil)
	if existingResourceGroup.ID != nil {
		tagList := []string{}

		for tagName, tagValue := range existingResourceGroup.Tags {
			tagList = append(tagList, fmt.Sprintf("%v=%v", tagName, to.String(tagValue)))
		}

		tagLine := ""
		if len(tagList) >= 1 {
			tagLine = fmt.Sprintf(" tags:%v", tagList)
		}

		c.respondError(ctx, fmt.Errorf("failed to create already existing Azure ResourceGroup: \"%s\"%s", formData.Name, tagLine))
		return
	}

	resourceGroup := armresources.ResourceGroup{
		Location: to.StringPtr(formData.Location),
		Tags:     tagList,
	}

	createdResourceGroup, err := client.CreateOrUpdate(ctx, formData.Name, resourceGroup, nil)
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to create Azure ResourceGroup: %w", err))
		return
	}

	err = c.createRoleAssignmentOnScope(*createdResourceGroup.ID, roleAssignmentList)
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
		ResoruceId: *createdResourceGroup.ID,
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

	if !c.config.Azure.RoleAssignment.Filter.ResourceId.Validate(formData.ResourceId) {
		c.respondError(ctx, fmt.Errorf("resource id not allowed (%v)", c.config.Azure.RoleAssignment.Filter.ResourceId.String()))
		return
	}

	// parse and validate resourceid
	resourceIdInfo, err := armclient.ParseResourceId(formData.ResourceId)
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
	client, err := armresources.NewResourceGroupsClient(subscriptionId, c.armClient.GetCred(), c.armClient.NewArmClientOptions())
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to create Azure Client: %w", err))
		return
	}

	// check for existing resourcegroup
	group, err := client.Get(ctx, resourceGroupName, nil)
	if err != nil {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("unable to fetch Azure ResourceGroup: %w", err))
		return
	}

	ownerList := []string{}

	// build owner list (from owner tag)
	resourceGroupTags := to.StringMap(group.Tags)
	if owner, exists := resourceGroupTags["owner"]; exists {
		owner = strings.ToLower(strings.TrimSpace(owner))
		if owner == "" {
			c.respondError(ctx, fmt.Errorf("found empty owner tag in Azure ResourceGroup"))
			return
		}

		ownerList = append(ownerList, owner)
	} else {
		c.respondError(ctx, fmt.Errorf("no owner tag found in Azure ResourceGroup"))
		return
	}

	// build owner list (from jitaccess tag)
	if val, exists := resourceGroupTags["jitaccess"]; exists {
		valList := strings.Split(val, ",")
		for _, owner := range valList {
			owner := strings.TrimSpace(owner)
			if owner == "" {
				continue
			}
			ownerList = append(ownerList, owner)
		}
	}

	scopeAccess := false
	for _, owner := range ownerList {
		// membership check
		if user.IsMemberOf(owner) {
			scopeAccess = true
		}
	}

	if !scopeAccess {
		c.respondError(ctx, fmt.Errorf("access to Azure ResourceGroup denied (owners: %v)", strings.Join(ownerList, ", ")))
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
		err = c.removeRoleAssignmentOnScope(formData.ResourceId, roleAssignmentList)
		if err != nil {
			c.respondError(ctx, fmt.Errorf("unable to remove RoleAssignments: %w", err))
			return
		}

		err = c.createRoleAssignmentOnScope(formData.ResourceId, roleAssignmentList)
		if err != nil {
			c.respondError(ctx, fmt.Errorf("unable to create RoleAssignments: %w", err))
			return
		}
		PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "createRoleAssignment"}).Inc()

		resp.Message = fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason)
		c.notificationMessage(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason))
		c.auditLog(ctx, fmt.Sprintf("Azure RoleAssignment for \"%s\" with role \"%s\" by \"%s\" created: %v", formData.ResourceId, formData.RoleDefinition, user.Username, formData.Reason), 1)
	case "delete":
		err = c.removeRoleAssignmentOnScope(formData.ResourceId, roleAssignmentList)
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

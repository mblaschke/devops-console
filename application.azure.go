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
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/go-uuid"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"strings"
	"sync"
	"time"
)

type ApplicationAzure struct {
	*Server
}

func (c *ApplicationAzure) ApiResourceGroupCreate(ctx iris.Context, user *models.User) {
	wg := sync.WaitGroup{}

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

	roleAssignmentList := []models.TeamAzureRoleAssignments{}
	roleAssignmentList = append(roleAssignmentList, models.TeamAzureRoleAssignments{
		Role:        "Owner",
		PrincipalId: user.Uuid,
	})

	// membership check
	if !user.IsMemberOf(formData.Team) {
		c.respondErrorWithPenalty(ctx, fmt.Errorf("access to team \"%s\" denied", err))
		return
	}

	if teamObj, err := user.GetTeam(formData.Team); err == nil {
		roleAssignmentList = append(roleAssignmentList, teamObj.AzureRoleAssignments...)
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
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		c.respondError(ctx, fmt.Errorf("unable to setup Azure Authorizer: %v", err))
		return
	}

	// setup clients
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = authorizer

	roleDefinitionsClient := authorization.NewRoleDefinitionsClient(subscriptionId)
	roleDefinitionsClient.Authorizer = authorizer

	roleAssignmentsClient := authorization.NewRoleAssignmentsClient(subscriptionId)
	roleAssignmentsClient.Authorizer = authorizer

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

	// translate role lookup
	roleAssignmentChannel := make(chan models.TeamAzureRoleAssignments, len(roleAssignmentList))
	for _, roleAssignment := range roleAssignmentList {
		wg.Add(1)
		go func(roleAssignment models.TeamAzureRoleAssignments) {
			defer wg.Done()

			// get role definition
			filter := fmt.Sprintf("roleName eq '%s'", roleAssignment.Role)
			roleDefinitions, err := roleDefinitionsClient.List(azureContext, "", filter)

			if len(roleDefinitions.Values()) != 1 {
				c.respondError(ctx, fmt.Errorf("error generating UUID for Azure RoleAssignment: %v", err))
				return
			}

			roleAssignment.Role = *roleDefinitions.Values()[0].ID

			roleAssignmentChannel <- roleAssignment
		}(roleAssignment)
	}
	wg.Wait()

	close(roleAssignmentChannel)
	roleAssignmentList = []models.TeamAzureRoleAssignments{}
	for roleAssignment := range roleAssignmentChannel {
		roleAssignmentList = append(roleAssignmentList, roleAssignment)
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

	// assign role to ResourceGroup
	for _, roleAssignment := range roleAssignmentList {
		wg.Add(1)
		go func(roleAssignment models.TeamAzureRoleAssignments) {
			defer wg.Done()
			// assign role to ResourceGroup
			properties := authorization.RoleAssignmentCreateParameters{
				Properties: &authorization.RoleAssignmentProperties{
					RoleDefinitionID: &roleAssignment.Role,
					PrincipalID:      &roleAssignment.PrincipalId,
				},
			}

			// create uuid
			roleAssignmentId, err := uuid.GenerateUUID()
			if err != nil {
				c.respondError(ctx, fmt.Errorf("unable to build UUID: %v", err))
				return
			}

			_, err = roleAssignmentsClient.Create(azureContext, to.String(group.ID), roleAssignmentId, properties)
			if err != nil {
				c.respondError(ctx, fmt.Errorf("unable to create Azure RoleAssignment: %v", err))
				return
			}
		}(roleAssignment)
	}
	wg.Wait()

	PrometheusActions.With(prometheus.Labels{"scope": "azure", "type": "createResourceGroup"}).Inc()

	resp := response.GeneralMessage{}

	resp.Message = fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name)
	c.notificationMessage(ctx, fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name))
	c.auditLog(ctx, fmt.Sprintf("Azure ResourceGroup \"%s\" created", formData.Name), 1)

	c.responseJson(ctx, resp)
}

package main

import (
	"context"
	"devops-console/models"
	"devops-console/models/formdata"
	"devops-console/models/response"
	"devops-console/services"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	coreV1 "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	settingsV1alpha1 "k8s.io/api/settings/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
)

type ApplicationKubernetes struct {
	*Server
}

func (c *ApplicationKubernetes) serviceKubernetes() (service *services.Kubernetes) {
	service = &services.Kubernetes{}

	if c.config.App.Kubernetes.Namespace.Filter.Access != "" {
		service.Filter.Namespace = regexp.MustCompile(c.config.App.Kubernetes.Namespace.Filter.Access)
	}

	return
}

func (c *ApplicationKubernetes) Kubeconfig(ctx iris.Context, user *models.User) {
	ret := c.config.Settings.Kubeconfig
	c.responseJson(ctx, ret)
}

func (c *ApplicationKubernetes) KubeconfigDownload(ctx iris.Context, user *models.User) {
	name := ctx.Params().GetString("name")

	if val, ok := c.config.Settings.Kubeconfig[name]; ok {
		PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "downloadKubeconfig"}).Inc()

		ctx.Header("ContentType", "text/yaml")
		ctx.Header("Content-Disposition", "attachment; filename=\"kubeconfig.yaml\"")
		if _, err := ctx.Binary([]byte(val.Content)); err != nil {
			c.logger.Errorln(err)
		}
	} else {
		c.respondError(ctx, errors.New("Kubeconfig name not valid"))
	}
}

func (c *ApplicationKubernetes) ApiCluster(ctx iris.Context, user *models.User) {
	service := c.serviceKubernetes()
	nodes, err := service.Nodes()
	if err != nil {
		c.respondError(ctx, errors.New("Unable to contact Kubernetes cluster"))
		return
	}

	ret := []response.KubernetesCluster{}

	for _, node := range nodes.Items {
		row := response.KubernetesCluster{
			Name:              node.Name,
			Version:           node.Status.NodeInfo.KubeletVersion,
			SpecMachineCPU:    node.Status.Capacity.Cpu().String(),
			SpecMachineMemory: humanize.Bytes(uint64(node.Status.Capacity.Memory().Value())),
			Status:            fmt.Sprintf("%v", node.Status.Phase),
			Created:           node.CreationTimestamp.UTC().String(),
			CreatedAgo:        humanize.Time(node.CreationTimestamp.UTC()),
		}

		for _, val := range node.Status.Conditions {
			if val.Reason == "KubeletReady" {
				row.Status = fmt.Sprintf("%v", val.Type)
			}
		}

		for _, item := range node.Status.Addresses {
			if item.Type == "InternalIP" {
				row.InternalIp = item.Address
			}
		}

		if val, ok := node.Labels["kubernetes.io/role"]; ok {
			row.Role = val
		}

		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			row.Role = "master"
		}

		if val, ok := node.Labels["kubernetes.io/arch"]; ok {
			row.SpecArch = val
		}

		if val, ok := node.Labels["kubernetes.io/os"]; ok {
			row.SpecOS = val
		}

		if val, ok := node.Labels["failure-domain.beta.kubernetes.io/region"]; ok {
			row.SpecRegion = val
		}

		if val, ok := node.Labels["failure-domain.beta.kubernetes.io/zone"]; ok {
			row.SpecZone = val
		}

		if val, ok := node.Labels["beta.kubernetes.io/instance-type"]; ok {
			row.SpecInstance = val
		}

		ret = append(ret, row)
	}

	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "listCluster"}).Inc()

	c.responseJson(ctx, ret)
}

func (c *ApplicationKubernetes) ApiNamespaceList(ctx iris.Context, user *models.User) {
	nsList, err := c.serviceKubernetes().NamespaceList()
	if err != nil {
		c.respondError(ctx, errors.New("Unable to contact Kubernetes cluster"))
		return
	}

	ret := []response.KubernetesNamespace{}

	for _, namespaceNative := range nsList {
		namespace := models.KubernetesNamespace{Namespace: &namespaceNative}

		if !c.kubernetesNamespaceAccessAllowed(ctx, namespace) {
			continue
		}

		namespaceParts := strings.Split(namespace.Name, "-")
		environment := ""
		if len(namespaceParts) > 2 {
			environment = namespaceParts[0]
		}

		row := response.KubernetesNamespace{
			Name:        namespace.Name,
			Environment: environment,
			Status:      fmt.Sprintf("%v", namespace.Status.Phase),
			Created:     namespace.CreationTimestamp.UTC().String(),
			CreatedAgo:  humanize.Time(namespace.CreationTimestamp.UTC()),
			Deleteable:  c.kubernetesNamespaceDeleteAllowed(ctx, &namespace),
			Settings:    namespace.SettingsExtract(c.config.Kubernetes),
		}

		if opts.EnableNamespacePodCount {
			row.PodCount = c.serviceKubernetes().NamespacePodCount(namespace.Name)
		}

		if val, ok := namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.Description]; ok {
			row.Description = val
		}

		if val, ok := namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.NetworkPolicy]; ok {
			row.NetworkPolicy = val
		}

		if val, ok := namespace.Labels["team"]; ok {
			row.OwnerTeam = val
		}

		if val, ok := namespace.Labels["user"]; ok {
			row.OwnerUser = val
		}

		ret = append(ret, row)
	}

	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "listNamespace"}).Inc()

	c.responseJson(ctx, ret)
}

func (c *ApplicationKubernetes) ApiNamespaceCreate(ctx iris.Context, user *models.User) {
	var namespaceName string
	var kubernetesEnvironment *models.AppConfigKubernetesEnvironment

	formData := formdata.KubernetesNamespaceCreate{}
	err := ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	if formData.Settings == nil || formData.Description == nil || formData.App == nil || formData.Environment == nil || formData.Team == nil {
		c.respondError(ctx, errors.New("Invalid form data"))
		return
	}

	username := user.Username

	if !regexp.MustCompile(c.config.App.Kubernetes.Namespace.Validation.App).MatchString(*formData.App) {
		c.respondError(ctx, errors.New("Invalid app value"))
		return
	}

	labels := map[string]string{
		c.config.App.Kubernetes.Namespace.Labels.Environment: *formData.Environment,
	}

	// validation
	nsSettings, validationMessages := c.validateSettings(*formData.Settings)
	if len(validationMessages) >= 1 {
		c.respondError(ctx, errors.New(strings.Join(validationMessages, "\n")))
		return
	}

	// check if environment is allowed
	environmentAllowed := false
	for _, env := range c.config.App.Kubernetes.Environments {
		if env.Name == *formData.Environment {
			environmentAllowed = true
			kubernetesEnvironment = &env
			break
		}
	}
	if !environmentAllowed || kubernetesEnvironment == nil {
		c.respondError(ctx, errors.New(fmt.Sprintf("Environment \"%s\" not allowed in this cluster", *formData.Environment)))
		return
	}

	// team filter check
	if !regexp.MustCompile(c.config.App.Kubernetes.Namespace.Validation.Team).MatchString(*formData.Team) {
		c.respondError(ctx, errors.New("Invalid team value"))
		return
	}

	// membership check
	if !user.IsMemberOf(*formData.Team) {
		c.respondErrorWithPenalty(ctx, errors.New(fmt.Sprintf("Access to team \"%s\" denied", *formData.Team)))
		return
	}

	// quota check
	switch kubernetesEnvironment.Quota {
	case "team":
		// quota check
		if err := c.checkNamespaceTeamQuota(*formData.Team); err != nil {
			c.respondError(ctx, errors.New(fmt.Sprintf("Error: %v", err)))
			return
		}
	case "user":
		// quota check
		if err := c.checkNamespaceUserQuota(username); err != nil {
			c.respondError(ctx, errors.New(fmt.Sprintf("Error: %v", err)))
			return
		}

		labels[c.config.App.Kubernetes.Namespace.Labels.User] = strings.ToLower(username)
	}

	// build namespace name
	namespaceName = kubernetesEnvironment.Template
	namespaceName = strings.Replace(namespaceName, "{env}", kubernetesEnvironment.Name, -1)
	namespaceName = strings.Replace(namespaceName, "{user}", username, -1)
	namespaceName = strings.Replace(namespaceName, "{team}", *formData.Team, -1)
	namespaceName = strings.Replace(namespaceName, "{app}", *formData.App, -1)

	// namespace filtering
	namespaceName = strings.ToLower(namespaceName)
	namespaceName = strings.Replace(namespaceName, "_", "", -1)

	labels[c.config.App.Kubernetes.Namespace.Labels.Team] = strings.ToLower(*formData.Team)

	// set name label
	labels[c.config.App.Kubernetes.Namespace.Labels.Name] = namespaceName

	namespace := models.KubernetesNamespace{Namespace: &coreV1.Namespace{}}
	namespace.Name = namespaceName
	namespace.SetLabels(labels)
	namespace.SettingsApply(nsSettings, c.config.Kubernetes)

	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}

	// NetworkPolicy
	if formData.NetworkPolicy != nil {
		namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.NetworkPolicy] = *formData.NetworkPolicy
	}

	if formData.Description != nil {
		namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.Description] = *formData.Description
	}

	if !c.kubernetesNamespaceAccessAllowed(ctx, namespace) {
		c.respondErrorWithPenalty(ctx, errors.New(fmt.Sprintf("Access to namespace \"%s\" denied", namespace.Name)))
		return
	}

	service := c.serviceKubernetes()

	// check if already exists
	existingNs, _ := service.NamespaceGet(namespace.Name)
	if existingNs != nil && existingNs.GetUID() != "" {
		message := ""
		if existingNsTeam, ok := existingNs.Labels["team"]; ok {
			message = fmt.Sprintf("Namespace \"%s\" already exists (owned by team \"%s\")", namespace.Name, existingNsTeam)
		} else if existingNsUser, ok := existingNs.Labels["user"]; ok {
			message = fmt.Sprintf("Namespace \"%s\" already exists (owned by user \"%s\")", namespace.Name, existingNsUser)
		} else {
			message = fmt.Sprintf("Namespace \"%s\" already exists", namespace.Name)
		}

		c.respondError(ctx, errors.New(message))
		return
	}

	// Namespace creation
	if newNamespace, err := service.NamespaceCreate(*namespace.Namespace); newNamespace != nil && err == nil {
		if err := c.updateNamespaceSettings(ctx, &models.KubernetesNamespace{Namespace: newNamespace}); err != nil {
			c.respondError(ctx, err)
			return
		}
	} else {
		c.respondError(ctx, err)
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "createNamespace"}).Inc()
	c.notificationMessage(ctx, fmt.Sprintf("Namespace \"%s\" created", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("Namespace \"%s\" created", namespace.Name), 1)

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("Namespace \"%s\" created", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) ApiNamespaceDelete(ctx iris.Context, user *models.User) {
	namespaceName := ctx.Params().GetString("namespace")

	if namespaceName == "" {
		c.respondError(ctx, errors.New("Invalid namespace"))
		return
	}

	// get namespace
	namespace, err := c.getNamespace(ctx, namespaceName)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if !c.kubernetesNamespaceDeleteAllowed(ctx, namespace) {
		c.respondError(ctx, errors.New(fmt.Sprintf("Deletion of namespace \"%s\" denied", namespace.Namespace)))
		return
	}

	if err := c.serviceKubernetes().NamespaceDelete(namespace.Name); err != nil {
		c.respondError(ctx, err)
		return
	}

	c.notificationMessage(ctx, fmt.Sprintf("Namespace \"%s\" deleted", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("Namespace \"%s\" deleted", namespace.Name), 1)
	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "deleteNamepace"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("Namespace \"%s\" deleted", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) ApiNamespaceUpdate(ctx iris.Context, user *models.User) {
	namespaceName := ctx.Params().GetString("namespace")

	if namespaceName == "" {
		c.respondError(ctx, errors.New("Invalid namespace"))
		return
	}

	formData := formdata.KubernetesNamespaceCreate{}
	err := ctx.ReadJSON(&formData)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	// get namespace
	namespace, err := c.getNamespace(ctx, namespaceName)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}
	// description
	if formData.Description != nil {
		namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.Description] = *formData.Description
	}
	// NetworkPolicy
	if formData.NetworkPolicy != nil {
		namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.NetworkPolicy] = *formData.NetworkPolicy
	}

	// labels
	if formData.Settings != nil {
		nsSettings, validationMessages := c.validateSettings(*formData.Settings)
		if len(validationMessages) >= 1 {
			c.respondError(ctx, errors.New(strings.Join(validationMessages, "\n")))
			return
		}

		namespace.SettingsApply(nsSettings, c.config.Kubernetes)
	}

	// update networkPolicy
	if err := c.updateNamespaceNetworkPolicy(namespace); err != nil {
		c.respondError(ctx, err)
		return
	}

	// update
	if _, err := c.serviceKubernetes().NamespaceUpdate(namespace.Namespace); err != nil {
		c.respondError(ctx, errors.New(fmt.Sprintf("Update of namespace \"%s\" failed: %v", namespace.Name, err)))
		return
	}

	c.notificationMessage(ctx, fmt.Sprintf("Namespace \"%s\" updated", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("Namespace \"%s\" updated", namespace.Name), 1)
	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "updateNamepace"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("Namespace \"%s\" updated", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) ApiNamespaceReset(ctx iris.Context, user *models.User) {
	namespaceName := ctx.Params().GetString("namespace")

	if namespaceName == "" {
		c.respondError(ctx, errors.New("Invalid namespace"))
		return
	}

	// get namespace
	namespace, err := c.getNamespace(ctx, namespaceName)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if namespace, err = c.updateNamespace(namespace); err != nil {
		c.respondError(ctx, err)
		return
	}

	if err := c.updateNamespaceSettings(ctx, namespace); err != nil {
		c.respondError(ctx, err)
		return
	}

	c.notificationMessage(ctx, fmt.Sprintf("Namespace \"%s\" reset", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("Namespace \"%s\" resetted", namespace.Name), 1)
	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "resetSettings"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("Namespace \"%s\" reset", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) updateNamespace(namespace *models.KubernetesNamespace) (*models.KubernetesNamespace, error) {
	doUpdate := false

	// add env label
	if _, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.Environment]; !ok {
		parts := strings.Split(namespace.Name, "-")

		if len(parts) > 1 {
			namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.Environment] = parts[0]
			doUpdate = true
		}
	}

	if doUpdate {
		if _, err := c.serviceKubernetes().NamespaceUpdate(namespace.Namespace); err != nil {
			return namespace, err
		}
	}

	return namespace, nil
}

func (c *ApplicationKubernetes) kubernetesNamespaceAccessAllowed(ctx iris.Context, namespace models.KubernetesNamespace) bool {
	user := c.getUserOrStop(ctx)

	username := strings.ToLower(user.Username)
	username = strings.Replace(username, "_", "", -1)

	// USER namespace
	regexpUser := regexp.MustCompile(fmt.Sprintf(c.config.App.Kubernetes.Namespace.Filter.User, regexp.QuoteMeta(username)))
	if regexpUser.MatchString(namespace.Name) {
		return true
	}

	if val, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.User]; ok {
		if val == user.Username {
			return true
		}
	}

	// ENV namespace (team labels)
	for _, team := range user.Teams {
		if val, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.Team]; ok {
			if val == team.Name {
				return true
			}
		}
	}

	// TEAM namespace
	teamsQuoted := []string{}
	for _, team := range user.Teams {
		teamsQuoted = append(teamsQuoted, regexp.QuoteMeta(team.Name))
	}

	regexpTeamStr := fmt.Sprintf(c.config.App.Kubernetes.Namespace.Filter.Team, "("+strings.Join(teamsQuoted, "|")+")")
	regexpTeam := regexp.MustCompile(regexpTeamStr)
	if regexpTeam.MatchString(namespace.Name) {
		return true
	}

	return false
}

func (c *ApplicationKubernetes) kubernetesNamespaceDeleteAllowed(ctx iris.Context, namespace *models.KubernetesNamespace) bool {
	ret := regexp.MustCompile(c.config.App.Kubernetes.Namespace.Filter.Delete).MatchString(namespace.Name)

	if val, ok := namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.Immortal]; ok {
		if val == "true" {
			ret = false
		}
	}

	if !c.kubernetesNamespaceCheckOwnership(ctx, namespace) {
		ret = false
	}

	return ret
}

func (c *ApplicationKubernetes) kubernetesNamespaceCheckOwnership(ctx iris.Context, namespace *models.KubernetesNamespace) bool {
	user := c.getUserOrStop(ctx)

	username := user.Username

	if labelUserVal, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.User]; ok {
		if labelUserVal == username {
			return true
		}
	} else if labelTeamVal, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.Team]; ok {
		// Team rolebinding
		if _, err := user.GetTeam(labelTeamVal); err == nil {
			return true
		}
	}

	return false
}

func (c *ApplicationKubernetes) updateNamespaceSettings(ctx iris.Context, namespace *models.KubernetesNamespace) (error error) {
	if err := c.kubernetesNamespacePermissionsUpdate(ctx, namespace); err != nil {
		return err
	}

	if err := c.updateNamespaceObjects(namespace); err != nil {
		return err
	}

	if err := c.updateNamespaceNetworkPolicy(namespace); err != nil {
		return err
	}

	return
}

func (c *ApplicationKubernetes) kubernetesNamespacePermissionsUpdate(ctx iris.Context, namespace *models.KubernetesNamespace) (error error) {
	if !c.kubernetesNamespaceAccessAllowed(ctx, *namespace) {
		return errors.New(fmt.Sprintf("Namespace \"%s\" not owned by current user", namespace.Name))
	}

	user := c.getUserOrStop(ctx)
	username := user.Username
	k8sUsername := user.Id

	if labelUserVal, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.User]; c.config.App.Kubernetes.Namespace.Role.Private && ok {
		if labelUserVal == username {
			// User rolebinding
			role := c.config.App.Kubernetes.Namespace.Role.User
			if _, err := c.serviceKubernetes().RoleBindingCreateNamespaceUser(namespace.Name, username, k8sUsername, role); err != nil {
				return errors.New(fmt.Sprintf("Error: %v", err))
			}
		} else {
			return errors.New(fmt.Sprintf("Namespace \"%s\" not owned by current user", namespace.Name))
		}
	} else if labelTeamVal, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.Team]; ok {
		// Team rolebinding
		if namespaceTeam, err := user.GetTeam(labelTeamVal); err == nil {
			for _, permission := range namespaceTeam.K8sPermissions {
				if _, err := c.serviceKubernetes().RoleBindingCreateNamespaceTeam(namespace.Name, labelTeamVal, permission); err != nil {
					return errors.New(fmt.Sprintf("Error: %v", err))
				}
			}
		}
	} else {
		return errors.New(fmt.Sprintf("Namespace \"%s\" cannot be resetted, labels not found", namespace.Name))
	}

	return
}

func (c *ApplicationKubernetes) updateNamespaceObjects(namespace *models.KubernetesNamespace) (error error) {
	var kubeObjectList *models.KubernetesObjectList

	if environment, ok := namespace.Labels[c.config.App.Kubernetes.Namespace.Labels.Environment]; ok {
		if configObjects, ok := c.config.App.Kubernetes.ObjectsList[environment]; ok {
			kubeObjectList = configObjects
		}
	}

	// if empty, try default
	if kubeObjectList == nil {
		if configObjects, ok := c.config.App.Kubernetes.ObjectsList["_default"]; ok {
			kubeObjectList = configObjects
		}
	}

	if kubeObjectList != nil {
		for _, kubeObject := range kubeObjectList.ConfigMaps {
			error = c.serviceKubernetes().NamespaceEnsureConfigMap(namespace.Name, kubeObject.Name, kubeObject.Object.(*coreV1.ConfigMap))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.ServiceAccounts {
			error = c.serviceKubernetes().NamespaceEnsureServiceAccount(namespace.Name, kubeObject.Name, kubeObject.Object.(*coreV1.ServiceAccount))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.Roles {
			error = c.serviceKubernetes().NamespaceEnsureRole(namespace.Name, kubeObject.Name, kubeObject.Object.(*rbacV1.Role))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.RoleBindings {
			error = c.serviceKubernetes().NamespaceEnsureRoleBindings(namespace.Name, kubeObject.Name, kubeObject.Object.(*rbacV1.RoleBinding))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.NetworkPolicies {
			error = c.serviceKubernetes().NamespaceEnsureNetworkPolicy(namespace.Name, kubeObject.Name, kubeObject.Object.(*networkingV1.NetworkPolicy))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.LimitRanges {
			error = c.serviceKubernetes().NamespaceEnsureLimitRange(namespace.Name, kubeObject.Name, kubeObject.Object.(*coreV1.LimitRange))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.PodPresets {
			error = c.serviceKubernetes().NamespaceEnsurePodPreset(namespace.Name, kubeObject.Name, kubeObject.Object.(*settingsV1alpha1.PodPreset))
			if error != nil {
				return
			}
		}

		for _, kubeObject := range kubeObjectList.ResourceQuotas {
			error = c.serviceKubernetes().NamespaceEnsureResourceQuota(namespace.Name, kubeObject.Name, kubeObject.Object.(*coreV1.ResourceQuota))
			if error != nil {
				return
			}
		}
	}

	return
}

func (c *ApplicationKubernetes) updateNamespaceNetworkPolicy(namespace *models.KubernetesNamespace) error {
	ctx := context.Background()
	var err error

	if namespace.Annotations == nil {
		return nil
	}

	if val, ok := namespace.Annotations[c.config.App.Kubernetes.Namespace.Annotations.NetworkPolicy]; ok {
		// delete default netpol
		if kubeObject, _ := c.serviceKubernetes().Client().NetworkingV1().NetworkPolicies(namespace.Name).Get(ctx, "default", metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
			err = c.serviceKubernetes().Client().NetworkingV1().NetworkPolicies(namespace.Name).Delete(ctx, "default", metav1.DeleteOptions{})
			if err != nil {
				c.logger.Info(fmt.Sprintf("Deletion of NetworkPolicy/default in namespace %v failed: %v", namespace.Name, err))
			}
		}

		// create netpol
		for _, netpol := range c.config.App.Kubernetes.Namespace.NetworkPolicy {
			if netpol.Name == val {
				k8sObject := netpol.GetKubernetesObject()
				if k8sObject != nil {
					_, err = c.serviceKubernetes().Client().NetworkingV1().NetworkPolicies(namespace.Name).Create(ctx, k8sObject ,metav1.CreateOptions{})
					if err != nil {
						c.logger.Error(fmt.Sprintf("Creation of NetworkPolicy in namespace %v failed: %v", namespace.Name, err))
					}
				}
				break
			}
		}
	}
	return nil
}

func (c *ApplicationKubernetes) checkNamespaceTeamQuota(team string) (err error) {
	var count int
	quota := c.config.App.Kubernetes.Namespace.Quota.Team

	if quota <= 0 {
		// no quota
		return
	}

	regexp := regexp.MustCompile(fmt.Sprintf(c.config.App.Kubernetes.Namespace.Filter.Team, regexp.QuoteMeta(team)))

	count, err = c.serviceKubernetes().NamespaceCount(regexp)
	if err != nil {
		return
	}

	if count >= quota {
		// quota exceeded
		err = errors.New(fmt.Sprintf("Team namespace quota of %v namespaces exceeded ", quota))
	}

	return
}

func (c *ApplicationKubernetes) checkNamespaceUserQuota(username string) (err error) {
	var count int
	quota := c.config.App.Kubernetes.Namespace.Quota.User

	if quota <= 0 {
		// no quota
		return
	}

	regexp := regexp.MustCompile(fmt.Sprintf(c.config.App.Kubernetes.Namespace.Filter.User, regexp.QuoteMeta(username)))

	count, err = c.serviceKubernetes().NamespaceCount(regexp)
	if err != nil {
		return
	}

	if count >= quota {
		// quota exceeded
		err = errors.New(fmt.Sprintf("Personal namespace quota of %v namespaces exceeded ", quota))
	}

	return
}

func (c *ApplicationKubernetes) validateSettings(formSettingList map[string]string) (ret map[string]string, validationMsgs []string) {
	validationMsgs = []string{}

	for _, setting := range c.config.Kubernetes.Namespace.Settings {
		settingValue := ""

		if val, ok := formSettingList[setting.Name]; ok {
			settingValue = val
		}

		if !setting.Validation.Validate(settingValue) {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
		}

		if val := setting.Transformation.Transform(settingValue); val != nil {
			formSettingList[setting.Name] = *val
		} else {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Parsing of \"%s\" failed", setting.Label))
		}
	}

	ret = formSettingList

	return
}

func (c *ApplicationKubernetes) getNamespace(ctx iris.Context, namespaceName string) (namespace *models.KubernetesNamespace, err error) {
	if namespaceName == "" {
		return nil, errors.New("Invalid namespace")
	}

	namespaceNative, err := c.serviceKubernetes().NamespaceGet(namespaceName)

	if err != nil {
		return nil, err
	}

	namespace = &models.KubernetesNamespace{Namespace: namespaceNative}

	if !c.kubernetesNamespaceAccessAllowed(ctx, *namespace) {
		return nil, errors.New(fmt.Sprintf("Access to namespace \"%s\" denied", namespace.Name))
	}

	return
}

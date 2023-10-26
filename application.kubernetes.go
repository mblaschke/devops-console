package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	iris "github.com/kataras/iris/v12"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/utils/to"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mblaschke/devops-console/models"
	"github.com/mblaschke/devops-console/models/formdata"
	"github.com/mblaschke/devops-console/models/response"
	"github.com/mblaschke/devops-console/services"
)

const (
	RedisKubernetesNamespaceList      = `kubernetes:namespace:list`
	KubernetesNamespaceLabelManagedBy = `devops-console`
)

type ApplicationKubernetes struct {
	*Server
}

func NewApplicationKubernetes(c *Server) *ApplicationKubernetes {
	app := ApplicationKubernetes{Server: c}
	return &app
}

func (c *ApplicationKubernetes) serviceKubernetes() (service *services.Kubernetes) {
	service = &services.Kubernetes{
		Config: c.config,
	}
	service.Filter.Namespace = c.config.Kubernetes.Namespace.Filter.Access

	return
}

func (c *ApplicationKubernetes) ApiNamespaceList(ctx iris.Context, user *models.User) {
	var (
		err    error
		nsList map[string]coreV1.Namespace
	)

	if ok := c.redisCacheLoad(RedisKubernetesNamespaceList, &nsList); !ok {
		nsList, err = c.serviceKubernetes().NamespaceList()
		if err != nil {
			c.respondError(ctx, fmt.Errorf("unable to contact Kubernetes cluster"))
			return
		}

		if err := c.redisCacheSet(RedisKubernetesNamespaceList, nsList, 30*time.Second); err != nil {
			c.logger.Errorf(`unable to save Kubernetes namespaces to cache: %v`, err)
		}
	}

	ret := []response.KubernetesNamespace{}

	for _, row := range nsList {
		nativeNamespace := row
		namespace := models.KubernetesNamespace{Namespace: &nativeNamespace}

		if !c.kubernetesNamespaceAccessAllowed(ctx, namespace, user) {
			continue
		}

		managedBy := ""
		if val, ok := namespace.Annotations[c.config.Kubernetes.Namespace.Labels.ManagedBy]; ok {
			// only show if managed by different source
			if !strings.EqualFold(val, KubernetesNamespaceLabelManagedBy) {
				managedBy = val
			}
		}

		row := response.KubernetesNamespace{
			Name:       namespace.Name,
			Status:     fmt.Sprintf("%v", namespace.Status.Phase),
			Created:    namespace.CreationTimestamp.UTC().String(),
			CreatedAgo: humanize.Time(namespace.CreationTimestamp.UTC()),
			ManagedBy:  managedBy,
			Editable:   c.kubernetesNamespaceEditAllowed(ctx, &namespace, user),
			Deleteable: c.kubernetesNamespaceDeleteAllowed(ctx, &namespace, user),
			Settings:   namespace.SettingsExtract(c.config.Kubernetes),
		}

		if opts.Kubernetes.EnableNamespacePodCount {
			row.PodCount = c.serviceKubernetes().NamespacePodCount(namespace.Name)
		}

		if val, ok := namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.Description]; ok {
			row.Description = val
		}

		if val, ok := namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.NetworkPolicy]; ok {
			row.NetworkPolicy = val
		}

		if val, ok := namespace.Labels[c.config.Kubernetes.Namespace.Labels.Team]; ok {
			row.OwnerTeam = val
		}

		ret = append(ret, row)
	}

	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "listNamespace"}).Inc()

	c.responseJson(ctx, ret)
}

func (c *ApplicationKubernetes) ApiNamespaceCreate(ctx iris.Context, user *models.User) {
	var namespaceName string

	formData, err := c.getJsonFromFormData(ctx)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	if formData.Settings == nil || formData.Description == nil || formData.Name == nil || formData.Team == nil {
		c.respondError(ctx, fmt.Errorf("invalid form data"))
		return
	}

	// team filter check
	if !c.config.Kubernetes.Namespace.Validation.Team.Validate(*formData.Team) {
		c.respondError(ctx, fmt.Errorf("invalid team value (%v)", c.config.Kubernetes.Namespace.Validation.Team.String()))
		return
	}

	labels := map[string]string{
		c.config.Kubernetes.Namespace.Labels.Team: *formData.Team,
	}

	// validation
	nsSettings, validationMessages := c.validateSettings(*formData.Settings)
	if len(validationMessages) >= 1 {
		c.respondError(ctx, fmt.Errorf(strings.Join(validationMessages, "\n")))
		return
	}

	// membership check
	if !user.IsMemberOf(*formData.Team) {
		c.respondError(ctx, fmt.Errorf("access to team \"%s\" denied", *formData.Team))
		return
	}

	// quota check
	if err := c.checkNamespaceTeamQuota(*formData.Team); err != nil {
		c.respondError(ctx, err)
		return
	}

	// build namespace name
	namespaceName = *formData.Name

	// namespace filtering
	namespaceName = strings.ToLower(namespaceName)
	namespaceName = strings.Replace(namespaceName, "_", "", -1)

	labels[c.config.Kubernetes.Namespace.Labels.Team] = strings.ToLower(*formData.Team)

	namespace := models.KubernetesNamespace{Namespace: &coreV1.Namespace{}}
	namespace.Name = namespaceName
	namespace.SetLabels(labels)
	namespace.SettingsApply(nsSettings, c.config.Kubernetes)

	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}

	// NetworkPolicy
	if formData.NetworkPolicy != nil {
		namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.NetworkPolicy] = *formData.NetworkPolicy
	}

	if formData.Description != nil {
		namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.Description] = *formData.Description
	}
	namespace.Labels[c.config.Kubernetes.Namespace.Labels.ManagedBy] = KubernetesNamespaceLabelManagedBy

	service := c.serviceKubernetes()

	// check if already exists
	existingNs, _ := service.NamespaceGet(namespace.Name)
	if existingNs != nil && existingNs.GetUID() != "" {
		message := ""
		if existingNsTeam, ok := existingNs.Labels[c.config.Kubernetes.Namespace.Labels.Team]; ok {
			message = fmt.Sprintf("namespace \"%s\" already exists (owned by team \"%s\")", namespace.Name, existingNsTeam)
		} else {
			message = fmt.Sprintf("namespace \"%s\" already exists", namespace.Name)
		}

		c.respondError(ctx, fmt.Errorf(message))
		return
	}

	// Namespace creation
	if newNamespace, err := service.NamespaceCreate(*namespace.Namespace); newNamespace != nil && err == nil {
		if err := c.updateNamespaceSettings(ctx, &models.KubernetesNamespace{Namespace: newNamespace}, user); err != nil {
			c.respondError(ctx, err)
			return
		}
	} else {
		c.respondError(ctx, err)
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "createNamespace"}).Inc()
	c.notificationMessage(ctx, fmt.Sprintf("namespace \"%s\" created", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("namespace \"%s\" created", namespace.Name), 1)

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("namespace \"%s\" created", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) ApiNamespaceDelete(ctx iris.Context, user *models.User) {
	namespaceName := ctx.Params().GetString("namespace")

	if namespaceName == "" {
		c.respondError(ctx, fmt.Errorf("invalid namespace"))
		return
	}

	// get namespace
	namespace, err := c.getNamespace(ctx, namespaceName, user)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if !c.kubernetesNamespaceDeleteAllowed(ctx, namespace, user) {
		c.respondError(ctx, fmt.Errorf("deletion of namespace \"%s\" denied", namespace.Namespace))
		return
	}

	if err := c.serviceKubernetes().NamespaceDelete(namespace.Name); err != nil {
		c.respondError(ctx, err)
		return
	}
	c.redisCacheInvalidate(RedisKubernetesNamespaceList)

	c.notificationMessage(ctx, fmt.Sprintf("namespace \"%s\" deleted", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("namespace \"%s\" deleted", namespace.Name), 1)
	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "deleteNamepace"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("namespace \"%s\" deleted", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) ApiNamespaceUpdate(ctx iris.Context, user *models.User) {
	namespaceName := ctx.Params().GetString("namespace")

	if namespaceName == "" {
		c.respondError(ctx, fmt.Errorf("invalid namespace"))
		return
	}

	formData, err := c.getJsonFromFormData(ctx)
	if err != nil {
		c.respondErrorWithPenalty(ctx, err)
		return
	}

	// get namespace
	namespace, err := c.getNamespace(ctx, namespaceName, user)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if !c.kubernetesNamespaceEditAllowed(ctx, namespace, user) {
		message := fmt.Sprintf("namespace \"%s\" exists and not managed by devops-console", namespace.Name)
		c.respondError(ctx, fmt.Errorf(message))
		return
	}

	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}
	// description
	if formData.Description != nil {
		namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.Description] = *formData.Description
	}
	// NetworkPolicy
	if formData.NetworkPolicy != nil {
		namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.NetworkPolicy] = *formData.NetworkPolicy
	}
	namespace.Labels[c.config.Kubernetes.Namespace.Labels.ManagedBy] = KubernetesNamespaceLabelManagedBy

	// labels
	if formData.Settings != nil {
		nsSettings, validationMessages := c.validateSettings(*formData.Settings)
		if len(validationMessages) >= 1 {
			c.respondError(ctx, fmt.Errorf(strings.Join(validationMessages, "\n")))
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
		c.respondError(ctx, fmt.Errorf("update of namespace \"%s\" failed: %w", namespace.Name, err))
		return
	}

	c.redisCacheInvalidate(RedisKubernetesNamespaceList)

	c.notificationMessage(ctx, fmt.Sprintf("namespace \"%s\" updated", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("namespace \"%s\" updated", namespace.Name), 1)
	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "updateNamepace"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("namespace \"%s\" updated", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) ApiNamespaceReset(ctx iris.Context, user *models.User) {
	namespaceName := ctx.Params().GetString("namespace")

	if namespaceName == "" {
		c.respondError(ctx, fmt.Errorf("invalid namespace"))
		return
	}

	// get namespace
	namespace, err := c.getNamespace(ctx, namespaceName, user)
	if err != nil {
		c.respondError(ctx, err)
		return
	}

	if !c.kubernetesNamespaceEditAllowed(ctx, namespace, user) {
		message := fmt.Sprintf("namespace \"%s\" exists and not managed by devops-console", namespace.Name)
		c.respondError(ctx, fmt.Errorf(message))
		return
	}

	if namespace, err = c.updateNamespace(namespace, false); err != nil {
		c.respondError(ctx, err)
		return
	}

	if err := c.updateNamespaceSettings(ctx, namespace, user); err != nil {
		c.respondError(ctx, err)
		return
	}

	c.notificationMessage(ctx, fmt.Sprintf("namespace \"%s\" reset", namespace.Name))
	c.auditLog(ctx, fmt.Sprintf("namespace \"%s\" resetted", namespace.Name), 1)
	PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "resetSettings"}).Inc()

	resp := response.GeneralMessage{
		Message: fmt.Sprintf("namespace \"%s\" reset", namespace.Name),
	}

	c.responseJson(ctx, resp)
}

func (c *ApplicationKubernetes) updateNamespace(namespace *models.KubernetesNamespace, force bool) (*models.KubernetesNamespace, error) {
	if force {
		if namespace.Labels == nil {
			namespace.Labels = map[string]string{}
		}
		namespace.Labels[c.config.Kubernetes.Namespace.Labels.ManagedBy] = KubernetesNamespaceLabelManagedBy

		if _, err := c.serviceKubernetes().NamespaceUpdate(namespace.Namespace); err != nil {
			return namespace, err
		}
		c.redisCacheInvalidate(RedisKubernetesNamespaceList)
	}

	return namespace, nil
}

func (c *ApplicationKubernetes) kubernetesNamespaceAccessAllowed(ctx iris.Context, namespace models.KubernetesNamespace, user *models.User) bool {
	if user == nil {
		return false
	}

	// check team
	if labelVal, ok := namespace.Labels[c.config.Kubernetes.Namespace.Labels.Team]; ok {
		if team, err := user.GetTeam(labelVal); team != nil && err == nil {
			return true
		}
	}

	return false
}

func (c *ApplicationKubernetes) kubernetesNamespaceEditAllowed(ctx iris.Context, namespace *models.KubernetesNamespace, user *models.User) bool {
	ret := false

	if val, ok := namespace.Labels[c.config.Kubernetes.Namespace.Labels.ManagedBy]; ok {
		if val == "" || strings.EqualFold(val, KubernetesNamespaceLabelManagedBy) {
			ret = true
		}
	} else {
		// no annotation set, it's ok
		ret = true
	}

	if !c.kubernetesNamespaceCheckOwnership(ctx, namespace, user) {
		ret = false
	}

	return ret
}
func (c *ApplicationKubernetes) kubernetesNamespaceDeleteAllowed(ctx iris.Context, namespace *models.KubernetesNamespace, user *models.User) bool {
	ret := c.config.Kubernetes.Namespace.Filter.Delete.Validate(namespace.Name)

	if val, ok := namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.Immortal]; ok {
		if val == "true" {
			ret = false
		}
	}

	if !c.kubernetesNamespaceCheckOwnership(ctx, namespace, user) {
		ret = false
	}

	return ret
}

func (c *ApplicationKubernetes) kubernetesNamespaceCheckOwnership(ctx iris.Context, namespace *models.KubernetesNamespace, user *models.User) bool {
	if user == nil {
		return false
	}

	if labelTeamVal, ok := namespace.Labels[c.config.Kubernetes.Namespace.Labels.Team]; ok {
		// Team rolebinding
		if _, err := user.GetTeam(labelTeamVal); err == nil {
			return true
		}
	}

	return false
}

func (c *ApplicationKubernetes) updateNamespaceSettings(ctx iris.Context, namespace *models.KubernetesNamespace, user *models.User) (error error) {

	if err := c.kubernetesNamespacePermissionsUpdate(ctx, namespace, user); err != nil {
		return err
	}

	if err := c.updateNamespaceObjects(namespace); err != nil {
		return err
	}

	if err := c.updateNamespaceNetworkPolicy(namespace); err != nil {
		return err
	}

	c.redisCacheInvalidate(RedisKubernetesNamespaceList)

	return
}

func (c *ApplicationKubernetes) kubernetesNamespacePermissionsUpdate(ctx iris.Context, namespace *models.KubernetesNamespace, user *models.User) (error error) {
	if !c.kubernetesNamespaceAccessAllowed(ctx, *namespace, user) {
		return fmt.Errorf("namespace \"%s\" not owned by current user", namespace.Name)
	}

	if labelTeamVal, ok := namespace.Labels[c.config.Kubernetes.Namespace.Labels.Team]; ok {
		// Team rolebinding
		if namespaceTeam, err := user.GetTeam(labelTeamVal); err == nil {
			if _, err := c.serviceKubernetes().RoleBindingCreateNamespaceTeam(namespace.Name, namespaceTeam); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("namespace \"%s\" cannot be resetted, labels not found", namespace.Name)
	}

	return
}

func (c *ApplicationKubernetes) updateNamespaceObjects(namespace *models.KubernetesNamespace) (error error) {
	var kubeObjectList *models.KubernetesObjectList

	// if empty, try default
	if kubeObjectList == nil {
		if configObjects, ok := c.config.Kubernetes.ObjectsList["_default"]; ok {
			kubeObjectList = &configObjects
		}
	}

	if kubeObjectList != nil {
		for _, row := range *kubeObjectList {
			kubeObject := row
			error = c.serviceKubernetes().EnsureResourceInNamespace(namespace.Name, &kubeObject.Object)
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

	if val, ok := namespace.Annotations[c.config.Kubernetes.Namespace.Annotations.NetworkPolicy]; ok {
		// delete default netpol
		if kubeObject, _ := c.serviceKubernetes().Client().NetworkingV1().NetworkPolicies(namespace.Name).Get(ctx, "default", metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
			err = c.serviceKubernetes().Client().NetworkingV1().NetworkPolicies(namespace.Name).Delete(ctx, "default", metav1.DeleteOptions{})
			if err != nil {
				c.logger.Info(fmt.Sprintf("deletion of NetworkPolicy/default in namespace %v failed: %v", namespace.Name, err))
			}
		}

		// create netpol
		for _, netpol := range c.config.Kubernetes.Namespace.NetworkPolicy {
			if netpol.Name == val {
				k8sObject := netpol.GetKubernetesObject()
				resource := k8sObject.DeepCopyObject()
				err = c.serviceKubernetes().EnsureResourceInNamespace(namespace.Name, &resource)
				if err != nil {
					c.logger.Error(fmt.Sprintf("creation of NetworkPolicy in namespace %v failed: %v", namespace.Name, err))
				}
				break
			}
		}
	}
	return nil
}

func (c *ApplicationKubernetes) checkNamespaceTeamQuota(team string) (err error) {
	var count int
	quota := c.config.Kubernetes.Namespace.Quota.Team

	if quota <= 0 {
		// no quota
		return
	}

	namespaceList, err := c.serviceKubernetes().NamespaceList()
	if err != nil {
		return
	}

	for _, namespace := range namespaceList {
		if labelTeamVal, ok := namespace.Labels[c.config.Kubernetes.Namespace.Labels.Team]; ok {
			if strings.EqualFold(labelTeamVal, team) {
				count++
			}
		}
	}

	if count >= quota {
		// quota exceeded
		err = fmt.Errorf("team namespace quota of %v namespaces exceeded ", quota)
	}

	return
}

func (c *ApplicationKubernetes) validateSettings(formSettingList map[string]string) (ret map[string]string, validationMsgs []string) {
	validationMsgs = []string{}

	for _, setting := range c.config.Kubernetes.Namespace.Settings {
		settingValue := ""

		if val, ok := formSettingList[setting.Name]; ok {
			settingValue = val
		} else {
			settingValue = setting.Default
		}

		if !setting.Validation.Validate(settingValue) {
			validationMsgs = append(validationMsgs, fmt.Sprintf("validation of \"%s\" failed (%v)", setting.Label, setting.Validation.HumanizeString()))
		}

		if val := setting.Transformation.Transform(settingValue); val != nil {
			formSettingList[setting.Name] = *val
		} else {
			validationMsgs = append(validationMsgs, fmt.Sprintf("parsing of \"%s\" failed", setting.Label))
		}
	}

	ret = formSettingList

	return
}

func (c *ApplicationKubernetes) getNamespace(ctx iris.Context, namespaceName string, user *models.User) (namespace *models.KubernetesNamespace, err error) {
	if namespaceName == "" {
		return nil, fmt.Errorf("invalid namespace")
	}

	namespaceNative, err := c.serviceKubernetes().NamespaceGet(namespaceName)

	if err != nil {
		return nil, err
	}

	namespace = &models.KubernetesNamespace{Namespace: namespaceNative}

	if !c.kubernetesNamespaceAccessAllowed(ctx, *namespace, user) {
		return nil, fmt.Errorf("access to namespace \"%s\" denied", namespace.Name)
	}

	return
}

func (c *ApplicationKubernetes) getJsonFromFormData(ctx iris.Context) (formData *formdata.KubernetesNamespaceCreate, err error) {
	formData = &formdata.KubernetesNamespaceCreate{}
	err = ctx.ReadJSON(&formData)
	if err != nil {
		return
	}

	// inject defaults
	if formData.NetworkPolicy == nil {
		for _, netpol := range c.config.Kubernetes.Namespace.NetworkPolicy {
			if netpol.Default {
				formData.NetworkPolicy = to.StringPtr(netpol.Name)
			}
		}
	}

	return
}

package services

import (
	"k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	v13 "k8s.io/api/rbac/v1"
	"k8s.io/api/settings/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccount
func (k *Kubernetes) NamespaceEnsureServiceAccount(namespace, name string, object *v1.ServiceAccount) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().CoreV1().ServiceAccounts(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().ServiceAccounts(namespace).Update(object)
	} else {
		_, error = k.Client().CoreV1().ServiceAccounts(namespace).Create(object)
	}

	return
}

// ResourceQuotas
func (k *Kubernetes) NamespaceEnsureResourceQuota(namespace, name string, object *v1.ResourceQuota) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().CoreV1().ResourceQuotas(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().ResourceQuotas(namespace).Update(object)
	} else {
		_, error = k.Client().CoreV1().ResourceQuotas(namespace).Create(object)
	}

	return
}

// LimitRange
func (k *Kubernetes) NamespaceEnsureLimitRange(namespace, name string, object *v1.LimitRange) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().CoreV1().LimitRanges(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().LimitRanges(namespace).Update(object)
	} else {
		_, error = k.Client().CoreV1().LimitRanges(namespace).Create(object)
	}

	return
}

// PodPresets
func (k *Kubernetes) NamespaceEnsurePodPreset(namespace, name string, object *v1alpha1.PodPreset) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().SettingsV1alpha1().PodPresets(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().SettingsV1alpha1().PodPresets(namespace).Update(object)
	} else {
		_, error = k.Client().SettingsV1alpha1().PodPresets(namespace).Create(object)
	}

	return
}

// NetworkPolicies
func (k *Kubernetes) NamespaceEnsureNetworkPolicy(namespace, name string, object *v12.NetworkPolicy) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().NetworkingV1().NetworkPolicies(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().NetworkingV1().NetworkPolicies(namespace).Update(object)
	} else {
		_, error = k.Client().NetworkingV1().NetworkPolicies(namespace).Create(object)
	}

	return
}

// Roles
func (k *Kubernetes) NamespaceEnsureRole(namespace, name string, object *v13.Role) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().RbacV1().Roles(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().RbacV1().Roles(namespace).Update(object)
	} else {
		_, error = k.Client().RbacV1().Roles(namespace).Create(object)
	}

	return
}

// RoleBinding
func (k *Kubernetes) NamespaceEnsureRoleBindings(namespace, name string, object *v13.RoleBinding) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().RbacV1().RoleBindings(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().RbacV1().RoleBindings(namespace).Update(object)
	} else {
		_, error = k.Client().RbacV1().RoleBindings(namespace).Create(object)
	}

	return
}

// ConfigMap
func (k *Kubernetes) NamespaceEnsureConfigMap(namespace, name string, object *v1.ConfigMap) (error error) {
	exists := false

	getOpts := metav1.GetOptions{}
	if kubeObject, _ := k.Client().CoreV1().ConfigMaps(namespace).Get(name, getOpts); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().ConfigMaps(namespace).Update(object)
	} else {
		_, error = k.Client().CoreV1().ConfigMaps(namespace).Create(object)
	}

	return
}

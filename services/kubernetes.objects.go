package services

import (
	"context"
	"k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	v13 "k8s.io/api/rbac/v1"
	"k8s.io/api/settings/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccount
func (k *Kubernetes) NamespaceEnsureServiceAccount(namespace, name string, object *v1.ServiceAccount) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().ServiceAccounts(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().CoreV1().ServiceAccounts(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

// ResourceQuotas
func (k *Kubernetes) NamespaceEnsureResourceQuota(namespace, name string, object *v1.ResourceQuota) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().ResourceQuotas(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().CoreV1().ResourceQuotas(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

// LimitRange
func (k *Kubernetes) NamespaceEnsureLimitRange(namespace, name string, object *v1.LimitRange) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().LimitRanges(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().CoreV1().LimitRanges(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

// PodPresets
func (k *Kubernetes) NamespaceEnsurePodPreset(namespace, name string, object *v1alpha1.PodPreset) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().SettingsV1alpha1().PodPresets(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().SettingsV1alpha1().PodPresets(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().SettingsV1alpha1().PodPresets(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

// NetworkPolicies
func (k *Kubernetes) NamespaceEnsureNetworkPolicy(namespace, name string, object *v12.NetworkPolicy) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().NetworkingV1().NetworkPolicies(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().NetworkingV1().NetworkPolicies(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

// Roles
func (k *Kubernetes) NamespaceEnsureRole(namespace, name string, object *v13.Role) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().RbacV1().Roles(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().RbacV1().Roles(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

// RoleBinding
func (k *Kubernetes) NamespaceEnsureRoleBindings(namespace, name string, object *v13.RoleBinding) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().RbacV1().RoleBindings(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().RbacV1().RoleBindings(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

// ConfigMap
func (k *Kubernetes) NamespaceEnsureConfigMap(namespace, name string, object *v1.ConfigMap) (error error) {
	ctx := context.Background()
	exists := false

	if kubeObject, _ := k.Client().CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{}); kubeObject != nil && kubeObject.GetUID() != "" {
		exists = true
		object.DeepCopyInto(kubeObject)
		object = kubeObject
	}

	if exists {
		_, error = k.Client().CoreV1().ConfigMaps(namespace).Update(ctx, object, metav1.UpdateOptions{})
	} else {
		_, error = k.Client().CoreV1().ConfigMaps(namespace).Create(ctx, object, metav1.CreateOptions{})
	}

	return
}

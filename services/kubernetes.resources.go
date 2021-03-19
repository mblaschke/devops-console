package services

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (k *Kubernetes) EnsureResourceInNamespace(namespace string, resource *runtime.Object) (err error) {
	ctx := context.Background()
	exists := false

	// conver to unstructured
	unstructuredMap, convertErr := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if convertErr != nil {
		return convertErr
	}
	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

	name := unstructuredObj.GetName()

	// translate to gvr
	gvr, gvrErr := k.findGVR(unstructuredObj.GroupVersionKind())
	if gvrErr != nil {
		return gvrErr
	}

	// use dynamic client to test if resource is there (update and use existing object)
	if kubeObject, kubeError := k.DynClient().Resource(gvr.Resource).Namespace(namespace).Get(ctx, name, metav1.GetOptions{}); kubeError == nil && kubeObject != nil {
		exists = true
		unstructuredObj.DeepCopyInto(kubeObject)
		unstructuredObj = kubeObject
	}

	if exists {
		_, err = k.DynClient().Resource(gvr.Resource).Namespace(namespace).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
	} else {
		_, err = k.DynClient().Resource(gvr.Resource).Namespace(namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	}

	return
}

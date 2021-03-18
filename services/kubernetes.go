package services

import (
	"context"
	"devops-console/models"
	"fmt"
	"k8s.io/api/core/v1"
	v12 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"regexp"
	"strings"
)

type Kubernetes struct {
	clientset *kubernetes.Clientset

	Filter struct {
		Namespace *regexp.Regexp
	}
}

// Create cached kubernetes client
func (k *Kubernetes) Client() (clientset *kubernetes.Clientset) {
	var err error
	var config *rest.Config

	if k.clientset == nil {
		if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
			// KUBECONFIG
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				panic(err)
			}
		} else {
			// K8S in cluster
			config, err = rest.InClusterConfig()
			if err != nil {
				panic(err)
			}
		}

		k.clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
	}

	return k.clientset
}

// Returns list of (filtered) namespaces
func (k *Kubernetes) NamespaceList() (nsList map[string]v1.Namespace, err error) {
	ctx := context.Background()
	result, kubeErr := k.Client().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	if kubeErr == nil {
		nsList = make(map[string]v1.Namespace, len(result.Items))
		for _, ns := range result.Items {
			if err := k.namespaceValidate(ns.Name); err == nil {
				nsList[ns.Name] = ns
			}
		}
	} else {
		err = kubeErr
	}

	return
}

// Returns count of namespaces
func (k *Kubernetes) NamespaceCount(regexp *regexp.Regexp) (count int, err error) {
	var nsList map[string]v1.Namespace
	nsList, err = k.NamespaceList()

	if err == nil {
		if regexp != nil {
			nsListTemp := make(map[string]v1.Namespace, len(nsList))
			for key, val := range nsList {
				if regexp.MatchString(key) {
					nsListTemp[key] = val
				}
			}
			nsList = nsListTemp
		}

		count = len(nsList)
	}

	return
}

// Returns list of nodes
func (k *Kubernetes) Nodes() (*v1.NodeList, error) {
	ctx := context.Background()
	return k.Client().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}

// Returns one namespace
func (k *Kubernetes) NamespaceGet(namespace string) (*v1.Namespace, error) {
	ctx := context.Background()
	if err := k.namespaceValidate(namespace); err != nil {
		return nil, err
	}

	return k.Client().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
}

// Create one namespace
func (k *Kubernetes) NamespaceCreate(namespace v1.Namespace) (*v1.Namespace, error) {
	ctx := context.Background()
	if err := k.namespaceValidate(namespace.Name); err != nil {
		return nil, err
	}

	namespaceNew, err := k.Client().CoreV1().Namespaces().Create(ctx, &namespace, metav1.CreateOptions{})

	return namespaceNew, err
}

// Updates namespace
func (k *Kubernetes) NamespaceUpdate(namespace *v1.Namespace) (*v1.Namespace, error) {
	ctx := context.Background()
	if err := k.namespaceValidate(namespace.Name); err != nil {
		return nil, err
	}

	namespaceNew, err := k.Client().CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})

	return namespaceNew, err
}

// Delete one namespace
func (k *Kubernetes) NamespaceDelete(namespace string) error {
	ctx := context.Background()

	if err := k.namespaceValidate(namespace); err != nil {
		return err
	}

	return k.Client().CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
}

func (k *Kubernetes) NamespacePodCount(namespace string) (podcount *int64) {
	ctx := context.Background()
	if err := k.namespaceValidate(namespace); err != nil {
		return
	}

	count := int64(0)
	result, err := k.Client().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		count += int64(len(result.Items))
		if result.RemainingItemCount != nil {
			count += *result.RemainingItemCount
		}
	}

	return &count
}

// Create cluster rolebinding for user for general access
func (k *Kubernetes) ClusterRoleBindingUser(username, userid, roleName string) (roleBinding *v12.ClusterRoleBinding, error error) {
	ctx := context.Background()
	roleBindName := fmt.Sprintf("user:%s", username)

	if rb, _ := k.Client().RbacV1().ClusterRoleBindings().Get(ctx, roleBindName, metav1.GetOptions{}); rb != nil && rb.GetUID() != "" {
		err := k.Client().RbacV1().ClusterRoleBindings().Delete(ctx, roleBindName, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	}

	annotiations := map[string]string{}
	annotiations["user"] = strings.ToLower(username)

	subject := v12.Subject{}
	subject.Name = userid
	subject.Kind = "User"

	role := v12.RoleRef{}
	role.Kind = "ClusterRole"
	role.Name = roleName

	roleBinding = &v12.ClusterRoleBinding{}
	roleBinding.SetAnnotations(annotiations)
	roleBinding.SetName(roleBindName)
	roleBinding.RoleRef = role
	roleBinding.Subjects = []v12.Subject{subject}

	return k.Client().RbacV1().ClusterRoleBindings().Create(ctx, roleBinding, metav1.CreateOptions{})
}

// Create rolebinding for user to gain access to namespace
func (k *Kubernetes) RoleBindingCreateNamespaceUser(namespace, username, userid, roleName string) (roleBinding *v12.RoleBinding, error error) {
	ctx := context.Background()
	roleBindName := fmt.Sprintf("user:%s", username)

	if rb, _ := k.Client().RbacV1().RoleBindings(namespace).Get(ctx, roleBindName, metav1.GetOptions{}); rb != nil && rb.GetUID() != "" {
		err := k.Client().RbacV1().RoleBindings(namespace).Delete(ctx, roleBindName, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	}

	annotiations := map[string]string{}
	annotiations["user"] = strings.ToLower(username)

	subject := v12.Subject{}
	subject.Name = userid
	subject.Kind = "User"

	role := v12.RoleRef{}
	role.Kind = "ClusterRole"
	role.Name = roleName

	roleBinding = &v12.RoleBinding{}
	roleBinding.SetAnnotations(annotiations)
	roleBinding.SetName(roleBindName)
	roleBinding.SetNamespace(namespace)
	roleBinding.RoleRef = role
	roleBinding.Subjects = []v12.Subject{subject}

	return k.Client().RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
}

// Create rolebinding for group to gain access to namespace
func (k *Kubernetes) RoleBindingCreateNamespaceTeam(namespace string, teamName string, permission models.TeamK8sPermissions) (roleBinding *v12.RoleBinding, error error) {
	ctx := context.Background()
	roleBindName := fmt.Sprintf("team:%s:%s", teamName, permission.Name)

	if rb, _ := k.Client().RbacV1().RoleBindings(namespace).Get(ctx, roleBindName, metav1.GetOptions{}); rb != nil && rb.GetUID() != "" {
		err := k.Client().RbacV1().RoleBindings(namespace).Delete(ctx, roleBindName, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	}

	annotiations := map[string]string{}
	annotiations["team"] = strings.ToLower(teamName)

	subjectList := []v12.Subject{}
	for _, group := range permission.Groups {
		subjectList = append(subjectList, v12.Subject{Kind: "Group", Name: group})
	}

	for _, user := range permission.Users {
		subjectList = append(subjectList, v12.Subject{Kind: "User", Name: user})
	}

	for _, serviceAccount := range permission.ServiceAccounts {
		if serviceAccount.Namespace == "" {
			// default is local namespace
			serviceAccount.Namespace = namespace
		}

		subjectList = append(subjectList, v12.Subject{Kind: "ServiceAccount", Name: serviceAccount.Name, Namespace: serviceAccount.Namespace})
	}

	role := v12.RoleRef{}
	role.Kind = "ClusterRole"
	role.Name = permission.ClusterRole

	roleBinding = &v12.RoleBinding{}
	roleBinding.SetAnnotations(annotiations)
	roleBinding.SetName(roleBindName)
	roleBinding.SetNamespace(namespace)
	roleBinding.RoleRef = role
	roleBinding.Subjects = subjectList

	return k.Client().RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
}

// Create rolebinding for group to gain access to namespace
func (k *Kubernetes) RoleBindingCreateNamespaceGroup(namespace, group, roleName string) (roleBinding *v12.RoleBinding, error error) {
	ctx := context.Background()
	roleBindName := fmt.Sprintf("group:%s", group)

	if rb, _ := k.Client().RbacV1().RoleBindings(namespace).Get(ctx, roleBindName, metav1.GetOptions{}); rb != nil && rb.GetUID() != "" {
		err := k.Client().RbacV1().RoleBindings(namespace).Delete(ctx, roleBindName, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	}

	annotiations := map[string]string{}
	annotiations["group"] = strings.ToLower(group)

	subject := v12.Subject{}
	subject.Name = group
	subject.Kind = "Group"

	role := v12.RoleRef{}
	role.Kind = "ClusterRole"
	role.Name = roleName

	roleBinding = &v12.RoleBinding{}
	roleBinding.SetAnnotations(annotiations)
	roleBinding.SetName(roleBindName)
	roleBinding.SetNamespace(namespace)
	roleBinding.RoleRef = role
	roleBinding.Subjects = []v12.Subject{subject}

	return k.Client().RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
}

// Create rolebinding for group to gain access to namespace
func (k *Kubernetes) RoleBindingCreateNamespaceServiceAccount(namespace, serviceaccount, roleName string) (roleBinding *v12.RoleBinding, error error) {
	ctx := context.Background()
	roleBindName := fmt.Sprintf("serviceaccount:%s", serviceaccount)

	if rb, _ := k.Client().RbacV1().RoleBindings(namespace).Get(ctx, roleBindName, metav1.GetOptions{}); rb != nil && rb.GetUID() != "" {
		err := k.Client().RbacV1().RoleBindings(namespace).Delete(ctx, roleBindName, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	}

	annotiations := map[string]string{}
	annotiations["serviceaccount"] = strings.ToLower(serviceaccount)

	subject := v12.Subject{}
	subject.Name = serviceaccount
	subject.Kind = "ServiceAccount"

	role := v12.RoleRef{}
	role.Kind = "ClusterRole"
	role.Name = roleName

	roleBinding = &v12.RoleBinding{}
	roleBinding.SetAnnotations(annotiations)
	roleBinding.SetName(roleBindName)
	roleBinding.SetNamespace(namespace)
	roleBinding.RoleRef = role
	roleBinding.Subjects = []v12.Subject{subject}

	return k.Client().RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
}

func (k *Kubernetes) namespaceValidate(name string) (err error) {
	if k.Filter.Namespace != nil && !k.Filter.Namespace.MatchString(name) {
		err = fmt.Errorf("namespace %v not allowed", name)
	}

	return
}

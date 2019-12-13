package main

import (
	"devops-console/models"
	"fmt"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	v1Networking "k8s.io/api/networking/v1"
	v1Rbac "k8s.io/api/rbac/v1"
	"k8s.io/api/settings/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createKubernetesObjectList() (list *models.KubernetesObjectList) {
	list = &models.KubernetesObjectList{}
	list.ConfigMaps = map[string]models.KubernetesObject{}
	list.ServiceAccounts = map[string]models.KubernetesObject{}
	list.Roles = map[string]models.KubernetesObject{}
	list.RoleBindings = map[string]models.KubernetesObject{}
	list.ResourceQuotas = map[string]models.KubernetesObject{}
	list.NetworkPolicies = map[string]models.KubernetesObject{}
	list.PodPresets = map[string]models.KubernetesObject{}
	list.LimitRanges = map[string]models.KubernetesObject{}
	return
}

func buildKubeConfigList(defaultPath, path string) *models.KubernetesObjectList {
	kubeConfigList := createKubernetesObjectList()

	if defaultPath != "" {
		addK8sConfigsFromPath(defaultPath, kubeConfigList)
	}

	if path != "" {
		addK8sConfigsFromPath(path, kubeConfigList)
	}

	return kubeConfigList
}

func addK8sConfigsFromPath(configPath string, list *models.KubernetesObjectList) {
	var fileList []string
	err := filepath.Walk(configPath, func(path string, f os.FileInfo, err error) error {
		if IsK8sConfigFile(path) {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR: %s", err))
	}

	for _, path := range fileList {
		item := models.KubernetesObject{}
		item.Path = path
		item.Object = KubeParseConfig(path)

		switch item.Object.GetObjectKind().GroupVersionKind().Kind {
		case "ConfigMap":
			item.Name = item.Object.(*v1.ConfigMap).Name
			list.ConfigMaps[item.Name] = item
		case "ServiceAccount":
			item.Name = item.Object.(*v1.ServiceAccount).Name
			list.ServiceAccounts[item.Name] = item
		case "Role":
			item.Name = item.Object.(*v1Rbac.Role).Name
			list.Roles[item.Name] = item
		case "RoleBinding":
			item.Name = item.Object.(*v1Rbac.RoleBinding).Name
			list.RoleBindings[item.Name] = item
		case "NetworkPolicy":
			item.Name = item.Object.(*v1Networking.NetworkPolicy).Name
			list.NetworkPolicies[item.Name] = item
		case "LimitRange":
			item.Name = item.Object.(*v1.LimitRange).Name
			list.LimitRanges[item.Name] = item
		case "PodPreset":
			item.Name = item.Object.(*v1alpha1.PodPreset).Name
			list.PodPresets[item.Name] = item
		case "ResourceQuota":
			item.Name = item.Object.(*v1.ResourceQuota).Name
			list.ResourceQuotas[item.Name] = item
		default:
			panic("Not allowed object found: " + item.Object.GetObjectKind().GroupVersionKind().Kind)
		}
	}
}

func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func IsRegularFile(path string) bool {
	fileInfo, _ := os.Stat(path)
	return fileInfo.Mode().IsRegular()
}

func IsK8sConfigFile(path string) bool {
	if !IsRegularFile(path) {
		return false
	}

	switch filepath.Ext(path) {
	case ".json":
		return true
	case ".yaml":
		return true
	}

	return false
}

func recursiveFileListByPath(path string) (list []string) {
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if IsK8sConfigFile(path) {
			list = append(list, path)
		}
		return nil
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR: %s", err))
	}

	return
}

func KubeParseConfig(path string) runtime.Object {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		panic(err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		panic(err)
	}
	return obj
}

func stringToPtr(val string) *string {
	return &val
}


func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"-_+=*")
	var b strings.Builder
	for i := 0; i < length; i++ {
		if _, err := b.WriteRune(chars[rand.Intn(len(chars))]); err != nil {
			fmt.Println(fmt.Sprintf("ERROR: %s", err))
		}
	}
	return b.String()
}

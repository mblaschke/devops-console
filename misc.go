package main

import (
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"devops-console/models"
)

func createKubernetesObjectList() models.KubernetesObjectList {
	return models.KubernetesObjectList{}
}

func buildKubeConfigList(defaultPath, path string) models.KubernetesObjectList {
	kubeConfigList := createKubernetesObjectList()

	if defaultPath != "" {
		kubeConfigList = addK8sConfigsFromPath(defaultPath, kubeConfigList)
	}

	if path != "" {
		kubeConfigList = addK8sConfigsFromPath(path, kubeConfigList)
	}

	return kubeConfigList
}

func addK8sConfigsFromPath(configPath string, list models.KubernetesObjectList) models.KubernetesObjectList {
	var fileList []string
	err := filepath.Walk(configPath, func(path string, f os.FileInfo, err error) error {
		if IsK8sConfigFile(path) {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, path := range fileList {
		item := models.KubernetesObject{}
		item.Path = path
		item.Object = KubeParseConfig(path)

		list = append(list, item)
	}

	return list
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

func KubeParseConfig(path string) runtime.Object {
	data, err := os.ReadFile(filepath.Clean(path))
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

func randomString(length int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

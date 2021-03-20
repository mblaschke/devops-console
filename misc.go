package main

import (
	"devops-console/models"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createKubernetesObjectList() (models.KubernetesObjectList) {
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

func addK8sConfigsFromPath(configPath string, list models.KubernetesObjectList) (models.KubernetesObjectList) {
	var fileList []string
	err := filepath.Walk(configPath, func(path string, f os.FileInfo, err error) error {
		if IsK8sConfigFile(path) {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
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

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" +
		"-_+=*")
	var b strings.Builder
	for i := 0; i < length; i++ {
		if _, err := b.WriteRune(chars[rand.Intn(len(chars))]); err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}
	return b.String()
}

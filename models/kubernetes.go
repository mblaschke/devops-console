package models

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type (
	KubernetesObjectList []KubernetesObject

	KubernetesObject struct {
		Name   string
		Path   string
		Object runtime.Object
	}

	KubernetesNamespace struct {
		*v1.Namespace
	}
)

func (n *KubernetesNamespace) SettingsApply(settings map[string]string, config AppConfigKubernetes) {
	if n.Labels == nil {
		n.Labels = map[string]string{}
	}

	if n.Annotations == nil {
		n.Annotations = map[string]string{}
	}

	for _, setting := range config.Namespace.Settings {
		switch setting.K8sType {
		case "label":
			if val, ok := settings[setting.Name]; ok && val != "" {
				// label has content, set/add it
				switch setting.Type {
				case "hidden":
					n.Labels[setting.K8sName] = setting.K8sValue
				case "checkbox":
					val := frontendBoolToBackendBool(val)

					if val == "true" {
						k8sValue := val
						if setting.K8sValue != "" {
							k8sValue = setting.K8sValue
						}
						n.Labels[setting.K8sName] = k8sValue
					} else {
						// label has no content, delete it
						delete(n.Labels, setting.K8sName)
					}
				default:
					n.Labels[setting.K8sName] = val
				}
			} else {
				// label has no content, delete it
				delete(n.Labels, setting.K8sName)
			}
		case "annotation":
			if val, ok := settings[setting.Name]; ok && val != "" {
				// annotation has content, set/add it
				switch setting.Type {
				case "hidden":
					n.Annotations[setting.K8sName] = setting.K8sValue
				case "checkbox":
					val := frontendBoolToBackendBool(val)

					if val == "true" {
						k8sValue := val
						if setting.K8sValue != "" {
							k8sValue = setting.K8sValue
						}
						n.Annotations[setting.K8sName] = k8sValue
					} else {
						// label has no content, delete it
						delete(n.Annotations, setting.K8sName)
					}
				default:
					n.Annotations[setting.K8sName] = val
				}
			} else {
				// annotation has no content, delete it
				delete(n.Annotations, setting.K8sName)
			}
		}
	}
}

func (n *KubernetesNamespace) SettingsExtract(config AppConfigKubernetes) (ret map[string]string) {
	ret = map[string]string{}

	for _, setting := range config.Namespace.Settings {
		switch setting.K8sType {
		case "label":
			if val, ok := n.Labels[setting.K8sName]; ok {
				ret[setting.Name] = val
			} else {
				ret[setting.Name] = ""
			}
		case "annotation":
			if val, ok := n.Annotations[setting.K8sName]; ok {
				ret[setting.Name] = val
			} else {
				ret[setting.Name] = ""
			}
		}

		switch setting.Type {
		case "checkbox":
			if _, ok := ret[setting.Name]; ok {
				val := ret[setting.Name]
				if setting.K8sValue != "" && ret[setting.Name] == setting.K8sValue {
					val = "true"
				}
				ret[setting.Name] = backendBoolToFrontendBool(val)
			}
		}
	}

	return
}

func frontendBoolToBackendBool(value string) (ret string) {
	ret = "false"

	value = strings.ToLower(value)

	switch value {
	case "on":
		return "true"
	case "yes":
		return "true"
	case "1":
		return "true"
	case "true":
		return "true"
	case "enabled":
		return "true"
	}

	return ret
}

func backendBoolToFrontendBool(value string) (ret string) {
	ret = ""

	value = strings.ToLower(value)

	switch value {
	case "on":
		return "enabled"
	case "yes":
		return "enabled"
	case "1":
		return "enabled"
	case "true":
		return "enabled"
	}

	return ret
}

package models

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

type (
	KubernetesObjectList struct {
		ConfigMaps      map[string]KubernetesObject
		ServiceAccounts map[string]KubernetesObject
		Roles           map[string]KubernetesObject
		RoleBindings    map[string]KubernetesObject
		PodPresets      map[string]KubernetesObject
		NetworkPolicies map[string]KubernetesObject
		LimitRanges     map[string]KubernetesObject
		ResourceQuotas  map[string]KubernetesObject
	}

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
				case "checkbox":
					val := frontendBoolToBackendBool(val)

					if val == "true" {
						n.Labels[setting.K8sName] = val
					} else {
						// label has no content, delete it
						delete(n.Labels, setting.K8sName)
					}

					break
				default:
					n.Labels[setting.K8sName] = val
					break
				}
			} else {
				// label has no content, delete it
				delete(n.Labels, setting.K8sName)
			}
		case "annotation":
			if val, ok := settings[setting.Name]; ok && val != "" {
				// annotation has content, set/add it
				switch setting.Type {
				case "checkbox":
					val := frontendBoolToBackendBool(val)

					if val == "true" {
						n.Annotations[setting.K8sName] = val
					} else {
						// label has no content, delete it
						delete(n.Annotations, setting.K8sName)
					}

					break
				default:
					n.Annotations[setting.K8sName] = val
					break
				}
			} else {
				// annotation has no content, delete it
				delete(n.Annotations, setting.K8sName)
			}
		}
	}

	return
}

func (n *KubernetesNamespace) SettingsExtract(config AppConfigKubernetes) (ret map[string]string) {
	ret = map[string]string{}

	for _, setting := range config.Namespace.Settings {
		switch setting.K8sType {
		case "label":
			if val, ok := n.Labels[setting.K8sName]; ok {
				ret[setting.Name] = val
			}
			break
		case "annotation":
			if val, ok := n.Annotations[setting.K8sName]; ok {
				ret[setting.Name] = val
			}
			break
		}

		switch setting.Type {
		case "checkbox":
			if _, ok := ret[setting.Name]; ok {
				ret[setting.Name] = backendBoolToFrontendBool(ret[setting.Name])
			}
			break
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

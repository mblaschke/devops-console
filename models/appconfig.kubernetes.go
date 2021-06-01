package models

import networkingV1 "k8s.io/api/networking/v1"

type (
	AppConfigKubernetes struct {
		ObjectsPath string `yaml:"objectsPath"`

		ObjectsList map[string]KubernetesObjectList

		RoleBinding AppConfigKubernetesRoleBindingMetaData `yaml:"roleBinding"`

		Environments []AppConfigKubernetesEnvironment `yaml:"environments"`

		Namespace AppConfigKubernetesNamespace
	}

	AppConfigKubernetesRoleBindingMetaData struct {
		Annotations map[string]string `yaml:"annotations"`
		Labels      map[string]string `yaml:"labels"`
	}

	AppConfigNamespaceSettings struct {
		Name           string
		Label          string
		Description    string
		K8sType        string
		K8sName        string
		K8sValue       string
		Type           string
		Default        string
		Placeholder    string
		Validation     AppInputValidation
		Transformation AppInputTransformation
	}

	AppConfigKubernetesNamespace struct {
		Filter struct {
			Access string
			Delete string
			User   string
			Team   string
		}

		Validation struct {
			App  string
			Team string
		}

		Annotations struct {
			Description   string
			Immortal      string
			NetworkPolicy string `yaml:"networkPolicy"`
		}

		Labels struct {
			Name        string
			User        string
			Team        string
			Environment string
		}

		Role struct {
			Team    string
			User    string
			Private bool
		}

		Quota struct {
			User int
			Team int
		}

		Settings      []AppConfigNamespaceSettings
		NetworkPolicy []AppConfigKubernetesNetworkPolicy `yaml:"networkPolicy"`
	}

	AppConfigKubernetesNetworkPolicy struct {
		Name        string
		Description string
		Path        string
		Default     bool
		netpol      *networkingV1.NetworkPolicy
	}

	ApplicationKubernetesNetworkPolicy struct {
		Name        string
		Description string
	}
)

func (netpol *AppConfigKubernetesNetworkPolicy) SetKubernetesObject(obj *networkingV1.NetworkPolicy) {
	netpol.netpol = obj
}

func (netpol *AppConfigKubernetesNetworkPolicy) GetKubernetesObject() *networkingV1.NetworkPolicy {
	return netpol.netpol
}

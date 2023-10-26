package models

import networkingV1 "k8s.io/api/networking/v1"

type (
	AppConfigKubernetes struct {
		ObjectsPath string `yaml:"objectsPath"`

		ObjectsList map[string]KubernetesObjectList

		RoleBinding AppConfigKubernetesRoleBindingMetaData `yaml:"roleBinding"`

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
		ClusterRoleName string `yaml:"clusterRoleName"`

		LabelSelector string `yaml:"labelSelector"`

		Filter struct {
			Access AppConfigFilter
			Delete AppConfigFilter
		}

		Validation struct {
			Namespace AppConfigFilter
			Team      AppConfigFilter
		}

		Annotations struct {
			Description   string
			Immortal      string
			NetworkPolicy string `yaml:"networkPolicy"`
		}

		Labels struct {
			Team      string
			ManagedBy string `yaml:"managedBy"`
		}

		Quota struct {
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

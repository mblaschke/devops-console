package response

type (
	ResponseConfig struct {
		User         ResponseConfigUser         `json:"user"`
		Teams        []ResponseConfigTeam       `json:"teams"`
		Quota        map[string]int             `json:"quota"`
		Azure        ResponseConfigAzure        `json:"azure"`
		Kubernetes   ResponseConfigKubernetes   `json:"kubernetes"`
		Alertmanager ResponseConfigAlertmanager `json:"alertmanager"`
	}

	ResponseConfigUser struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	}

	ResponseConfigTeam struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	ResponseConfigAzure struct {
		ResourceGroup  ResponseConfigAzureResourceGroup  `json:"resourceGroup"`
		RoleAssignment ResponseConfigAzureRoleAssignment `json:"roleAssignment"`
	}

	ResponseConfigAzureResourceGroup struct {
		Tags []ResponseConfigAzureResourceGroupTag `json:"tags"`
	}

	ResponseConfigAzureRoleAssignment struct {
		RoleDefinitions []string `json:"roleDefinitions"`
		Ttl             []string `json:"ttl"`
	}

	ResponseConfigAzureResourceGroupTag struct {
		Name        string `json:"name"`
		Label       string `json:"label"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Default     string `json:"default"`
		Placeholder string `json:"placeholder"`
	}

	ResponseConfigKubernetes struct {
		Namespace    ResponseConfigKubernetesNamespace               `json:"namespace"`
		Environments []ResponseConfigKubernetesNamespaceEnvironments `json:"environments"`
	}

	ResponseConfigKubernetesNamespaceEnvironments struct {
		Environment string `json:"environment"`
		Description string `json:"description"`
		Template    string `json:"template"`
	}

	ResponseConfigKubernetesNamespace struct {
		NetworkPolicy []ResponseConfigKubernetesNamespaceNetworkPolicy `json:"networkPolicy"`
		Settings      []ResponseConfigKubernetesNamespaceSetting       `json:"settings"`
	}

	ResponseConfigKubernetesNamespaceSetting struct {
		Name        string `json:"name"`
		Label       string `json:"label"`
		Description string `json:"description"`
		K8sType     string `json:"k8sType"`
		K8sName     string `json:"k8sName"`
		Type        string `json:"type"`
		Default     string `json:"default"`
		Placeholder string `json:"placeholder"`
		Required    bool   `json:"required"`
	}

	ResponseConfigKubernetesNamespaceNetworkPolicy struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	ResponseConfigAlertmanager struct {
		Instances []string `json:"instances"`
	}
)

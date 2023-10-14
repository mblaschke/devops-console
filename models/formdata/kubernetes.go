package formdata

type (
	KubernetesNamespaceCreate struct {
		Name          *string            `json:"name"`
		Team          *string            `json:"team"`
		Description   *string            `json:"description"`
		NetworkPolicy *string            `json:"networkPolicy"`
		Settings      *map[string]string `json:"settings"`
	}
)

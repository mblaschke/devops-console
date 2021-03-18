package formdata

type (
	KubernetesNamespaceCreate struct {
		Environment   *string            `json:"environment"`
		Team          *string            `json:"team"`
		App           *string            `json:"app"`
		Description   *string            `json:"description"`
		NetworkPolicy *string            `json:"networkPolicy"`
		Settings      *map[string]string `json:"settings"`
	}
)

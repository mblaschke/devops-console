package formdata

type (
	KubernetesNamespaceCreate struct {
		Environment   *string
		Team          *string
		App           *string
		Description   *string
		NetworkPolicy *string `json:"networkPolicy"`
		Settings      *map[string]string
	}
)

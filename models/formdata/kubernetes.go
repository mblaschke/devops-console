package formdata

type (
	KubernetesNamespaceCreate struct {
		Environment *string
		Team        *string
		App         *string
		Description *string
		Settings    *map[string]string
	}
)

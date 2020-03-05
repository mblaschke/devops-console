package formdata

type (
	KubernetesNamespaceCreate struct {
		Environment *string
		Team        *string
		App         *string
		Description *string
		Netpol      *string
		Settings    *map[string]string
	}
)

package response

type (
	KubernetesNamespace struct {
		Name          string
		PodCount      *int64
		Environment   string
		Description   string
		NetworkPolicy string
		OwnerTeam     string
		OwnerUser     string
		Status        string
		Created       string
		CreatedAgo    string
		Deleteable    bool
		Settings      map[string]string
	}
)

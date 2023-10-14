package response

type (
	KubernetesNamespace struct {
		Name          string            `json:"name"`
		PodCount      *int64            `json:"podCount"`
		Description   string            `json:"description"`
		NetworkPolicy string            `json:"networkPolicy"`
		OwnerTeam     string            `json:"ownerTeam"`
		Status        string            `json:"status"`
		Created       string            `json:"created"`
		CreatedAgo    string            `json:"createdAgo"`
		Deleteable    bool              `json:"deleteable"`
		Settings      map[string]string `json:"settings"`
	}
)

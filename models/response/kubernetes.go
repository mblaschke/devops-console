package response

type (
	KubernetesCluster struct {
		Name              string
		Role              string
		Version           string
		SpecArch          string
		SpecOS            string
		SpecMachineCPU    string
		SpecMachineMemory string
		SpecRegion        string
		SpecZone          string
		SpecInstance      string
		InternalIp        string
		Status            string
		Created           string
		CreatedAgo        string
	}

	KubernetesNamespace struct {
		Name        string
		PodCount    *int64
		Environment string
		Description string
		Netpol      string
		OwnerTeam   string
		OwnerUser   string
		Status      string
		Created     string
		CreatedAgo  string
		Deleteable  bool
		Settings    map[string]string
	}
)

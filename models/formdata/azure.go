package formdata

type (
	AzureResourceGroup struct {
		Name     string
		Location string
		Team     string
		Personal bool
		Tag      map[string]string
	}
)

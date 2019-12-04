package formdata

import (
	"github.com/prometheus/alertmanager/api/v2/models"
)

type (
	AlertmanagerForm struct {
		Team string
		Silence models.Silence
	}
)

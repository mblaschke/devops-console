package response

type (
	GeneralStats struct {
		Name  string             `json:"name"`
		Stats []GeneralStatsLine `json:"stats"`
	}

	GeneralStatsLine struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	GeneralMessage struct {
		Message string `json:"message"`
	}
)

func NewGeneralStats(name string) *GeneralStats {
	ret := GeneralStats{}
	ret.Name = name
	ret.Stats = []GeneralStatsLine{}
	return &ret
}

func (s *GeneralStats) Add(name, value string) {
	s.Stats = append(s.Stats, GeneralStatsLine{Name: name, Value: value})
}

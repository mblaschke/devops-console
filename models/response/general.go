package response

type (
	GeneralStats struct {
		Name  string
		Stats []GeneralStatsLine
	}

	GeneralStatsLine struct {
		Name  string
		Value string
	}

	GeneralMessage struct {
		Message string
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

package input

// Mode represents the application mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeSearch
	ModeSort
)

// String returns the string representation of the mode
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeSearch:
		return "SEARCH"
	case ModeSort:
		return "SORT"
	default:
		return "UNKNOWN"
	}
}

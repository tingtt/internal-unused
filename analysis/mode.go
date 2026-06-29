package analysis

// Mode controls which declarations are reported as unused.
type Mode int

const (
	// ModeAll reports declarations that are unused in both production and test code.
	ModeAll Mode = iota
	// ModeProduction reports declarations unused in production code,
	// distinguishing test-only usage from completely unused.
	ModeProduction
)

// ParseMode parses a mode string. Returns (mode, true) on success.
func ParseMode(s string) (Mode, bool) {
	switch s {
	case "all":
		return ModeAll, true
	case "production":
		return ModeProduction, true
	default:
		return 0, false
	}
}

func (m Mode) String() string {
	switch m {
	case ModeAll:
		return "all"
	case ModeProduction:
		return "production"
	default:
		return "unknown"
	}
}

package account

// ToolFlagSkipSwitcher excludes an account from the window-switcher cycling loop.
const ToolFlagSkipSwitcher uint32 = 1 << 0

// ToolFlagOption describes a single per-account tool flag bit.
type ToolFlagOption struct {
	Bit         uint32
	Name        string
	Description string
}

var toolFlagOptions = []ToolFlagOption{
	{
		Bit:         ToolFlagSkipSwitcher,
		Name:        "跳過切換",
		Description: "switcher 不切換到此帳號",
	},
}

// ToolFlagOptions returns a copy of all supported tool flag options.
func ToolFlagOptions() []ToolFlagOption {
	out := make([]ToolFlagOption, len(toolFlagOptions))
	copy(out, toolFlagOptions)
	return out
}

// SupportedToolFlagsMask returns a bitmask of all supported tool flags.
func SupportedToolFlagsMask() uint32 {
	var mask uint32
	for _, o := range toolFlagOptions {
		mask |= o.Bit
	}
	return mask
}

// SanitizeToolFlags clears any bits not in SupportedToolFlagsMask.
func SanitizeToolFlags(flags uint32) uint32 {
	return flags & SupportedToolFlagsMask()
}

// SkipSwitcher reports whether the ToolFlagSkipSwitcher bit is set.
func SkipSwitcher(flags uint32) bool {
	return flags&ToolFlagSkipSwitcher != 0
}

// ExcludedFromSwitcher returns the DisplayNames of accounts that have ToolFlagSkipSwitcher set.
func ExcludedFromSwitcher(accounts []Account) []string {
	var names []string
	for _, a := range accounts {
		if SkipSwitcher(a.ToolFlags) {
			names = append(names, a.DisplayName)
		}
	}
	return names
}

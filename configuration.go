package mikrogen

// defines configuration of blocking
type Configuration struct {
	IdentifierPrefix string

	AccessFilters map[string]AccessFilter
}

// Defines a filter for restricting access to specific addresses and TLS hosts.
type AccessFilter struct {
	// List of addresses to be managed
	TargetAddresses []string
	// List of addresses specifying TLS/HTTPS access filtering.
	TargetTLSHosts  []string

	DisableIntervals []Interval
}

type Interval struct {
	Start, End string
}

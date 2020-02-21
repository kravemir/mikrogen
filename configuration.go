package mikrogen

// defines configuration of blocking
type Configuration struct {
	IdentifierPrefix string

	DNSBlockedAddresses []string
	TLSBlockedAddresses []string

	DisableIntervals []Interval
}

type Interval struct {
	Start, End string
}

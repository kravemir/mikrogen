package mikrogen

// defines configuration of blocking
type Configuration struct {
	IdentifierPrefix string

	AccessBlockers map[string]AccessBlocker
}

type AccessBlocker struct {
	DNSBlockedAddresses []string
	TLSBlockedAddresses []string

	DisableIntervals []Interval
}

type Interval struct {
	Start, End string
}

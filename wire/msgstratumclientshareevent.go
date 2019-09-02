package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientShareEventMsg{}

type ShareEvent int8

const (
	ShareEventSubmitted ShareEvent = 1
	ShareEventAccepted  ShareEvent = 2
	ShareEventDeclined  ShareEvent = 3
)

// StratumClientShareEventMsg is sent from the stratum client
// to the hub for share-level events (submitted, accepted, declined)
type StratumClientShareEventMsg struct {
	Header   MsgHeader
	ShareID  int64
	Event    ShareEvent
	Observed int64
	Details  string
}

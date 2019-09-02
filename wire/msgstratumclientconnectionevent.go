package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientConnectionEventMsg{}

type ConnectionEvent int8

const (
	ConnectionEventConnected               ConnectionEvent = 1
	ConnectionEventDisconnected            ConnectionEvent = 2
	ConnectionEventAuthenticationSucceeded ConnectionEvent = 3
	ConnectionEventAuthenticationFailed    ConnectionEvent = 4
	ConnectionEventShareSubmitted          ConnectionEvent = 5
	ConnectionEventShareAccepted           ConnectionEvent = 6
	ConnectionEventShareDeclined           ConnectionEvent = 7
)

// StratumClientConnectionEventMsg is sent from the stratum client
// to the hub for connection-level events (connected, disconnected, authenticated, etc)
type StratumClientConnectionEventMsg struct {
	Header         MsgHeader
	PoolObserverID int
	Event          ConnectionEvent
	Observed       int64
}

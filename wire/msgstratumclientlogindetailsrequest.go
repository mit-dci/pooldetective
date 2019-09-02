package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientLoginDetailsRequestMsg{}

// StratumClientLoginDetailsRequestMsg is sent from the stratum client
// to the hub to indicate it wants to receive stratum details
// to connect to (host, port, login, pwd)
type StratumClientLoginDetailsRequestMsg struct {
	Header         MsgHeader
	PoolObserverID int
}

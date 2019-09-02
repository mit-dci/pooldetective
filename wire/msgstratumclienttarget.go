package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientTargetMsg{}

// StratumClientTargetMsg is sent from the stratum client to the hub
// when it received a new target
type StratumClientTargetMsg struct {
	Header         MsgHeader
	PoolObserverID int
	Target         []byte
}

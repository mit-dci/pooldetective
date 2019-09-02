package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientExtraNonceMsg{}

// StratumClientExtraNonceMsg is sent from the stratum client to the hub
// when it received a change in extranonce data
type StratumClientExtraNonceMsg struct {
	Header          MsgHeader
	PoolObserverID  int
	ExtraNonce1     []byte
	ExtraNonce2Size int8
}

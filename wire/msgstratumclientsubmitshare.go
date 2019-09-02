package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientSubmitShareMsg{}

// StratumClientSubmitShareMsg is sent from the hub to the stratum client
// to have the stratum client submit the share
type StratumClientSubmitShareMsg struct {
	Header                 MsgHeader
	ShareID                int64
	PoolJobID              []byte
	ExtraNonce2            []byte
	Time                   int64
	Nonce                  int64
	AdditionalSolutionData [][]byte
}

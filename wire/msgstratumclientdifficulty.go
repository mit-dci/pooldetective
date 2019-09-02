package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientDifficultyMsg{}

// StratumClientDifficultyMsg is sent from the stratum client to the hub
// when it received a change in difficulty
type StratumClientDifficultyMsg struct {
	Header         MsgHeader
	PoolObserverID int
	Difficulty     float64
}

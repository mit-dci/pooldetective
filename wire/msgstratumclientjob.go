package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientJobMsg{}

// StratumClientJobMsg is sent from the stratum client to the hub
// when it received a new job
type StratumClientJobMsg struct {
	Header            MsgHeader
	PoolObserverID    int
	Observed          int64
	JobID             []byte
	PreviousBlockHash []byte
	GenTX1            []byte
	GenTX2            []byte
	MerkleBranches    [][]byte
	BlockVersion      []byte
	DifficultyBits    []byte
	Timestamp         int64
	CleanJobs         bool
	Reserved          []byte
}

package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &CoordinatorGetConfigResponseMsg{}

type CoordinatorGetConfigResponseStratumServer struct {
	ID          int
	AlgorithmID int
	Port        int
	Protocol    int
}
type CoordinatorGetConfigResponseMsg struct {
	Header                       MsgHeader
	StratumServers               []CoordinatorGetConfigResponseStratumServer
	StratumClientPoolObserverIDs []int
}

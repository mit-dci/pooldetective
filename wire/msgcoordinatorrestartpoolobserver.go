package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &CoordinatorRestartPoolObserverMsg{}

type CoordinatorRestartPoolObserverMsg struct {
	Header         MsgHeader
	PoolObserverID int
}

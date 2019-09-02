package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &CoordinatorGetConfigRequestMsg{}

type CoordinatorGetConfigRequestMsg struct {
	Header     MsgHeader
	LocationID int
}

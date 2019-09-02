package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &CoordinatorRefreshConfigMsg{}

type CoordinatorRefreshConfigMsg struct {
	Header MsgHeader
}

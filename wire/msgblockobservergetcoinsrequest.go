package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &BlockObserverBlockObservedMsg{}

// BlockObserverGetCoinsRequestMsg is sent from the block observer to the hub
// to discover the known coins and their ID
type BlockObserverGetCoinsRequestMsg struct {
	Header MsgHeader
}

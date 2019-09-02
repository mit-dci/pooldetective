package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &BlockObserverBlockObservedMsg{}

// BlockObserverBlockObservedMsg is sent from the block observer to the hub
// when one of its peers announces a block
type BlockObserverBlockObservedMsg struct {
	Header     MsgHeader
	LocationID int
	Observed   int64
	CoinID     int
	PeerIP     []byte
	PeerPort   int
	BlockHash  [32]byte
}

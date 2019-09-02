package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &BlockObserverBlockObservedMsg{}

type BlockObserverGetCoinsResponseCoin struct {
	Ticker string
	CoinID int
}

// BlockObserverBlockObservedMsg is sent from the block observer to the hub
// when one of its peers announces a block
type BlockObserverGetCoinsResponseMsg struct {
	Header MsgHeader
	Coins  []BlockObserverGetCoinsResponseCoin
}

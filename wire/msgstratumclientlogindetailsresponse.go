package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &StratumClientLoginDetailsResponseMsg{}

// StratumClientLoginDetailsResponseMsg is sent from the stratum client
// to the hub to indicate it wants to receive stratum details
// to connect to (host, port, login, pwd)
type StratumClientLoginDetailsResponseMsg struct {
	Header   MsgHeader
	Host     string
	Port     int
	Login    string
	Password string
	Protocol int
}

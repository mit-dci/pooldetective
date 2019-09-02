package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &AckMsg{}

type AckMsg struct {
	Header MsgHeader
	YourID int
}

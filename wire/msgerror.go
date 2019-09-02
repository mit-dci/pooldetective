package wire

// Compile-time type assertion
var _ PoolDetectiveMsg = &ErrorMsg{}

type ErrorMsg struct {
	Header MsgHeader
	YourID int
	Error  string
}

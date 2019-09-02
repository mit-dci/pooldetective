// Copyright (c) 2019 The bchd developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"
)

// MsgXVerAck defines a bitcoin xverack message which is used for a peer to
// acknowledge an xversion message (MsgXVersion).
// It implements the Message interface.
//
// This message has no payload.
type MsgXVerAck struct{}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgXVerAck) BtcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgXVerAck) BtcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	return nil
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgXVerAck) Command() string {
	return CmdXVerAck
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgXVerAck) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgXVerAck returns a new bitcoin verack message that conforms to the
// Message interface.
func NewMsgXVerAck() *MsgXVerAck {
	return &MsgXVerAck{}
}

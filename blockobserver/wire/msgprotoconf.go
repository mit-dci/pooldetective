// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"

	"github.com/mit-dci/pooldetective/blockobserver/logging"
)

type MsgProtoConf struct {
	NumFields            uint64
	MaxRecvPayloadLength uint32
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgProtoConf) BtcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	logging.Debugf("Reading protoconf message")
	var err error
	msg.NumFields, err = ReadVarInt(r, pver)
	if err != nil {
		logging.Debugf("Could not read numfields: %v", err)
		return err
	}

	if msg.NumFields > 0 {
		err = readElement(r, &msg.MaxRecvPayloadLength)
		if err != nil {
			logging.Debugf("Could not read MaxRecvPayloadLength: %v", err)
			return err
		}
	}
	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgProtoConf) BtcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	logging.Debugf("Writing protoconf message")
	err := WriteVarInt(w, pver, msg.NumFields)
	if err != nil {
		logging.Debugf("Could not write numfields: %v", err)
		return err
	}
	if msg.NumFields > 0 {
		err = writeElement(w, msg.MaxRecvPayloadLength)
		if err != nil {
			logging.Debugf("Could not write MaxRecvPayloadLength: %v", err)
			return err
		}
	}
	return nil
}

func (msg *MsgProtoConf) Command() string {
	return CmdProtoConf
}

func (msg *MsgProtoConf) MaxPayloadLength(pver uint32) uint32 {
	return 12
}

func NewMsgProtoConf(maxRecvPayloadLength uint32) *MsgProtoConf {
	return &MsgProtoConf{
		NumFields:            1,
		MaxRecvPayloadLength: maxRecvPayloadLength,
	}
}

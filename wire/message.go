package wire

import (
	"log"
	"reflect"
)

type PoolDetectiveMsg interface {
}

type MsgHeader struct {
	ID int
}

func GetMessageID(m PoolDetectiveMsg) int {
	val := reflect.ValueOf(m)
	if val.IsZero() {
		return -1
	}

	if val.Kind() != reflect.Struct {
		return -1
	}

	hdr := val.FieldByName("MsgHeader")
	if hdr.IsZero() {
		return -1
	}

	id := hdr.FieldByName("ID")
	if id.IsZero() {
		return -1
	}

	idInt, ok := id.Elem().Interface().(int)
	if !ok {
		return -1
	}
	return idInt
}

func NewMessage(mt MessageType) PoolDetectiveMsg {
	newType, ok := MessageTypeToTypeMap[mt]
	if !ok {
		log.Printf("Unable to create message of type %d", mt)
		return nil
	}

	return reflect.New(newType.Elem()).Interface()
}

func GetMessageType(m PoolDetectiveMsg) MessageType {
	mt, ok := TypeToMessageTypeMap[reflect.TypeOf(m)]
	if !ok {
		log.Printf("Unable to detect type of message %T", m)
		return MessageType(0)
	}
	return mt
}

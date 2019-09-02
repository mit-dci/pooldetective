package wire

import "reflect"

type MessageType int16

var TypeToMessageTypeMap = map[reflect.Type]MessageType{
	reflect.TypeOf(&BlockObserverBlockObservedMsg{}):        MessageType(1),
	reflect.TypeOf(&StratumClientConnectionEventMsg{}):      MessageType(2),
	reflect.TypeOf(&StratumClientDifficultyMsg{}):           MessageType(3),
	reflect.TypeOf(&StratumClientExtraNonceMsg{}):           MessageType(4),
	reflect.TypeOf(&StratumClientJobMsg{}):                  MessageType(5),
	reflect.TypeOf(&StratumClientLoginDetailsRequestMsg{}):  MessageType(6),
	reflect.TypeOf(&StratumClientLoginDetailsResponseMsg{}): MessageType(7),
	reflect.TypeOf(&StratumClientShareEventMsg{}):           MessageType(8),
	reflect.TypeOf(&StratumClientSubmitShareMsg{}):          MessageType(9),
	reflect.TypeOf(&AckMsg{}):                               MessageType(10),
	reflect.TypeOf(&ErrorMsg{}):                             MessageType(11),
	reflect.TypeOf(&BlockObserverGetCoinsRequestMsg{}):      MessageType(12),
	reflect.TypeOf(&BlockObserverGetCoinsResponseMsg{}):     MessageType(13),
	reflect.TypeOf(&CoordinatorGetConfigRequestMsg{}):       MessageType(14),
	reflect.TypeOf(&CoordinatorGetConfigResponseMsg{}):      MessageType(15),
	reflect.TypeOf(&CoordinatorRefreshConfigMsg{}):          MessageType(16),
	reflect.TypeOf(&CoordinatorRestartPoolObserverMsg{}):    MessageType(17),
	reflect.TypeOf(&StratumClientTargetMsg{}):               MessageType(18),
}

var MessageTypeToTypeMap = map[MessageType]reflect.Type{}

func init() {
	for k, v := range TypeToMessageTypeMap {
		MessageTypeToTypeMap[v] = k
	}
}

package serializer

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtobufToJSON converts protocol buffer message to JSON string
func ProtobufToJSON(message proto.Message) (string, error) {
	b := protojson.MarshalOptions{
		Indent:          " ",
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	data, err := b.Marshal(message)
	return string(data), err
}

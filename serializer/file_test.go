package serializer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thewalkers2012/grpc-example/pb"
	"github.com/thewalkers2012/grpc-example/sample"
	"github.com/thewalkers2012/grpc-example/serializer"
	"google.golang.org/protobuf/proto"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop"
	jsonFile := "../tmp/laptop.json"

	labtop1 := sample.NewLaptop()
	err := serializer.WriteProtobufToBinaryFile(labtop1, binaryFile)
	assert.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = serializer.ReadProtobufFromBinaryFile(laptop2, binaryFile)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(labtop1, laptop2))

	err = serializer.WriteProtobufToJSONFile(labtop1, jsonFile)
	assert.NoError(t, err)
}

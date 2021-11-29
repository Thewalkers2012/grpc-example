package service_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thewalkers2012/grpc-example/pb"
	"github.com/thewalkers2012/grpc-example/sample"
	"github.com/thewalkers2012/grpc-example/serializer"
	"github.com/thewalkers2012/grpc-example/service"
	"google.golang.org/grpc"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := service.NewInMemoryLaptopStore()
	serverAddress := startTestLaptopServer(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, expectedID, res.Id)

	// check that the laptop is saved to the store
	other, err := laptopStore.Find(res.Id)
	assert.NoError(t, err)
	assert.NotNil(t, other)

	// check that the saved laptop is the same as the one we sent
	requireSameLaptop(t, laptop, other)
}

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	store := service.NewInMemoryLaptopStore()
	expectedIDs := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 4086, Unit: pb.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		default:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &pb.Memory{Value: 64, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		}
		err := store.Save(laptop)
		assert.NoError(t, err)
	}

	serverAddress := startTestLaptopServer(t, store, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	assert.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		log.Print(res)
		assert.NoError(t, err)
		assert.True(t, expectedIDs[res.Laptop.GetId()])

		found++
	}

	assert.Equal(t, len(expectedIDs), found)
}

func TestClientUploadImage(t *testing.T) {
	t.Parallel()

	testImageFolder := "../tmp"

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore(testImageFolder)

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	assert.NoError(t, err)

	serverAddress := startTestLaptopServer(t, laptopStore, imageStore)
	laptopClient := newTestLaptopClient(t, serverAddress)

	imagePath := fmt.Sprintf("%s/laptop.jpeg", testImageFolder)
	file, err := os.Open(imagePath)
	assert.NoError(t, err)
	defer file.Close()

	stream, err := laptopClient.UploadImage(context.Background())
	assert.NoError(t, err)

	imageType := filepath.Ext(imagePath)
	req := &pb.UploadmageRequest{
		Data: &pb.UploadmageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:   laptop.GetId(),
				ImageTypes: imageType,
			},
		},
	}

	err = stream.Send(req)
	assert.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		assert.NoError(t, err)
		size += n

		req := &pb.UploadmageRequest{
			Data: &pb.UploadmageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		assert.NoError(t, err)
	}

	res, err := stream.CloseAndRecv()
	assert.NoError(t, err)
	assert.NotZero(t, res.GetId())
	assert.EqualValues(t, res.GetSize(), size)

	saveImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, res.GetId(), imageType)
	assert.FileExists(t, saveImagePath)
	assert.NoError(t, os.Remove(saveImagePath))
}

func startTestLaptopServer(t *testing.T, laptopStore service.LaptopStore, imageStore service.ImageStore) string {
	laptopServer := service.NewLaptopService(laptopStore, imageStore)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	val := 8000 + rand.Intn(100)
	address := fmt.Sprintf("127.0.0.1:%d", val)
	log.Println("address", address)

	listener, err := net.Listen("tcp", address) // random available port
	assert.NoError(t, err)

	go grpcServer.Serve(listener) // non block

	return listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	assert.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	assert.NoError(t, err)
	json2, err := serializer.ProtobufToJSON(laptop2)
	assert.NoError(t, err)
	assert.Equal(t, json1, json2)
}

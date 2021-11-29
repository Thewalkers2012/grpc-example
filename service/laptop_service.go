package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/thewalkers2012/grpc-example/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopService is the service that provides laptop services
type LaptopServer struct {
	Store LaptopStore
	pb.UnimplementedLaptopServiceServer
}

// NewLaptopService returns a new laptopServer
func NewLaptopService(store LaptopStore) *LaptopServer {
	return &LaptopServer{
		Store: store,
	}
}

func (s *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive a create-laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		// check if it's a valid UUID
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// save the laptop to in-memory store
	err := s.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop to the store: %v", err)
	}

	log.Printf("save laptop with id: %s", laptop.Id)

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}

	return res, nil
}

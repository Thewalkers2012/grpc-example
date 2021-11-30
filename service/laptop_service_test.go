package service_test

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thewalkers2012/grpc-example/pb"
	"github.com/thewalkers2012/grpc-example/sample"
	"github.com/thewalkers2012/grpc-example/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid-uuid"

	laptopDuplicateID := sample.NewLaptop()
	storeDuplicateID := service.NewInMemoryLaptopStore()
	err := storeDuplicateID.Save(laptopDuplicateID)
	assert.Nil(t, err)

	testCases := []struct {
		name   string
		laptop *pb.Laptop
		store  service.LaptopStore
		code   codes.Code
	}{
		{
			name:   "success_with_id",
			laptop: sample.NewLaptop(),
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "success_no_id",
			laptop: laptopNoID,
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "failure_invalid_id",
			laptop: laptopInvalidID,
			store:  service.NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
		{
			name:   "failure_duplicate_id",
			laptop: laptopDuplicateID,
			store:  storeDuplicateID,
			code:   codes.AlreadyExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := service.NewLaptopService(tc.store, nil, nil)
			res, err := server.CreateLaptop(context.Background(), req)
			log.Println(tc.name, tc.code)
			if tc.code == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.NotEmpty(t, res.Id)
				if len(tc.laptop.Id) > 0 {
					assert.Equal(t, tc.laptop.Id, res.Id)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, res)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.code, st.Code())
			}
		})
	}
}

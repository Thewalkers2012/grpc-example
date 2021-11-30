package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/thewalkers2012/grpc-example/pb"
	"github.com/thewalkers2012/grpc-example/service"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on post %d", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopService(laptopStore, imageStore, ratingStore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("can not start server: ", err)
	}

	err = grpcServer.Serve(listen)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}

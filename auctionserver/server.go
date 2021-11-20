package main

import (
	"log"
	"net"

	"github.com/kaeppen/disys-miniproject3/auctionator"
	a "github.com/kaeppen/disys-miniproject3/auctionator"
	"google.golang.org/grpc"
)

type Server struct {
	a.UnimplementedAuctionatorServer
}

func main() {
	var port = ":8080" //replace with some docker stuff
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %v", err)
	}
	grpcServer := grpc.NewServer()
	auctionator.RegisterAuctionatorServer(grpcServer, &Server{})
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

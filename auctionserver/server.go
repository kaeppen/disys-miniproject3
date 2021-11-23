package main

import (
	"context"
	"log"
	"net"

	"github.com/kaeppen/disys-miniproject3/auctionator"
	a "github.com/kaeppen/disys-miniproject3/auctionator"
	"google.golang.org/grpc"
)

type Server struct {
	a.UnimplementedAuctionatorServer
	Id         int32
	isPrimary  bool                          //am i primary
	primary    a.AuctionatorServer           //who is my primary
	backups    map[int32]a.AuctionatorServer //the backups
	clients    map[int32]a.AuctionatorClient //the registered clients
	highestBid Bid
	isOver     bool
	result     int32
}

type Bid struct {
	ClientId int32
	Bid      int32
}

func (s *Server) Bid(ctx context.Context, amount *a.Amount) (*a.Acknowledgement, error) {
	var client = amount.ClientId
	if s.clients[client] == nil {
		s.clients[client] = nil // nil er placeholder, hvordan lægger vi en klient ind? skal det være en stream eller en ip f.eks.? se evt. inspiration i chitty
	}
	var ack = &a.Acknowledgement{}
	if amount.Amount > s.highestBid.Bid {
		ack.Ack = "Success"
	} else {
		ack.Ack = "Fail"
	}
	//what about exception case? (as noted in the requirements)
	return ack, nil
}

func (s *Server) Result(ctx context.Context, void *a.Empty) (*a.Outcome, error) {
	var outcome = &a.Outcome{}
	if s.isOver {
		outcome.Over = true
		outcome.Result = s.result
	} else {
		outcome.Over = false
		outcome.Result = s.highestBid.Bid
	}

	return outcome, nil
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
	//husk at initialisere map et eller andet sted
}

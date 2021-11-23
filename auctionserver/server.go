package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/kaeppen/disys-miniproject3/auctionator"
	a "github.com/kaeppen/disys-miniproject3/auctionator"
	"google.golang.org/grpc"
)

type Server struct {
	a.UnimplementedAuctionatorServer
	Id         int32
	isPrimary  bool                          //am i primary
	primary    a.AuctionatorServer           //who is my primary
	backups    map[int32]Server              //the backups
	clients    map[int32]a.AuctionatorClient //the registered clients
	highestBid Bid
	isOver     bool
	result     int32
	responses  map[int32]string //hvordan skal en server huske hvad den har svaret + husk at bid/result skal lægge response herind
}

type Bid struct {
	ClientId int32
	Bid      int32
}

func main() {
	var port = ":8080" //replace with some docker stuff
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %v", err)
	}
	grpcServer := grpc.NewServer()
	server := Server{}
	//set up the server
	server.setupServer()

	//start listening
	auctionator.RegisterAuctionatorServer(grpcServer, &server)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}

}

func (s *Server) notifyBackups() {
	//loop over all my backups and update their maps to mine
	for i := range s.backups {
		//no need to clear out the map first, as it will only hold stuff that the primary also holds
		for k, v := range s.responses {
			s.backups[i].responses[k] = v
		}
	}
}

func (s *Server) setupServer() {
	s.backups = make(map[int32]Server)
	//set the servers id from environment variable
	id, _ := strconv.Atoi(os.Getenv("ID"))
	s.Id = int32(id)
	//tell the server if it is primary replica manager
	isPrimary := os.Getenv("ISPRIMARY")
	if isPrimary != "FALSE" {
		s.isPrimary = true
	} else {
		primary := os.Getenv("PRIMARY")
		s.primary = nil //nil er pladsholder, vi skal sætte primary på ene eller anden måde :) måske kan den få en adresse/port ind denne vej
	}
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

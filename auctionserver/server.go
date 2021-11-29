package main

import (
	"context"
	"fmt"
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
	clients    map[int32]bool //the registered clients - only known by id (bad?) -> can be something else than a map?
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
	var port = os.Getenv("PORT")
	log.Println(port)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %v", listen.Addr())
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

	if server.Id == 1 {
		os.Exit(0)
	}

}

func (s *Server) setupServer() {
	//set the servers id from environment variable
	id, _ := strconv.Atoi(os.Getenv("ID"))
	s.Id = int32(id)
	s.responses = make(map[int32]string)
	s.clients = make(map[int32]bool)
}

func (s *Server) HelloWorld(context.Context, *a.Empty) (*a.Empty, error) {
	log.Printf("Helloworld kaldt på server %v", s.Id)

	return &a.Empty{}, nil
}

func (s *Server) Bid(ctx context.Context, amount *a.Amount) (*a.Acknowledgement, error) {
	var client = amount.ClientId
	_, ok := s.clients[client]
	if !ok {
		s.clients[client] = true //stream or port/ip instead? maybe reference to a client-struct? find inspiration somewhere
	}

	var ack = &a.Acknowledgement{}
	if amount.Amount > s.highestBid.Bid {
		ack.Ack = "Success"
	} else {
		ack.Ack = "Fail"
	}
	//what about exception case? (as noted in the requirements)

	//store the response
	s.responses[amount.Uid] = ack.Ack

	return ack, nil
}

func (s *Server) Result(ctx context.Context, uid *a.Uid) (*a.Outcome, error) {
	var outcome = &a.Outcome{}
	if s.isOver {
		outcome.Over = true
		outcome.Result = s.result
	} else {
		outcome.Over = false
		outcome.Result = s.highestBid.Bid
	}

	//store the response
	s.responses[uid.Uid] = fmt.Sprintf("%v %v", outcome.Over, outcome.Result)

	return outcome, nil
}

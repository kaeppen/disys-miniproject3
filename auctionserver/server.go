package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/kaeppen/disys-miniproject3/auctionator"
	a "github.com/kaeppen/disys-miniproject3/auctionator"
	"google.golang.org/grpc"
)

type Server struct {
	a.UnimplementedAuctionatorServer
	Id         int32
	clients    map[int32]bool //the registered clients
	highestBid Bid
	isOver     bool
	result     int32
	responses  map[string]string
}

type Bid struct {
	ClientId int32
	Amount   int32
}

func main() {
	var port = os.Getenv("PORT")
	log.Printf("port: %v", port)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %v", listen.Addr())
	}
	grpcServer := grpc.NewServer()
	server := Server{}
	//set up the server
	server.setupServer()

	if server.Id == 1 {
		go server.Kill()
	}

	go server.RunAuction()
	//start listening
	auctionator.RegisterAuctionatorServer(grpcServer, &server)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func (s *Server) Kill() {
	var random = rand.Intn(8) + 2
	time.Sleep(time.Duration(random) * time.Second)
	log.Printf("Server %v shutting down", s.Id)
	os.Exit(0)
}

func (s *Server) RunAuction() {
	time.Sleep(30 * time.Second)
	s.isOver = true
	log.Printf("The auction is over!")
	log.Printf("--------------------")
	log.Printf("The winner is client: %v with a bid of %v", s.highestBid.ClientId, s.highestBid.Amount)
	log.Printf("--------------------")
	log.Printf("Shutting down")
	time.Sleep(5 * time.Second)
	os.Exit(0)
}

func (s *Server) setupServer() {
	//set the servers id from environment variable
	id, _ := strconv.Atoi(os.Getenv("ID"))
	s.Id = int32(id)
	s.responses = make(map[string]string)
	s.clients = make(map[int32]bool)
}

func (s *Server) Bid(ctx context.Context, amount *a.Amount) (*a.Acknowledgement, error) {
	var client = amount.ClientId
	_, ok := s.clients[client]
	if !ok {
		s.clients[client] = true
	}

	var ack = &a.Acknowledgement{}
	if amount.Amount > s.highestBid.Amount && !s.isOver {
		ack.Ack = "Success"
		s.highestBid.Amount = amount.Amount
		s.highestBid.ClientId = amount.ClientId
	} else {
		ack.Ack = "Fail"
	}
	//what about exception case? (as noted in the requirements) -> doesn't matter according to TA's.

	//store the response
	s.responses[amount.Timestamp] = ack.Ack

	return ack, nil
}

func (s *Server) Result(ctx context.Context, timestamp *a.Timestamp) (*a.Outcome, error) {
	var outcome = &a.Outcome{}
	if s.isOver {
		outcome.Over = true
		outcome.Result = s.result
	} else {
		outcome.Over = false
		outcome.Result = s.highestBid.Amount
	}

	//store the response
	s.responses[timestamp.Timestamp] = fmt.Sprintf("%v %v", outcome.Over, outcome.Result)

	return outcome, nil
}

package main

import (
	"context"
	"fmt"
	"io"
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
	isPrimary  bool                                                    //am i primary
	primary    a.AuctionatorServer                                     //who is my primary
	backups    map[int32]a.Auctionator_EstablishBackupConnectionServer //connection to backups
	clients    map[int32]bool                                          //the registered clients - only known by id (bad?) -> can be something else than a map?
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

}

// func (s *Server) notifyBackups() {
// 	//loop over all my backups and update their maps to mine
// 	for i := range s.backups {
// 		//no need to clear out the map first, as it will only hold stuff that the primary also holds
// 		for k, v := range s.responses {
// 			s.backups[i].responses[k] = v
// 		}
// 	}
// }

func (s *Server) setupServer() {
	s.backups = make(map[int32]a.Auctionator_EstablishBackupConnectionServer)
	//set the servers id from environment variable
	id, _ := strconv.Atoi(os.Getenv("ID"))
	s.Id = int32(id)
	//tell the server if it is primary replica manager
	isPrimary := os.Getenv("ISPRIMARY")
	if isPrimary != "FALSE" {
		s.isPrimary = true
	} else {
		primary := os.Getenv("PRIMARY")
		s.connectToPrimary(primary) //der skal måske laves lidt formattering her
		//s.primary = nil //nil er pladsholder, vi skal sætte primary på ene eller anden måde :) måske kan den få en adresse/port ind denne vej
	}
}

func (s *Server) connectToPrimary(port string) {
	conn, err := grpc.Dial(port, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect %s", err)
	}
	ctx := context.Background()
	defer conn.Close()
	primary := a.NewAuctionatorClient(conn)
	setup := &a.ConnectionSetup{Port: "skal fjernes, da det jo ikke er nødvendigt", Id: s.Id}
	stream, err := primary.EstablishBackupConnection(ctx, setup)
	if err != nil {
		log.Fatalf("Error while connecting to primray %v", err)
	}
	//listen on the stream, might have to be changed
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalf("Backup with id %v failed to recieve message", s.Id)
			}
			s.updateResponses(in)
		}
	}()
}

func (s *Server) EstablishBackupConnection(setup *a.ConnectionSetup, stream a.Auctionator_EstablishBackupConnectionServer) error {
	//store the stream to the backup
	id := setup.Id
	s.backups[id] = stream

	//keep the stream alive
	for {
		select {
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *Server) updateResponses(response *a.UpdateResponse) {
	s.responses[response.Key] = response.Value
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

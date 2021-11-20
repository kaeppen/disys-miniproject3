package main

import (
	"fmt"
	"log"

	"github.com/kaeppen/disys-miniproject3/auctionator"
	"google.golang.org/grpc"
)

func main() {
	var port = "1337" //replace with somer docker stuff
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Client %v could not connect: %s", "klientnavn", err)
	}
	defer conn.Close()

	client := auctionator.NewAuctionatorClient(conn)
	fmt.Printf("client: %v\n", client) // placeholder line
}

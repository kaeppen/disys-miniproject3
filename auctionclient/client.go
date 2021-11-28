package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	a "github.com/kaeppen/disys-miniproject3/auctionator"
	"google.golang.org/grpc"
)

type Frontend struct {
	//det er mig der er grpc klient
	connection a.AuctionatorClient
	ctx        context.Context
	uid        int32 //unique identifier, needs to be incorporated in requests!
	servers    map[int]a.AuctionatorClient
}

type Client struct {
	Id    int32
	front Frontend
}

func main() {
	c := Client{}
	id, _ := strconv.Atoi(os.Getenv("ID"))
	c.Id = int32(id)
	c.setupFrontend()
	log.Print("Client has managed to set up frontend, nice")

}

//overvej om der skal returværdi på denne?
func (c *Client) Bid(amount int32) {
	input := &a.Amount{Amount: amount, ClientId: c.Id}
	ack, err := c.front.connection.Bid(c.front.ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Client %v got response %v", c.Id, ack.Ack)
}

func (c *Client) Result() {
	outcome, err := c.front.connection.Result(c.front.ctx, &a.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Client %v: auction is over: %v and the highest bidder/result is: %v", c.Id, outcome.Over, outcome.Result)
}

func (c *Client) setupFrontend() {
	num, _ := strconv.Atoi(os.Getenv("NUMSERVERS"))
	var numservers = int(num)
	c.front.servers = make(map[int]a.AuctionatorClient)
	for i := 0; i < numservers; i++ {
		var conn *grpc.ClientConn
		var port = os.Getenv(fmt.Sprintf("SERVER%v", i))
		log.Printf("Trying to connect to server on port %v", port)
		conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		client := a.NewAuctionatorClient(conn)
		c.front.servers[i] = client
	}
	ctx := context.Background()
	c.front.ctx = ctx
}

package main

import (
	"context"
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
	var conn *grpc.ClientConn
	log.Print("Trying to connect to server")
	var port = os.Getenv("PORT")
	conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	defer conn.Close()
	ctx := context.Background()
	client := a.NewAuctionatorClient(conn)
	c.front.connection = client
	c.front.ctx = ctx
}

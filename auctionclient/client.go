package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	a "github.com/kaeppen/disys-miniproject3/auctionator"
	"google.golang.org/grpc"
)

type Frontend struct {
	//det er mig der er grpc klient
	ctx     context.Context
	uid     int32 //unique identifier, needs to be incorporated in requests!
	servers map[int]a.AuctionatorClient
}

type Client struct {
	Id    int32
	front Frontend
}

func main() {
	c := Client{}
	id, _ := strconv.Atoi(os.Getenv("ID"))
	c.Id = int32(id)
	time.Sleep(5 * time.Second)
	c.setupFrontend()
	log.Print("Client has managed to set up frontend, nice")
	time.Sleep(5 * time.Second)
	c.Demo()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		for scanner.Scan() {
			if scanner.Text() == "Bid" {
				bidAmount, _ := strconv.Atoi(scanner.Text())
				c.Bid(int32(bidAmount))

			} else if scanner.Text() == "Result" {
				c.Result()
			}
		}
	}
}

func (c *Client) Demo() {
	if c.Id == 1 {
		c.Bid(500)
		time.Sleep(2 * time.Second)
		c.Result()
		time.Sleep(2 * time.Second)
		c.Bid(250)
		time.Sleep(2 * time.Second)
		c.Bid(900)
		c.Result()
	} else {
		c.Result()
		c.Bid(100)
		time.Sleep(2 * time.Second)
		time.Sleep(2 * time.Second)
		c.Bid(1500)
		time.Sleep(2 * time.Second)
	}
}

func (c *Client) Bid(amount int32) {
	c.front.uid++ //update the unique identifier
	for i := range c.front.servers {
		input := &a.Amount{Amount: amount, ClientId: c.Id, Uid: c.front.uid}
		ack, err := c.front.servers[i].Bid(c.front.ctx, input)
		if err != nil {
			//log.Print(err)
			log.Printf("Error when attempting to reach server %v - removing", i)
			delete(c.front.servers, i) //delete the server fom the map
		} else {
			log.Printf("Client %v got response %v from server %v", c.Id, ack.Ack, i)
		}
	}
}

func (c *Client) Result() {
	c.front.uid++ //update the unique identifier
	for i := range c.front.servers {
		outcome, err := c.front.servers[i].Result(c.front.ctx, &a.Uid{Uid: c.front.uid})
		if err != nil {
			//log.Print(err)
			log.Printf("Error when attempting to reach server %v - removing", i)
			delete(c.front.servers, i) //delete the server fom the map
		} else {
			log.Printf("Client %v: auction is over: %v and the highest bidder/result is: %v", c.Id, outcome.Over, outcome.Result)
		}
	}
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
		client := a.NewAuctionatorClient(conn)
		c.front.servers[i] = client
	}
	ctx := context.Background()
	c.front.ctx = ctx
}

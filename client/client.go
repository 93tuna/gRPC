package main

import (
	"bufio"
	"context"
	"fmt"
	"grpc-chat/chatserver"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	address = "49.247.192.42:5000"
)

// type wrappedStream struct {
// 	grpc.ClientStream
// }

// func (w *wrappedStream) RecvMsg(m interface{}) error {
// 	log.Printf("====== [Client Stream Interceptor] Receive a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))

// 	return w.ClientStream.RecvMsg(m)
// }

// func (w *wrappedStream) SendMsg(m interface{}) error {
// 	log.Printf("====== [Client Stream Interceptor] Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
// 	return w.ClientStream.SendMsg(m)
// }

// func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
// 	return &wrappedStream{s}
// }

// func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

// 	log.Println("======= [Client Interceptor] ", method)
// 	s, err := streamer(ctx, desc, cc, method, opts...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return newWrappedStream(s), nil
// }

func main() {

	//connect to grpc server
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Faile to conncet to gRPC server :: %v", err)
	}
	defer conn.Close()

	reader1 := bufio.NewReader(os.Stdin)
	fmt.Printf("PNumber : ")
	pNumber, err := reader1.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}

	pNumber = strings.Trim(pNumber, "\r\n")

	reader2 := bufio.NewReader(os.Stdin)
	fmt.Printf("Room : ")
	room, err := reader2.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}

	room = strings.Trim(room, "\r\n")

	md := metadata.Pairs(
		"pnumber", pNumber,
		"room", room,
	)

	mdCtx := metadata.NewOutgoingContext(context.Background(), md)

	//call ChatService to create a stream
	client := chatserver.NewServicesClient(conn)

	stream, err := client.ChatService(mdCtx)
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}

	// implement communication with gRPC server
	ch := clienthandle{stream: stream, pNumber: pNumber, room: room}
	go ch.sendMessage()
	go ch.receiveMessage()

	//blocker
	bl := make(chan bool)
	<-bl

}

//clienthandle
type clienthandle struct {
	stream  chatserver.Services_ChatServiceClient
	pNumber string
	room    string
}

//send message
func (ch *clienthandle) sendMessage() {

	// create a loop
	for {

		reader := bufio.NewReader(os.Stdin)
		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf(" Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")

		clientMessageBox := &chatserver.FromClient{
			Pnumber: ch.pNumber,
			Body:    clientMessage,
			Room:    ch.room,
		}

		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}

	}

}

//receive message
func (ch *clienthandle) receiveMessage() {

	//create a loop
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in receiving message from server :: %v", err)
		}

		//print message to console
		fmt.Printf("%s : %s \n", mssg.Pnumber, mssg.Body)
	}
}

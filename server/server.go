package main

import (
	"log"
	"net"
	"os"
	"time"

	"grpc-chat/chatserver"

	"google.golang.org/grpc"
)

type wrappedStream struct {
	grpc.ServerStream
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Wrapper] Receive a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Wrapper] Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

// // Server :: Unary Interceptor
// func orderUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 	// Pre-processing logic
// 	// Gets info about the current RPC call by examining the args passed in
// 	log.Println("======= [Server Interceptor] ", info.FullMethod)
// 	log.Printf(" Pre Proc Message : %s", req)

// 	// Invoking the handler to complete the normal execution of a unary RPC.
// 	m, err := handler(ctx, req)

// 	// Post processing logic
// 	log.Printf(" Post Proc Message : %s", m)
// 	return m, err
// }

func chatServerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Pre-processing
	log.Println("====== [Server Stream Interceptor] ", info.FullMethod)

	// Invoking the StreamHandler to complete the execution of RPC invocation
	err := handler(srv, newWrappedStream(ss))
	if err != nil {
		log.Printf("RPC failed with error %v", err)
	}
	return err
}

func main() {

	//assign port
	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "5000" //default Port set to 5000 if PORT is not set in env
	}

	//init listener
	listen, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", Port, err)
	}
	log.Println("Listening @ : " + Port)

	//gRPC server instance
	// grpcserver := grpc.NewServer(
	// 	grpc.StreamInterceptor(chatServerStreamInterceptor))
	grpcserver := grpc.NewServer()

	//register ChatService
	s := chatserver.ChatServer{}
	chatserver.RegisterServicesServer(grpcserver, &s)

	//grpc listen and serve
	err = grpcserver.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}

}

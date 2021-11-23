package echo

import (
	"context"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Server struct{}

func (s Server) Echo(ctx context.Context, request *EchoRequest) (*EchoResponse, error) {
	log.Printf("Handling Echo request [%v] with context %v", request, ctx)
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Unable to get hostname %v", err)
		hostname = ""
	}
	grpc.SendHeader(ctx, metadata.Pairs("hostname", hostname))
	return &EchoResponse{Content: request.Content}, nil
}

func NewServer() EchoServer {
	return &Server{}
}

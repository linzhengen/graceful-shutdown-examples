//go:generate protoc --go_out=. --go_opt paths=source_relative --go-grpc_out=. --go-grpc_opt paths=source_relative --go-grpc_opt require_unimplemented_servers=false ./echo/echo.proto

package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/linzhengen/graceful-shutdown-examples/go/grpc/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	echo.RegisterEchoServer(grpcServer, echo.NewServer())
	reflection.Register(grpcServer)
	log.Printf("Listening for Echo on port %s", port)
	healthServer.SetServingStatus("grpc.health.v1.echo", healthpb.HealthCheckResponse_SERVING)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully")

	healthServer.SetServingStatus("grpc.health.v1.echo", healthpb.HealthCheckResponse_NOT_SERVING)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		// close listeners to stop accepting new connections,
		// will block on any existing transports
		grpcServer.GracefulStop()
	}()
	select {
	case <-ch:
		log.Printf("Graceful stopped")
	case <-time.After(5 * time.Second):
		// took too long, manually close open transports
		// e.g. watch streams
		log.Printf("Graceful stop timeout, force stop!!")
		grpcServer.Stop()
		<-ch
	}
	log.Println("Server successfully stopped")
}

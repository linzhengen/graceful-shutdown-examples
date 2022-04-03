package main

import (
	"context"
	"log"
	"os/signal"
	"sync/atomic"
	"syscall"

	"cloud.google.com/go/pubsub"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	projectID := "hoge_project_id"
	subID := "hoge_sub_id"
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	// Turn on synchronous mode. This makes the subscriber use the Pull RPC rather
	// than the StreamingPull RPC, which is useful for guaranteeing MaxOutstandingMessages,
	// the max number of messages the client will hold in memory at a time.
	sub.ReceiveSettings.Synchronous = true
	sub.ReceiveSettings.MaxOutstandingMessages = 10

	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		atomic.AddInt32(&received, 1)
		msg.Ack()
	})
	if err != nil {
		log.Printf("sub.Receive: %v", err)
	}
	log.Printf("Received %d messages\n", received)

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("successfully stopped")
}

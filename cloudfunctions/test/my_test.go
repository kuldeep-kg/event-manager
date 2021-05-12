package test

import (
	"context"
	"log"
	"testing"

	"github.com/maira-io/event-manager/cloudfunctions"
	"github.com/stretchr/testify/assert"
)

func TestWatermill(t *testing.T) {
	router := cloudfunctions.CreateWatermillRouter()

	subscriber := cloudfunctions.CreateWatermillSubscriber()

	publisher := cloudfunctions.CreateWatermillPublisher()

	cloudfunctions.SetHandlers(router, subscriber, publisher)
	log.Println("Created handler")

	// Now that all handlers are registered, we're running the Router.
	// Run is blocking while the router is running.
	ctx := context.Background()
	if err := router.Run(ctx); err != nil {
		panic(err)
		t.Fatal(err)
	}

	msg := "Hello! This is test message."
	log.Printf("sending message %s\n", msg)

	cloudfunctions.PublishMessages(publisher, []byte(msg))
	assert.NotNil(t, router)

}

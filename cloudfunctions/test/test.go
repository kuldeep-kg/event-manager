package main

import (
	"context"

	"github.com/maira-io/event-manager/cloudfunctions"
)

func main() {
	router := cloudfunctions.CreateWatermillRouter()

	subscriber := cloudfunctions.CreateWatermillSubscriber()

	publisher := cloudfunctions.CreateWatermillPublisher()

	cloudfunctions.SetHandlers(router, subscriber, publisher)

	// Now that all handlers are registered, we're running the Router.
	// Run is blocking while the router is running.
	ctx := context.Background()
	if err := router.Run(ctx); err != nil {
		panic(err)
	}

	msg := "Hello! This is test message."

	cloudfunctions.PublishMessages(publisher, []byte(msg))

}

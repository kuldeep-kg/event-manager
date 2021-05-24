package test

import (
	"log"
	"testing"

	"github.com/maira-io/event-manager/cloudfunctions"
	"github.com/maira-io/event-manager/cloudfunctions/watermillpubsub"
)

func TestWatermill(t *testing.T) {

	publisher := watermillpubsub.CreateWatermillPublisher()

	log.Println("Created handler")
	msg := "Hello! This is test message."
	log.Printf("sending message %s\n", msg)

	cloudfunctions.PublishMessages(publisher, []byte(msg))

	cloudfunctions.PublishMessages(publisher, []byte(msg))

	cloudfunctions.PublishMessages(publisher, []byte(msg))

	log.Println("router started..")

}

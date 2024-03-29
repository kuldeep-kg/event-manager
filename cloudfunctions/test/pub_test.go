package test

import (
	"log"
	"testing"

	"github.com/kuldeep-kg/event-manager/cloudfunctions"
	"github.com/kuldeep-kg/event-manager/watermillpubsub"
)

func TestWatermill(t *testing.T) {

	publisher := watermillpubsub.CreateWatermillPublisher()

	log.Println("Created handler")
	msg := "Hello! This is test message."
	log.Printf("sending message %s\n", msg)

	for {
		cloudfunctions.PublishMessages(publisher, []byte(msg))
	}

	log.Println("router started..")

}

package cloudfunctions

import (
	"context"
	"fmt"
	"log"
	"time"

	"encoding/json"

	"net"
	"net/http"
	"net/url"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"

	event_pb "github.com/maira-io/apiserver/generated/event"
	"github.com/prometheus/alertmanager/template"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PrometheusWebhookWithWatermill(w http.ResponseWriter, r *http.Request) {
	logger := watermill.NewStdLogger(false, false)
	subscriber, err := googlecloud.NewSubscriber(
		googlecloud.SubscriberConfig{
			// custom function to generate Subscription Name,
			// there are also predefined TopicSubscriptionName and TopicSubscriptionNameWithSuffix available.
			GenerateSubscriptionName: func(topic string) string {
				return "maira-sub_" + topic
			},
			ProjectID: "maira-event-manager",
		},
		logger,
	)
	if err != nil {
		panic(err)
	}

	// Subscribe will create the subscription. Only messages that are sent after the subscription is created may be received.
	messages, err := subscriber.Subscribe(context.Background(), "event.topic")
	if err != nil {
		panic(err)
	}

	go process(messages)

	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
		ProjectID: "maira-event-manager",
	}, logger)
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()

	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	for _, alert := range data.Alerts {
		event := &event_pb.Event{
			Tenant:      "customer1",
			Namespace:   "default",
			EventType:   alert.Labels["alertname"],
			Site:        getHost(data.ExternalURL),
			Timestamp:   timestamppb.New(alert.StartsAt),
			Message:     alert.Annotations["description"],
			Severity:    alert.Labels["severity"],
			Labels:      alert.Labels,
			Annotations: alert.Annotations,
		}
		log.Println(event)
		payload, _ := json.Marshal(event)

		publishMessages(publisher, payload)
	}

	fmt.Fprint(w, "successfully processed")
}

func getHost(externalurl string) string {
	u, err := url.Parse(externalurl)
	if err != nil {
		panic(err)
	}
	host, _, _ := net.SplitHostPort(u.Host)
	return host
}

func publishMessages(publisher message.Publisher, payload message.Payload) {
	msg := message.NewMessage(watermill.NewUUID(), payload)

	if err := publisher.Publish("event.topic", msg); err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
}

func process(messages <-chan *message.Message) {
	for msg := range messages {
		log.Printf("received message: %s, payload: %s", msg.UUID, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}

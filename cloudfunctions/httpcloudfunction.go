package cloudfunctions

import (
	"fmt"
	"log"

	"encoding/json"

	"net"
	"net/http"
	"net/url"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	event_pb "github.com/maira-io/apiserver/generated/event"
	"github.com/prometheus/alertmanager/template"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PrometheusWebhookWithWatermillRouter(w http.ResponseWriter, r *http.Request) {

	// router := CreateWatermillRouter()

	// subscriber := CreateWatermillSubscriber()

	publisher := CreateWatermillPublisher()

	// SetHandlers(router, subscriber, publisher)

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

		PublishMessages(publisher, payload)
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

func PublishMessages(publisher message.Publisher, payload message.Payload) {
	msg := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(watermill.NewUUID(), msg)
	log.Printf("sending message %s, correlation id: %s\n", msg.UUID, middleware.MessageCorrelationID(msg))
	if err := publisher.Publish("maira.event", msg); err != nil {
		panic(err)
	}
}

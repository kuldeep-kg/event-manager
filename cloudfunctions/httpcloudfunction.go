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
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"

	event_pb "github.com/maira-io/apiserver/generated/event"
	"github.com/prometheus/alertmanager/template"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// For this example, we're using just a simple logger implementation,
	// You probably want to ship your own implementation of `watermill.LoggerAdapter`.
	logger = watermill.NewStdLogger(false, false)
)

func PrometheusWebhookWithWatermillRouter(w http.ResponseWriter, r *http.Request) {

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	// SignalsHandler will gracefully shutdown Router when SIGTERM is received.
	// You can also close the router by just calling `r.Close()`.
	router.AddPlugin(plugin.SignalsHandler)

	// Router level middleware are executed for every message sent to the router
	router.AddMiddleware(
		// CorrelationID will copy the correlation id from the incoming message's metadata to the produced messages
		middleware.CorrelationID,

		// The handler function is retried if it returns an error.
		// After MaxRetries, the message is Nacked and it's up to the PubSub to resend it.
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			Logger:          logger,
		}.Middleware,

		// Recoverer handles panics from handlers.
		// In this case, it passes them as errors to the Retry middleware.
		middleware.Recoverer,
	)

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

	/*	// Subscribe will create the subscription. Only messages that are sent after the subscription is created may be received.
		messages, err := subscriber.Subscribe(context.Background(), "event.topic")
		if err != nil {
			panic(err)
		}

		go process(messages)
	*/
	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
		ProjectID: "maira-event-manager",
	}, logger)
	if err != nil {
		panic(err)
	}

	// AddHandler returns a handler which can be used to add handler level middleware
	handler := router.AddHandler(
		"struct_handler",       // handler name, must be unique
		"incoming.event.topic", // topic from which we will read events
		subscriber,
		"outgoing.event.topic", // topic to which we will publish events
		publisher,
		structHandler{}.Handler,
	)

	// Handler level middleware is only executed for a specific handler
	// Such middleware can be added the same way the router level ones
	handler.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(message *message.Message) ([]*message.Message, error) {
			log.Println("executing handler specific middleware for ", message.UUID)

			return h(message)
		}
	})

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

	// just for debug, we are printing all messages received on `incoming.event.topic`
	router.AddNoPublisherHandler(
		"print_incoming_messages",
		"incoming.event.topic",
		subscriber,
		printMessages,
	)

	// just for debug, we are printing all events sent to `outgoing.event.topic`
	router.AddNoPublisherHandler(
		"print_outgoing_messages",
		"outgoing.event.topic",
		subscriber,
		printMessages,
	)

	// Now that all handlers are registered, we're running the Router.
	// Run is blocking while the router is running.
	ctx := context.Background()
	if err := router.Run(ctx); err != nil {
		panic(err)
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
	middleware.SetCorrelationID(watermill.NewUUID(), msg)
	log.Printf("sending message %s, correlation id: %s\n", msg.UUID, middleware.MessageCorrelationID(msg))
	if err := publisher.Publish("incoming.event.topic", msg); err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
}

/*func process(messages <-chan *message.Message) {
	for msg := range messages {
		log.Printf("received message: %s, payload: %s", msg.UUID, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}*/

func printMessages(msg *message.Message) error {
	log.Printf(
		"\n> Received message: %s\n> %s\n> metadata: %v\n\n",
		msg.UUID, string(msg.Payload), msg.Metadata,
	)
	return nil
}

type structHandler struct {
	// we can add some dependencies here
}

func (s structHandler) Handler(msg *message.Message) ([]*message.Message, error) {
	log.Println("structHandler received message", msg.UUID)

	msg = message.NewMessage(watermill.NewUUID(), []byte("message produced by structHandler"))
	return message.Messages{msg}, nil
}

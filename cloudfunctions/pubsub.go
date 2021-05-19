package cloudfunctions

import (
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
)

var (
	// For this example, we're using just a simple logger implementation,
	// You probably want to ship your own implementation of `watermill.LoggerAdapter`.
	logger = watermill.NewStdLogger(false, false)
)

func CreateWatermillRouter() *message.Router {
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

	return router
}

func CreateWatermillSubscriber() *googlecloud.Subscriber {
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
	return subscriber
}

func CreateWatermillPublisher() *googlecloud.Publisher {
	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
		ProjectID: "maira-event-manager",
	}, logger)
	if err != nil {
		panic(err)
	}
	return publisher
}

func SetHandlers(router *message.Router, subscriber *googlecloud.Subscriber, publisher *googlecloud.Publisher) {
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

	// // just for debug, we are printing all messages received on `incoming.event.topic`
	// router.AddNoPublisherHandler(
	// 	"print_incoming_messages",
	// 	"incoming.event.topic",
	// 	subscriber,
	// 	printMessages,
	// )

	// just for debug, we are printing all events sent to `outgoing.event.topic`
	router.AddNoPublisherHandler(
		"print_outgoing_messages",
		"outgoing.event.topic",
		subscriber,
		printMessages,
	)
}

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

	printMessages(msg)

	msg = message.NewMessage(watermill.NewUUID(), []byte("message produced by structHandler"))
	return message.Messages{msg}, nil
}

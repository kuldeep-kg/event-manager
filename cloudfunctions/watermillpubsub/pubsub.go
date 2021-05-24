package cloudfunctions

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
)

var (
	// For this example, we're using just a simple logger implementation,
	// You probably want to ship your own implementation of `watermill.LoggerAdapter`.
	logger = watermill.NewStdLogger(false, false)
)

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

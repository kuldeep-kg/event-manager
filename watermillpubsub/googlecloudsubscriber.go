// Sources for https://watermill.io/docs/getting-started/
package watermillpubsub

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GoRoutineThreads int `yaml:"go-routine-threads"`
}

func main() {
	maxNumberOfRoutines := 0

	//Read go-routine-threads configuration from file
	f, err := os.Open("../config.yml")
	if err != nil {
		log.Println("Error while opening the configuration file: " + err.Error())
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Println("Error while decoding the file: " + err.Error())
	}

	log.Printf("Fetched number of go routine threads from config.yml : %v\n", cfg.GoRoutineThreads)

	if cfg.GoRoutineThreads != 0 {
		maxNumberOfRoutines = cfg.GoRoutineThreads
	}

	//Read go-routine-threads configuration from command line arguments
	args := os.Args[1:]

	if len(args) > 0 {
		maxGoProcesses := args[0]

		i, err := strconv.Atoi(maxGoProcesses)

		if err == nil {
			log.Printf("Number of go routine threads from command line : %v\n", i)
			maxNumberOfRoutines = i
		}
	}

	//Set default number (50) of go routines threads if not provided
	if maxNumberOfRoutines <= 0 {
		maxNumberOfRoutines = 50
	}

	log.Printf("Thread : %v\n", maxNumberOfRoutines)
	subscriber := CreateWatermillSubscriber()

	// Subscribe will create the subscription. Only messages that are sent after the subscription is created may be received.
	messages, err := subscriber.Subscribe(context.Background(), "maira.event")
	if err != nil {
		panic(err)
	}

	process(messages, maxNumberOfRoutines)
}

func process(messages <-chan *message.Message, maxNumberOfRoutines int) {
	var wg sync.WaitGroup
	// Create bufferized channel with size 5
	goroutines := make(chan struct{}, maxNumberOfRoutines)
	for msg := range messages {
		log.Printf("received message: %s, payload: %s", msg.UUID, string(msg.Payload))

		// 1 struct{}{} - 1 goroutine
		goroutines <- struct{}{}
		wg.Add(1)
		go ProcessEvent(msg, goroutines, &wg)
	}
	wg.Wait()
	close(goroutines)
}

func ProcessEvent(msg *message.Message, goroutines <-chan struct{}, wg *sync.WaitGroup) {
	log.Println("Starting Processing the event further to ingest the data into the database")

	// write code here - replace/remove the sleep command
	time.Sleep(time.Second * 300)

	log.Println("Finished Processing the event further to ingest the data into the database")

	// we need to Acknowledge that we received and processed the message,
	// otherwise, it will be resent over and over again.
	msg.Ack()

	<-goroutines
	wg.Done()
}

package events

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventTypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"

	"context"
	"encoding/json"
	"log"
)

// Function to that read items from channel and send them to EventBridge bus
func SendEvents(eventsClient *eventbridge.Client, eventBusName string, ch chan map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()

	for item := range ch {
		item["status"] = "row_sent" // need to add this status for EventBridge rule
		details, err := json.Marshal(item)
		if err != nil {
			log.Fatalln("Could not serialize row. Error - ", err)
		}
		_, err = eventsClient.PutEvents(context.TODO(), &eventbridge.PutEventsInput{
			Entries: []eventTypes.PutEventsRequestEntry{{
				EventBusName: aws.String(eventBusName),
				Source:       aws.String("app.container"),
				DetailType:   aws.String("extraction-process"),
				Detail:       aws.String(string(details)),
			}},
		})
		if err != nil {
			log.Fatalln("Got error calling PutEvents - ", err)
		}
	}
}

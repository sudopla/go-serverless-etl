// Lambda functions that gets trigger with EventBridge rule
// It applies some changes on the item and sends it back to the bus

package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventTypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

type BusEvent struct {
	Version    string          `json:"version"`
	ID         string          `json:"id"`
	DetailType string          `json:"detail-type"`
	Source     string          `json:"source"`
	AccountID  string          `json:"account"`
	Time       time.Time       `json:"time"`
	Region     string          `json:"region"`
	Resources  []string        `json:"resources"`
	Detail     json.RawMessage `json:"detail"`
}

type OriginalItem struct {
	Street    string  `json:"street"`
	City      string  `json:"city"`
	Zip       string  `json:"zip"`
	State     string  `json:"state"`
	Beds      int     `json:"beds,string"`
	Baths     int     `json:"baths,string"`
	Sq_ft     int     `json:"sq__ft,string"`
	Type      string  `json:"type"`
	SalesDate string  `json:"sale_date"`
	Price     int     `json:"price,string"`
	Latitude  float64 `json:"latitude,string"`
	Longitude float64 `json:"longitude,string"`
}

type TransformedItem struct {
	OriginalItem
	Status       string `json:"status"`
	PricePerSqFt int    `json:"price_per_sq_ft"`
}

var (
	eventBusName string
	eventsClient *eventbridge.Client
)

func init() {
	log.Println("Initializing Lambda execution environment")
	log.Println("Initializing EventBridge client ...")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	eventsClient = eventbridge.NewFromConfig(cfg)

	log.Print("Getting environment variables ...")
	eventBusName = os.Getenv("EVENTBRIDGE_BUS_NAME")
}

func handler(ctx context.Context, event BusEvent) {
	// Apply transformations to item
	var item = OriginalItem{}
	if err := json.Unmarshal(event.Detail, &item); err != nil {
		log.Fatalf("Could not unmarshall item - %s\n Error - ", string(event.Detail), err)
	}

	log.Println("Parse date ...")
	dLayout := "Mon Jan 2 15:04:05 MST 2006"
	dParsed, err := time.Parse(dLayout, item.SalesDate)
	if err != nil {
		log.Fatalln("Could not parse date. Error - ", err)
	}
	item.SalesDate = dParsed.Format("2006-01-02")

	log.Println("Add price per square foot")
	pSqFt := int(math.Ceil(float64(item.Price) / float64(item.Sq_ft)))

	transformedItem := TransformedItem{
		OriginalItem: item,
		Status:       "item_transformed",
		PricePerSqFt: pSqFt,
	}
	details, err := json.Marshal(transformedItem)
	if err != nil {
		log.Fatalln("Could not serialize item. Error - ", err.Error())
	}

	// Send transformed item to EventBridge
	log.Println("Send tranformed item to EventBridge")
	_, err = eventsClient.PutEvents(context.TODO(), &eventbridge.PutEventsInput{
		Entries: []eventTypes.PutEventsRequestEntry{{
			EventBusName: aws.String(eventBusName),
			Source:       aws.String("app.transform"),
			DetailType:   aws.String("transform-process"),
			Detail:       aws.String(string(details)),
		}},
	})
	if err != nil {
		log.Fatalf("Got error calling PutEvents %s", err)
	}
}

func main() {
	lambda.Start(handler)
}

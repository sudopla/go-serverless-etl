// Lambda functions that gets trigger with EventBridge rule
// It stores item in Dyanamo table

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
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

type DynamoItem struct {
	Id           string  `json:"id"`
	Street       string  `json:"street"`
	City         string  `json:"city"`
	Zip          string  `json:"zip"`
	State        string  `json:"state"`
	Beds         int     `json:"beds,string"`
	Baths        int     `json:"baths,string"`
	Sq_ft        int     `json:"sq__ft,string"`
	Type         string  `json:"type"`
	SalesDate    string  `json:"sale_date"`
	Price        int     `json:"price,string"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Status       string  `json:"status"`
	PricePerSqFt int     `json:"price_per_sq_ft,string"`
}

var (
	tableName    string
	dynamoClient *dynamodb.Client
)

func init() {
	log.Println("Initializing Lambda execution environment")
	log.Println("Initializing Dynamo Client...")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)

	log.Println("Getting environment variables...")
	tableName = os.Getenv("TABLE_NAME")
}

func handler(ctx context.Context, event BusEvent) {
	item := DynamoItem{}
	json.Unmarshal(event.Detail, &item)
	item.Id = uuid.NewString()

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		log.Fatalf("Got error marshalling item: %s", err)
	}

	_, err = dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
	}
}

func main() {
	lambda.Start(handler)
}

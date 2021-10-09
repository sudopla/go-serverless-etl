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
	dynamoTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	Id           string  `json:"id,omitempty"`
	Street       string  `json:"street"`
	City         string  `json:"city"`
	Zip          string  `json:"zip"`
	State        string  `json:"state"`
	Ceds         int     `json:"beds,string"`
	Baths        int     `json:"baths,string"`
	Sq_ft        int     `json:"sq__ft,string"`
	Type         string  `json:"type"`
	SalesDate    string  `json:"sale_date"`
	Price        int     `json:"price,string"`
	Latitude     float64 `json:"latitude,string"`
	Longitude    float64 `json:"longitude,string"`
	Status       string  `json:"status"`
	PricePerSqFt int     `json:"price_per_sq_ft,string"`
}

var (
	tableName    string
	dynamoClient *dynamodb.Client
)

func setEncoderOpts(encoder *attributevalue.EncoderOptions) {
	encoder.TagKey = "json"
}

// Same sdk method but with the encoder options
func marshalMap(in interface{}) (map[string]dynamoTypes.AttributeValue, error) {
	av, err := attributevalue.NewEncoder(setEncoderOpts).Encode(in)

	asMap, ok := av.(*dynamoTypes.AttributeValueMemberM)
	if err != nil || av == nil || !ok {
		return map[string]dynamoTypes.AttributeValue{}, err
	}

	return asMap.Value, nil
}

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
	log.Printf("EventDetail - %+v\n", item)

	av, err := marshalMap(item)
	if err != nil {
		log.Fatalf("Got error marshalling item: %s", err)
	}

	log.Printf("Store item in table - %v", av)
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

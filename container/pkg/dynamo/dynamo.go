package dynamo

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"log"
)

func InstertItem(svc *dynamodb.DynamoDB, ch chan map[string]string, wg *sync.WaitGroup) {

	defer wg.Done()

	for item := range ch {
		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			log.Fatalf("Got error marshalling new movie item: %s", err)
		}

		// Create item in table Movies
		tableName := "Homes"

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String(tableName),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			log.Fatalf("Got error calling PutItem: %s", err)
		}
		// fmt.Println("Added Item!")
	}

}

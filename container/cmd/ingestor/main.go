package main

import (
	"sync"

	"ingestor/internal/events"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	s3Manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"context"
	"encoding/csv"
	"io"
	"log"
	"os"
)

func main() {
	// Environment variables from runTask call
	bucket := os.Getenv("BUCKET")
	key := os.Getenv("KEY")

	// Environment variables from the CDK stack
	eventBusName := os.Getenv("EVENTBRIDGE_BUS_NAME")

	log.Println("Initializing S3 and EventBridge clients")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	s3Downloader := s3Manager.NewDownloader(s3.NewFromConfig(cfg))
	eventsClient := eventbridge.NewFromConfig(cfg)

	log.Printf("Download file %s", key)
	csvFile, err := os.Create("/data_volume/input.csv") // data_volume is the path where the container volume is mounted
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	numBytes, err := s3Downloader.Download(context.TODO(), csvFile, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Downloaded", csvFile.Name(), numBytes, "bytes")

	log.Println("Extract rows and send them to EventBrdige")
	// Read headers first
	reader := csv.NewReader(csvFile)
	headers, err := reader.Read()
	if err != nil {
		log.Fatalln("Couldn't read headers")
	}

	// Create Workers to send rows to EventBridge bus- Don't want to run thousands of concurrent requests
	numWorkes := 100 // This number can be changed. EventBridge putEvents has a soft limit of 10,000 requests per second
	var wg sync.WaitGroup
	ch := make(chan map[string]string)
	for w := 0; w < numWorkes; w++ {
		wg.Add(1)
		go events.SendEvents(eventsClient, eventBusName, ch, &wg)
	}

	// Read rows in file
	counter := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		item := map[string]string{}
		for i, v := range row {
			item[headers[i]] = v
		}

		ch <- item
		log.Println("Item added to channel - ", item)
		counter++
	}

	close(ch)
	wg.Wait()
	log.Printf("Sent %d rows to EventBridge", counter)
}

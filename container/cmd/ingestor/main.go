package main

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"encoding/csv"
	"io"
	"log"
	"os"

	"ingestor/pkg/dynamo"
)

func main() {

	csvfile, err := os.Open("test.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
		panic(err)
	}
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)

	// Read headers - []string
	headers, err := reader.Read()
	if err != nil {
		log.Fatalln("Couldn't read headers")
		panic(err)
	}

	// Create AWS Session and Dynamo Client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	// Create Workers - Don't want to run thousands of concurrent requests
	var wg sync.WaitGroup
	ch := make(chan map[string]string)
	for w := 0; w < 50; w++ {
		wg.Add(1)
		go dynamo.InstertItem(svc, ch, &wg)
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
			continue
		}

		item := map[string]string{}

		for i, v := range row {
			item[headers[i]] = v
		}

		ch <- item
		// fmt.Println("Item added to channel - ", item)
		counter++
	}

	fmt.Println("It will print - ", counter)
	close(ch)
	wg.Wait()

}

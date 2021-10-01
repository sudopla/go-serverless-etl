package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var (
	ecsClient      *ecs.Client
	clusterName    string
	taskDefinition string
	subnet1        string
	subnet2        string
	containerName  string
)

func init() {
	log.Println("Initializing Lambda execution environment")
	log.Println("Initializing ECS Client...")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	ecsClient = ecs.NewFromConfig(cfg)

	log.Println("Getting environment variables...")
	clusterName = os.Getenv("CLUSTER_NAME")
	taskDefinition = os.Getenv("TASK_DEFINITION")
	subnet1 = os.Getenv("SUBNET_1")
	subnet2 = os.Getenv("SUBNET_2")
	containerName = os.Getenv("CONTAINER_NAME")
}

func handler(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3 := record.S3
		bucket := s3.Bucket.Name
		key := s3.Object.Key

		log.Printf("Input object. Bucket - %s, Key - %s", bucket, key)

		log.Println("Start Fargate Task")
		_, err := ecsClient.RunTask(context.TODO(), &ecs.RunTaskInput{
			Cluster:        &clusterName,
			TaskDefinition: &taskDefinition,
			NetworkConfiguration: &types.NetworkConfiguration{
				AwsvpcConfiguration: &types.AwsVpcConfiguration{
					Subnets: []string{subnet1, subnet2},
				},
			},
			Overrides: &types.TaskOverride{
				ContainerOverrides: []types.ContainerOverride{{
					Name: &containerName,
					Environment: []types.KeyValuePair{
						{Name: aws.String("BUCKET"), Value: &bucket},
						{Name: aws.String("KEY"), Value: &key},
					},
				}},
			},
		})

		if err != nil {
			log.Fatalf("Could not start Fargate Task for file %s\n Error - %s", key, err.Error())
		} else {
			log.Printf("Fargate task started successfully for file %s", key)
		}
	}
}

func main() {
	lambda.Start(handler)
}

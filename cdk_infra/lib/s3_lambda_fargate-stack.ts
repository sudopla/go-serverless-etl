/**
 * ECS Fargate Task that will process spreadsheets
 */
import * as path from 'path'
import * as ec2 from '@aws-cdk/aws-ec2'
import * as ecs from '@aws-cdk/aws-ecs'
import * as iam from '@aws-cdk/aws-iam'
import * as lambdaEvents from '@aws-cdk/aws-lambda-event-sources'
import * as goLambda from '@aws-cdk/aws-lambda-go'
import * as log from '@aws-cdk/aws-logs'
import * as s3 from '@aws-cdk/aws-s3'
import * as cdk from '@aws-cdk/core'
import * as perms from './iam_perms'

export interface S3LambdaFargateProps{
  eventBusName: string
}

export class S3LambdaFargateStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props: S3LambdaFargateProps) {
    super(scope, id, {
      description: 'Stack that defines S3 bucket, Fargate Task and Lambda that runs task'
    })

    const bucketName = 'real-estate-transactions-bucket'
    const ecsClusterName = 'IngestorFargateCluster'
    const containerName = 'Ingestor'

    /**
     * Fargate Task
     */
    // New VPC for Fargate tasks (Only public subnets, don't want to pay for NAT Gateways)
    const vpc = new ec2.Vpc(this, 'ECSFargateVPC', {
      cidr: '10.100.10.0/24',
      maxAzs: 2,
      subnetConfiguration: [
        { name: 'Subnet', subnetType: ec2.SubnetType.PUBLIC } // Subnet for each AZ
      ]
    })

    // ECS Cluster
    new ecs.Cluster(this, 'FargateCluster', {
      clusterName: ecsClusterName,
      vpc,
      containerInsights: true
    })

    // Task Role
    const taskRole = new iam.Role(this, 'TaskRole', { assumedBy: new iam.ServicePrincipal('ecs-tasks.amazonaws.com') })
    // Give container access to S3 bucket
    taskRole.addToPolicy(new iam.PolicyStatement({
      actions: [...perms.S3_READ_ACCESS, ...perms.S3_WRITE_ACCESS],
      resources: [`arn:aws:s3:::${bucketName}`, `arn:aws:s3:::${bucketName}/*`]
    }))
    // Give container access to EventBridge Bus
    taskRole.addToPolicy(new iam.PolicyStatement({
      actions: perms.EVENTBRIDGE_PUT_EVENTS,
      resources: [`arn:aws:events:${this.region}:${this.account}:event-bus/${props.eventBusName}`]
    }))

    // Create Task Definition
    const taskDefinition = new ecs.FargateTaskDefinition(this, 'CreateTenantTask', {
      cpu: 512, // .5 vCPU
      memoryLimitMiB: 1024, // Container memory
      taskRole
    })

    // Create log driver for container definition
    const logging = new ecs.AwsLogDriver({
      logGroup: new log.LogGroup(this, 'FargateTaskLogs', { logGroupName: '/aws/ecs/FaragateTask', retention: log.RetentionDays.ONE_MONTH }),
      streamPrefix: 'app' // Prefix for the log streams
    })

    // Create container image
    const containerImage = new ecs.AssetImage(path.join(__dirname, '..', '..', 'container'))

    // Add container definition to fargate task
    taskDefinition.addContainer(containerName, {
      image: containerImage,
      environment: {
        EVENTBRIDGE_BUS_NAME: props.eventBusName
      },
      logging
    })

    /**
     * Lambda that runs Fargate task after S3 event
     */
    const runFargateLambda = new goLambda.GoFunction(this, 'RunFarageTask', {
      functionName: 'RunFaragteTask',
      entry: path.join(__dirname, '..', '..', 'lambdas', 'RunFargateTask'),
      reservedConcurrentExecutions: 10,
      environment: {
        clusterName: ecsClusterName,
        taskDefinition: taskDefinition.taskDefinitionArn,
        subnet1: vpc.publicSubnets[0].subnetId,
        subnet2: vpc.publicSubnets[1].subnetId,
        CONTAINER_NAME: containerName
      }
    })

    // Allow Lambda function to run the task
    runFargateLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: perms.ECS_RUN_TASK,
      resources: [`arn:aws:ecs:${this.region}:${this.account}:task-definition/${taskDefinition.taskDefinitionArn}`]
    }))

    // The Lambda function needs iam:PassRole to pass the task execution and task role to the Fargate task
    runFargateLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: ['iam:PassRole'],
      resources: [taskDefinition.obtainExecutionRole().roleArn, taskRole.roleArn]
    }))

    /**
     * S3 Bucket
     */
    const bucket = new s3.Bucket(this, 'InputBucket', {
      bucketName,
      encryption: s3.BucketEncryption.S3_MANAGED
    })

    // Add S3 Lambda trigger for when files are uploaded to bucket
    runFargateLambda.addEventSource(new lambdaEvents.S3EventSource(bucket, {
      events: [s3.EventType.OBJECT_CREATED],
      filters: [{ prefix: 'upload/' }]
    }))

  }
}
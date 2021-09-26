#!/usr/bin/env node
import 'source-map-support/register'
import * as cdk from '@aws-cdk/core'
import { DynamoStack } from './dynamo-stack'
import { EventBridgeLambdasStack } from './eventbridge_lambdas-stack'
import { S3LambdaFargateStack } from './s3_lambda_fargate-stack'


const app = new cdk.App()

const tableName = 'real-state'
const eventBusName = 'etl-bus'

new DynamoStack(app, 'TableStack', {
  tableName
})

new S3LambdaFargateStack(app, 'S3LambdaFargateStack', {
  eventBusName
})

new EventBridgeLambdasStack(app, 'EventBridgeLambdasStack', {
  eventBusName,
  tableName
})

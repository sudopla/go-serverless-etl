/**
 * EventBridge Bus and rules with Lambda targets
 */
import * as path from 'path'
import * as events from '@aws-cdk/aws-events'
import * as eventTargates from '@aws-cdk/aws-events-targets'
import * as iam from '@aws-cdk/aws-iam'
import * as goLambda from '@aws-cdk/aws-lambda-go'
import * as sqs from '@aws-cdk/aws-sqs'
import * as cdk from '@aws-cdk/core'
import * as perms from './iam_perms'

export interface EventBridgeLambdasProps {
  tableName: string,
  eventBusName: string
}

export class EventBridgeLambdasStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props: EventBridgeLambdasProps) {
    super(scope, id)

    /**
     * Lambdas
     */
    const transformLambda = new goLambda.GoFunction(this, 'TransformLambda', {
      functionName: 'Transform',
      entry: path.join(__dirname, '..', '..', 'lambdas', 'Transform'),
      reservedConcurrentExecutions: 10,
      environment: {
        EVENTBRIDGE_BUS_NAME: props.eventBusName
      }
    })
    transformLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: perms.EVENTBRIDGE_PUT_EVENTS,
      resources: [`arn:aws:events:${this.region}:${this.account}:event-bus/${props.eventBusName}`]
    }))

    const loadLambda = new goLambda.GoFunction(this, 'LoadLambda', {
      functionName: 'Load',
      entry: path.join(__dirname, '..', '..', 'lambdas', 'Load'),
      reservedConcurrentExecutions: 10,
      environment: {
        TABLE_NAME: props.tableName
      }
    })
    loadLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: perms.Dynamo_WRITE_DATA_ACTIONS,
      resources: [`arn:aws:dynamodb:${this.region}:${this.account}:table/${props.tableName}`]
    }))

    /**
     * EventBridge Bus and Rules
     */
    const eventBus = new events.EventBus(this, 'EventBus', {
      eventBusName: props.eventBusName
    })

    const transformRule = new events.Rule(this, 'TransformRule', {
      eventBus,
      eventPattern: {
        source: ['app.container'],
        detailType: ['extraction-process'],
        detail: {
          status: ['row_sent']
        }
      }
    })
    transformRule.addTarget(new eventTargates.LambdaFunction(transformLambda, {
      deadLetterQueue: new sqs.Queue(this, 'TransformEventDlq', {
        queueName: 'TrasnformEventDlq',
        encryption: sqs.QueueEncryption.KMS_MANAGED
      }),
      retryAttempts: 2
    }))

    const loadRule = new events.Rule(this, 'LoadRule', {
      eventBus,
      eventPattern: {
        source: ['app.transform'],
        detailType: ['transform-process'],
        detail: {
          status: ['item_transformed']
        }
      }
    })
    loadRule.addTarget(new eventTargates.LambdaFunction(transformLambda, {
      deadLetterQueue: new sqs.Queue(this, 'LoadEventDlq', {
        queueName: 'LoadEventDlq',
        encryption: sqs.QueueEncryption.KMS_MANAGED
      }),
      retryAttempts: 2
    }))

  }
}
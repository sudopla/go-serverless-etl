/**
 * DynamoDB Table
 */
import * as dynamodb from '@aws-cdk/aws-dynamodb'
import * as cdk from '@aws-cdk/core'

export interface DynamoStackProps {
  tableName: string
}

export class DynamoStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props: DynamoStackProps) {
    super(scope, id, {
      description: 'DynamoDB Table Stack'
    })

    // Table
    new dynamodb.Table(this, 'Table', {
      tableName: props.tableName,
      encryption: dynamodb.TableEncryption.AWS_MANAGED,
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST, // on-demand
      partitionKey: { name: 'id', type: dynamodb.AttributeType.STRING }
      // sortKey: {}
    })

  }
}

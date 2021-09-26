/**
 * IAM actions
 */

// DynamoDB
export const Dynamo_WRITE_DATA_ACTIONS = [
  'dynamodb:BatchWriteItem',
  'dynamodb:PutItem',
  'dynamodb:UpdateItem',
  'dynamodb:DeleteItem'
]

// S3
export const S3_READ_ACCESS = [
  's3:ListBucket',
  's3:GetObject',
  's3:GetObjectTagging',
  's3:GetObjectVersionTagging'
]

export const S3_WRITE_ACCESS = [
  's3:ListBucket',
  's3:PutObject',
  's3:PutObjectAcl',
  's3:DeleteObject',
  's3:DeleteObjectVersion',
  's3:DeleteObjectTagging',
  's3:DeleteObjectVersionTagging',
  's3:PutObjectTagging',
  's3:PutObjectVersionTagging'
]

// EventBridge
export const EVENTBRIDGE_PUT_EVENTS = [
  'events:PutEvents'
]

// ECS
export const ECS_RUN_TASK = [
  'ecs:RunTask'
]

// Lambda
export const LAMBDA_INVOKE = [
  'lambda:InvokeFunction'
]
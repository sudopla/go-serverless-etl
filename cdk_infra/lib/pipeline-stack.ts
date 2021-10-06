/**
 * CodePipeline Stack
 */
import * as codebuild from '@aws-cdk/aws-codebuild'
import * as codepipeline from '@aws-cdk/aws-codepipeline'
import * as codepipeline_actions from '@aws-cdk/aws-codepipeline-actions'
import * as iam from '@aws-cdk/aws-iam'
import * as cdk from '@aws-cdk/core'

export interface CodePipelineStackProps {
  repoOwner: string,
  repoName: string
}

export class CodePipelineStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props: CodePipelineStackProps) {
    super(scope, id, {
      description: 'Stack that defines deployment pipeline'
    })

    const sourceArtifact = new codepipeline.Artifact()

    // CodeBuild role to deploy CDK stacks
    const codeBuildRole = new iam.Role(this, 'Role', {
      assumedBy: new iam.ServicePrincipal('codebuild.amazonaws.com')
    })
    // Restrict these policies later
    codeBuildRole.addToPolicy(new iam.PolicyStatement({
      resources: ['*'],
      actions: ['*']
    }))

    new codepipeline.Pipeline(this, 'Pipeline', {
      pipelineName: `${props.repoName}-pipeline`,
      stages: [
        {
          stageName: 'Source',
          actions: [
            new codepipeline_actions.GitHubSourceAction({
              actionName: 'GithubSource',
              owner: props.repoOwner,
              repo: props.repoName,
              branch: 'main',
              oauthToken: cdk.SecretValue.secretsManager('GithubCodePipelineToken', { jsonField: 'token' }),
              output: sourceArtifact
            })
          ]
        },
        {
          stageName: 'DeployApp',
          actions: [
            new codepipeline_actions.CodeBuildAction({
              actionName: 'DeployCdkStacks',
              input: sourceArtifact,
              project: new codebuild.PipelineProject(this, 'CodeBuildProject', {
                role: codeBuildRole,
                environment: {
                  buildImage: codebuild.LinuxBuildImage.STANDARD_5_0,
                  privileged: true // true to enable building docker image for fargate
                },
                buildSpec: codebuild.BuildSpec.fromObject({
                  version: '0.2',
                  phases: {
                    install: {
                      commands: [
                        'npm install'
                      ]
                    },
                    build: {
                      commands: [
                        'npm run cdk -- deploy TableStack S3LambdaFargateStack EventBridgeLambdasStack --require-approval never'
                      ]
                    }
                  }
                })
              })
            })
          ]
        }
      ]
    })


  }
}
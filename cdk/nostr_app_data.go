package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	codebuild "github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/pipelines"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewCdkStack(scope constructs.Construct, id *string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, id, props)

	connectHandler := lambdaFunction(stack, "Connect", "../app/functions/connect", nil)
	disconnectHandler := lambdaFunction(stack, "Disconnect", "../app/functions/disconnect", nil)
	defaultHandler := lambdaFunction(stack, "Default", "../app/functions/default", nil)
	requestHandler := lambdaFunction(stack, "Request", "../app/functions/request", nil)
	eventHandler := lambdaFunction(stack, "Event", "../app/functions/event", nil)

	webSocketApi := awsapigatewayv2.NewWebSocketApi(stack, jsii.String("mywsapi"), &awsapigatewayv2.WebSocketApiProps{
		ConnectRouteOptions: &awsapigatewayv2.WebSocketRouteOptions{
			Integration: awsapigatewayv2integrations.NewWebSocketLambdaIntegration(jsii.String("ConnectIntegration"), connectHandler, nil),
		},
		DisconnectRouteOptions: &awsapigatewayv2.WebSocketRouteOptions{
			Integration: awsapigatewayv2integrations.NewWebSocketLambdaIntegration(jsii.String("DisconnectIntegration"), disconnectHandler, nil),
		},
		DefaultRouteOptions: &awsapigatewayv2.WebSocketRouteOptions{
			Integration: awsapigatewayv2integrations.NewWebSocketLambdaIntegration(jsii.String("DefaultIntegration"), defaultHandler, nil),
		},
		RouteSelectionExpression: jsii.String("$request.body.[0]"),
	})
	webSocketApi.AddRoute(jsii.String("REQ"), &awsapigatewayv2.WebSocketRouteOptions{
		Integration: awsapigatewayv2integrations.NewWebSocketLambdaIntegration(jsii.String("RequestIntegration"), requestHandler, nil),
	})
	webSocketApi.AddRoute(jsii.String("EVENT"), &awsapigatewayv2.WebSocketRouteOptions{
		Integration: awsapigatewayv2integrations.NewWebSocketLambdaIntegration(jsii.String("EventIntegration"), eventHandler, nil),
	})
	wsStage := awsapigatewayv2.NewWebSocketStage(stack, jsii.String("mywsstage"), &awsapigatewayv2.WebSocketStageProps{
		AutoDeploy:   jsii.Bool(true),
		StageName:    jsii.String("test"),
		WebSocketApi: webSocketApi,
	})

	eventHandler.AddEnvironment(
		jsii.String("WS_API_ENDPOINT"),
		jsii.String(fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", *webSocketApi.ApiId(), *env().Region, *wsStage.StageName())),
		nil)

	//postHandler := lambdaFunction(stack, "Post", "../app/functions/post",
	//	map[string]*string{"WS_API_ENDPOINT": jsii.String(fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", *webSocketApi.ApiId(), *env().Region, *wsStage.StageName()))})
	//
	//httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("myhttpapi"), &awsapigatewayv2.HttpApiProps{
	//	CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
	//		AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{awsapigatewayv2.CorsHttpMethod_POST, awsapigatewayv2.CorsHttpMethod_OPTIONS},
	//		AllowOrigins: &[]*string{jsii.String("*")},
	//	},
	//	CreateDefaultStage: jsii.Bool(false),
	//	//DefaultIntegration: awsapigatewayv2integrations.NewHttpLambdaIntegration(jsii.String("PostIntegration"), postHandler, nil),
	//})
	//httpApi.AddStage(jsii.String("myhttpstage"), &awsapigatewayv2.HttpStageOptions{
	//	AutoDeploy: jsii.Bool(true),
	//	StageName:  jsii.String("test"),
	//})
	//httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
	//	Integration: awsapigatewayv2integrations.NewHttpLambdaIntegration(jsii.String("PostIntegration"), postHandler, nil),
	//	Path:        jsii.String("/"),
	//	Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_POST},
	//})
	//webSocketApi.GrantManageConnections(postHandler)

	awscdk.NewCfnOutput(stack, jsii.String("WSSApiURL"), &awscdk.CfnOutputProps{
		Value:       webSocketApi.ApiEndpoint(),
		Description: jsii.String("the URL to the WSS API"),
		ExportName:  jsii.String("WSSApiURL"),
	})

	//awscdk.NewCfnOutput(stack, jsii.String("HTTPApiURL"), &awscdk.CfnOutputProps{
	//	Value:       httpApi.ApiEndpoint(),
	//	Description: jsii.String("the URL to the HTTP API"),
	//	ExportName:  jsii.String("HTTPApiURL"),
	//})

	return stack
}

func lambdaFunction(stack awscdk.Stack, name, path string, env map[string]*string) awslambda.Function {
	awslogs.NewLogGroup(stack, jsii.String(name+"LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/" + name),
		Retention:     awslogs.RetentionDays_ONE_WEEK,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	return awslambda.NewFunction(stack, jsii.String(name+"Function"), &awslambda.FunctionProps{
		Code: awslambda.Code_FromAsset(jsii.String(path), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.23")),
				Command: &[]*string{
					jsii.String("bash"),
					jsii.String("-c"),
					jsii.String("GOCACHE=/tmp go mod tidy && GOCACHE=/tmp GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o /asset-output/bootstrap"),
				},
			},
		}),
		FunctionName: jsii.String(name),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(3)),
		Handler:      jsii.String("bootstrap"),
		Architecture: awslambda.Architecture_ARM_64(),
		Environment:  &env,
	})
}

func NewCdkApplication(scope constructs.Construct, id *string, props *awscdk.StageProps) awscdk.Stage {
	stage := awscdk.NewStage(scope, id, props)

	_ = NewCdkStack(stage, jsii.String("cdk-stack"), &awscdk.StackProps{Env: props.Env})

	return stage
}

func NewCdkPipeline(scope constructs.Construct, id *string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, id, props)

	// GitHub repo with owner and repository name
	githubRepo := pipelines.CodePipelineSource_GitHub(jsii.String("superkruger/nostr_app_data"), jsii.String("master"), &pipelines.GitHubSourceOptions{
		Authentication: awscdk.SecretValue_SecretsManager(jsii.String("github-token"), &awscdk.SecretsManagerSecretOptions{
			JsonField: jsii.String("github-token"),
		}),
	})

	// self mutating pipeline
	myPipeline := pipelines.NewCodePipeline(stack, jsii.String("cdkPipeline"), &pipelines.CodePipelineProps{
		PipelineName: jsii.String("CdkPipeline"),
		// self mutation true - pipeline changes itself before application deployment
		SelfMutation: jsii.Bool(true),
		CodeBuildDefaults: &pipelines.CodeBuildOptions{
			BuildEnvironment: &codebuild.BuildEnvironment{
				// image version 6.0 recommended for newer go version
				BuildImage: codebuild.LinuxBuildImage_FromCodeBuildImageId(jsii.String("aws/codebuild/standard:6.0")),
			},
		},
		Synth: pipelines.NewCodeBuildStep(jsii.String("Synth"), &pipelines.CodeBuildStepProps{
			Commands: &[]*string{
				jsii.String("cd cdk"),
				jsii.String("npm install -g aws-cdk"),
				jsii.String("cdk synth"),
				jsii.String("cd .."),
			},
			Input:                  githubRepo,
			PrimaryOutputDirectory: jsii.String("cdk/cdk.out"),
		}),
	})

	// deployment of actual CDK application
	myPipeline.AddStage(NewCdkApplication(stack, jsii.String("MyApplication"), &awscdk.StageProps{
		Env: env(),
	}), &pipelines.AddStageOpts{
		Post: &[]pipelines.Step{
			pipelines.NewCodeBuildStep(jsii.String("Manual Steps"), &pipelines.CodeBuildStepProps{
				Commands: &[]*string{
					jsii.String("echo \"My CDK App deployed, manual steps go here ... \""),
				},
			}),
		},
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkPipeline(app, jsii.String("CdkPipelineStack"), &awscdk.StackProps{
		Env: env(),
	})

	//stages := []string{"test", "prod"}

	//for _, stage := range stages {
	//NewNostrAppDataStack(app, prepend("test", "NostrAppDataStack"), &NostrAppDataStackProps{
	//	awscdk.StackProps{
	//		Env: env(),
	//	},
	//})
	//}

	app.Synth(nil)
}

func prepend(stage, id string) string {
	return fmt.Sprintf("%s-%s", stage, id)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	//return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String("418272791745"),
		Region:  jsii.String("us-east-1"),
	}

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}

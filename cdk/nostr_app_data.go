package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type NostrAppDataStackProps struct {
	awscdk.StackProps
}

func NewNostrAppDataStack(scope constructs.Construct, id string, props *NostrAppDataStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

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
				Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.21.13")),
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

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	//stages := []string{"test", "prod"}

	//for _, stage := range stages {
	NewNostrAppDataStack(app, prepend("test", "NostrAppDataStack"), &NostrAppDataStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})
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

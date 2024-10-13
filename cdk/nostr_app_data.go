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

	logGroup := awslogs.NewLogGroup(stack, jsii.String("Connect2LogGroup"), &awslogs.LogGroupProps{
		LogGroupName: jsii.String("/aws/lambda/Connect2"),
		Retention:    awslogs.RetentionDays_ONE_WEEK,
	})

	connectHandler := awslambda.NewFunction(stack, jsii.String("Connect2Function"), &awslambda.FunctionProps{
		Code: awslambda.Code_FromAsset(jsii.String("../app/functions/test"), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.21.13")),
				Command: &[]*string{
					jsii.String("bash"),
					jsii.String("-c"),
					jsii.String("GO111MODULE=on GOCACHE=/tmp go mod tidy && GOCACHE=/tmp GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o /asset-output/bootstrap"),
				},
			},
		}),
		FunctionName: jsii.String("Connect2"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		MemorySize:   jsii.Number(128),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
		Handler:      jsii.String("bootstrap"),
		Architecture: awslambda.Architecture_ARM_64(),
	})

	webSocketApi := awsapigatewayv2.NewWebSocketApi(stack, jsii.String("mywsapi"), &awsapigatewayv2.WebSocketApiProps{
		ConnectRouteOptions: &awsapigatewayv2.WebSocketRouteOptions{
			Integration: awsapigatewayv2integrations.NewWebSocketLambdaIntegration(jsii.String("ConnectIntegration"), connectHandler, nil),
		},
		//DisconnectRouteOptions: &apigwv2.WebSocketRouteOptions{
		//	Integration: awsapigwintegrations.NewWebSocketLambdaIntegration(jsii.String("DisconnectIntegration"), disconnectHandler),
		//},
		//DefaultRouteOptions: &apigwv2.WebSocketRouteOptions{
		//	Integration: awsapigwintegrations.NewWebSocketLambdaIntegration(jsii.String("DefaultIntegration"), defaultHandler),
		//},
	})

	awsapigatewayv2.NewWebSocketStage(stack, jsii.String("mywsstage"), &awsapigatewayv2.WebSocketStageProps{
		AutoDeploy:   jsii.Bool(true),
		StageName:    jsii.String("test"),
		WebSocketApi: webSocketApi,
	})

	//echoLambdaFunc := awscdklambdago.NewGoFunction(stack, jsii.String("EchoFunc"), &awscdklambdago.GoFunctionProps{
	//	FunctionName: jsii.String("EchoFunc"),
	//	Description:  jsii.String("an apigw handler that returns IP and User-Agent as JSON"),
	//	Entry:        jsii.String("../app/functions/test"),
	//})
	//echoApi := awscdkapigw.NewHttpApi(stack, jsii.String("EchoApi"), nil)
	//
	//echoApi.AddRoutes(&awscdkapigw.AddRoutesOptions{
	//	Path:        jsii.String("/"),
	//	Methods:     &[]awscdkapigw.HttpMethod{awscdkapigw.HttpMethod_GET},
	//	Integration: awsapigwintegrations.NewHttpLambdaIntegration(jsii.String("EchoApiIntegration"), echoLambdaFunc, nil),
	//})

	awscdk.NewCfnOutput(stack, jsii.String("WSSApiURL"), &awscdk.CfnOutputProps{
		Value:       webSocketApi.ApiEndpoint(),
		Description: jsii.String("the URL to the WSS API"),
		ExportName:  jsii.String("WSSApiURL"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("LogGroupOut"), &awscdk.CfnOutputProps{
		Value:       logGroup.LogGroupArn(),
		Description: jsii.String("log group"),
		ExportName:  jsii.String("Connect2LogGroup"),
	})

	return stack
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

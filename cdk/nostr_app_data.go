package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	awscdkapigw "github.com/aws/aws-cdk-go/awscdkapigatewayv2alpha/v2"
	awsapigwintegrations "github.com/aws/aws-cdk-go/awscdkapigatewayv2integrationsalpha/v2"
	awscdklambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
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

	echoLambdaFunc := awscdklambdago.NewGoFunction(stack, jsii.String("EchoFunc"), &awscdklambdago.GoFunctionProps{
		FunctionName: jsii.String("EchoFunc"),
		Description:  jsii.String("an apigw handler that returns IP and User-Agent as JSON"),
		Entry:        jsii.String("../app/functions/test"),
	})

	echoApi := awscdkapigw.NewHttpApi(stack, jsii.String("EchoApi"), nil)

	echoApi.AddRoutes(&awscdkapigw.AddRoutesOptions{
		Path:        jsii.String("/"),
		Methods:     &[]awscdkapigw.HttpMethod{awscdkapigw.HttpMethod_GET},
		Integration: awsapigwintegrations.NewHttpLambdaIntegration(jsii.String("EchoApiIntegration"), echoLambdaFunc, nil),
	})

	awscdk.NewCfnOutput(stack, jsii.String("EchoApiURL"), &awscdk.CfnOutputProps{
		Value:       echoApi.ApiEndpoint(),
		Description: jsii.String("the URL to the Echo API"),
		ExportName:  jsii.String("EchoApiURL"),
	})

	// Create role for lambda function.
	//lambdaRole := awsiam.NewRole(stack, jsii.String("LambdaRole"), &awsiam.RoleProps{
	//	RoleName:  jsii.String(*stack.StackName() + "-LambdaRole"),
	//	AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	//	ManagedPolicies: &[]awsiam.IManagedPolicy{
	//		awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBFullAccess")),
	//		awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("CloudWatchFullAccess")),
	//	},
	//})

	//_ = awscdklambdago.NewGoFunction(stack, jsii.String("TestFunction"), &awscdklambdago.GoFunctionProps{
	//	FunctionName: jsii.String("TestFunction"),
	//	Description:  jsii.String("a test function"),
	//	Entry:        jsii.String("functions/test"),
	//	Role:         lambdaRole,
	//})

	// Create put-chat-records function.
	//_ = awslambda.NewFunction(stack, jsii.String("TestFunction"), &awslambda.FunctionProps{
	//	FunctionName: jsii.String("TestFunction"),
	//	Runtime:      awslambda.Runtime_GO_1_X(),
	//	MemorySize:   jsii.Number(128),
	//	Timeout:      awscdk.Duration_Seconds(jsii.Number(60)),
	//	Code:         awslambda.AssetCode_FromAsset(jsii.String("functions/test/."), nil),
	//
	//	//Handler:      jsii.String("test"),
	//	Architecture: awslambda.Architecture_X86_64(),
	//	Role:         lambdaRole,
	//	LogRetention: awslogs.RetentionDays_ONE_WEEK,
	//	Environment:  &map[string]*string{},
	//})

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("NostrAppDataQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

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

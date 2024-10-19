package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	codebuild "github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/pipelines"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/superkruger/nostr_app_data/cdk/config"
)

func NewCdkAppStack(scope constructs.Construct, id *string, cfg config.Config, props *awscdk.StackProps) awscdk.Stack {
	name := func(name string) string {
		return fmt.Sprintf("%s-%s", *id, name)
	}
	stack := awscdk.NewStack(scope, id, props)

	connectHandler := lambdaFunction(stack, name("Connect"), "./functions/connect", map[string]*string{
		"DB_SECRET": jsii.String(cfg.DBSecret),
	})
	disconnectHandler := lambdaFunction(stack, name("Disconnect"), "./functions/disconnect", map[string]*string{
		"DB_SECRET": jsii.String(cfg.DBSecret),
	})
	defaultHandler := lambdaFunction(stack, name("Default"), "./functions/default", nil)
	requestHandler := lambdaFunction(stack, name("Request"), "./functions/request", map[string]*string{
		"DB_SECRET": jsii.String(cfg.DBSecret),
	})
	eventHandler := lambdaFunction(stack, name("Event"), "./functions/event", map[string]*string{
		"DB_SECRET": jsii.String(cfg.DBSecret),
	})

	webSocketApi := awsapigatewayv2.NewWebSocketApi(stack, jsii.String(name("WSSAPI")), &awsapigatewayv2.WebSocketApiProps{
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
	awsapigatewayv2.NewWebSocketStage(stack, jsii.String("WSSStage"), &awsapigatewayv2.WebSocketStageProps{
		AutoDeploy:   jsii.Bool(true),
		StageName:    jsii.String(cfg.Name),
		WebSocketApi: webSocketApi,
	})

	eventHandler.AddEnvironment(
		jsii.String("WS_API_ENDPOINT"),
		jsii.String(fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com", *webSocketApi.ApiId(), *props.Env.Region)),
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

	awscdk.NewCfnOutput(stack, jsii.String(name("WSSApiURL")), &awscdk.CfnOutputProps{
		Value:       webSocketApi.ApiEndpoint(),
		Description: jsii.String("the URL to the WSS API"),
		ExportName:  jsii.String(name("WSSApiURL")),
	})

	//awscdk.NewCfnOutput(stack, jsii.String("HTTPApiURL"), &awscdk.CfnOutputProps{
	//	Value:       httpApi.ApiEndpoint(),
	//	Description: jsii.String("the URL to the HTTP API"),
	//	ExportName:  jsii.String("HTTPApiURL"),
	//})

	return stack
}

func lambdaFunction(stack awscdk.Stack, name, path string, env map[string]*string) awslambda.Function {
	lambdaRole := awsiam.NewRole(stack, jsii.String(name+"Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		RoleName:  jsii.String(name + "-lambda-role"),
	})

	lambdaRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   &[]*string{jsii.String("ssm:GetParameter"), jsii.String("secretsmanager:GetSecretValue"), jsii.String("kms:Decrypt")},
		Effect:    awsiam.Effect_ALLOW,
		Resources: &[]*string{jsii.String("arn:aws:secretsmanager:*:*")},
	}))

	awslogs.NewLogGroup(stack, jsii.String(name+"LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/" + name),
		Retention:     awslogs.RetentionDays_ONE_WEEK,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	return awslambda.NewFunction(stack, jsii.String(name+"Function"), &awslambda.FunctionProps{
		Code: awslambda.Code_FromAsset(jsii.String("../app"), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awscdk.DockerImage_FromRegistry(jsii.String("golang:1.23.2")),
				Command: &[]*string{
					jsii.String("bash"),
					jsii.String("-c"),
					jsii.String("GOCACHE=/tmp GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o /asset-output/bootstrap " + path),
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
		Role:         lambdaRole,
	})
}

func NewCdkApplication(scope constructs.Construct, id *string, cfg config.Config, props *awscdk.StageProps) awscdk.Stage {
	name := func(name string) string {
		return fmt.Sprintf("%s-%s", *id, name)
	}
	stage := awscdk.NewStage(scope, id, props)
	_ = NewCdkAppStack(stage, jsii.String(name("Stack")), cfg, &awscdk.StackProps{Env: props.Env})
	return stage
}

func NewCdkPipeline(scope constructs.Construct, id *string, cfg config.Config, props *awscdk.StackProps) awscdk.Stack {
	name := func(name string) string {
		return fmt.Sprintf("%s-%s", *id, name)
	}
	stack := awscdk.NewStack(scope, id, props)
	// GitHub repo with owner and repository name
	githubRepo := pipelines.CodePipelineSource_GitHub(jsii.String("superkruger/nostr_app_data"), jsii.String(cfg.Branch), &pipelines.GitHubSourceOptions{
		Authentication: awscdk.SecretValue_SecretsManager(jsii.String("github-token"), &awscdk.SecretsManagerSecretOptions{
			JsonField: jsii.String("github-token"),
		}),
	})
	// self mutating pipeline
	myPipeline := pipelines.NewCodePipeline(stack, jsii.String(name("Pipeline")), &pipelines.CodePipelineProps{
		PipelineName: jsii.String(name("Pipeline")),
		// self mutation true - pipeline changes itself before application deployment
		SelfMutation: jsii.Bool(true),
		CodeBuildDefaults: &pipelines.CodeBuildOptions{
			BuildEnvironment: &codebuild.BuildEnvironment{
				// image version 6.0 recommended for newer go version
				BuildImage: codebuild.LinuxBuildImage_FromCodeBuildImageId(jsii.String("aws/codebuild/standard:7.0")),
			},
		},
		Synth: pipelines.NewCodeBuildStep(jsii.String("Synth"), &pipelines.CodeBuildStepProps{
			Commands: &[]*string{
				//jsii.String("go install golang.org/x/tools/gopls@latest"),
				jsii.String("cd cdk"),
				jsii.String("npm install -g aws-cdk"),
				jsii.String("cdk synth --context environment=" + cfg.Name),
				jsii.String("cd .."),
			},
			Input:                  githubRepo,
			PrimaryOutputDirectory: jsii.String("cdk/cdk.out"),
		}),
	})
	// deployment of actual CDK application
	myPipeline.AddStage(NewCdkApplication(stack, jsii.String(fmt.Sprintf("%s-%s", cfg.Name, "NostrAppData")), cfg, &awscdk.StageProps{
		Env: props.Env,
	}), &pipelines.AddStageOpts{
		Post: &[]pipelines.Step{
			pipelines.NewCodeBuildStep(jsii.String("Manual Steps"), &pipelines.CodeBuildStepProps{
				Commands: &[]*string{
					jsii.String("echo \"NostrAppData deployed, manual steps go here ... \""),
				},
			}),
		},
	})
	return stack
}

func main() {
	defer jsii.Close()
	app := awscdk.NewApp(nil)
	envName := app.Node().GetContext(jsii.String("environment"))
	cfg := config.MustNewConfig(envName.(string))
	NewCdkPipeline(app, jsii.String(fmt.Sprintf("%s-%s", cfg.Name, "PipelineStack")), cfg, &awscdk.StackProps{
		Env: env(cfg),
	})
	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env(cfg config.Config) *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	//return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(cfg.AccountID), //jsii.String("418272791745"),
		Region:  jsii.String(cfg.Region),    //jsii.String("us-east-1"),
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

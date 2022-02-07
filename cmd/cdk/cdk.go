package main

import (
	"cdk/internal/cloudfront"
	"cdk/internal/github"
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/awss3"
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

type OidcProviderStack struct {
	awscdk.Stack
	OIDCProvider awsiam.CfnOIDCProvider
}

// NewOidcProviderStack creates a GitHub OIDC provider to assume deployment role
// without a need to setup IAM Users for programmatic access.
// The GitHub OIDC Provider only needs to be created once per account
// (i.e. multiple IAM Roles that can be assumed by the GitHub's OIDC can share a single OIDC Provider)
//
// See https://github.com/aws-actions/configure-aws-credentials
func NewOidcProviderStack(scope constructs.Construct, id string, props *awscdk.StackProps) *OidcProviderStack {
	stack := &OidcProviderStack{Stack: awscdk.NewStack(scope, &id, props)}

	stack.OIDCProvider = awsiam.NewCfnOIDCProvider(stack.Stack, jsii.String("GitHubOIDCProvider"), &awsiam.CfnOIDCProviderProps{
		Url:            jsii.String("https://token.actions.githubusercontent.com"),
		ClientIdList:   jsii.Strings("sts.amazonaws.com"),
		ThumbprintList: jsii.Strings("a031c46782e6e6c662c2c87c76da9aa62ccabd8e"),
	})

	return stack
}

type AppStackProps struct {
	awscdk.StackProps
	OIDCProvider awsiam.CfnOIDCProvider
	BucketName   string
	GitHubOrg    string
	RepoName     string
}

func NewAppStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	bucket := awss3.NewBucket(stack, jsii.String("Bucket"), &awss3.BucketProps{
		AutoDeleteObjects: jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		BucketName:        &props.BucketName,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
	})

	awscdk.NewCfnOutput(stack, jsii.String("S3BuckeName"), &awscdk.CfnOutputProps{
		Value: bucket.BucketName(),
	})

	distr := awscloudfront.NewDistribution(stack, jsii.String("Distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewS3Origin(bucket, nil),
		},
	})

	cloudfront.SetRealtimeMetrics(stack, distr, true)

	awscdk.NewCfnOutput(stack, jsii.String("DistributionDomainName"), &awscdk.CfnOutputProps{
		Value: distr.DistributionDomainName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("DistributionID"), &awscdk.CfnOutputProps{
		Value: distr.DistributionId(),
	})

	role := github.NewDeploymentRole(stack, props.OIDCProvider, props.GitHubOrg, props.RepoName)

	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("cloudfront:CreateInvalidation"),
		Resources: jsii.Strings(fmt.Sprintf("arn:aws:cloudfront::%s:distribution/*", *stack.Account())),
	}))

	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"s3:DeleteObject",
			"s3:GetBucketLocation",
			"s3:GetObject",
			"s3:ListBucket",
			"s3:PutObject",
		),
		Resources: jsii.Strings(
			*bucket.BucketArn(),
			fmt.Sprintf("%s/*", *bucket.BucketArn()),
		),
	}))

	awscdk.NewCfnOutput(stack, jsii.String("DeploymentRoleArn"), &awscdk.CfnOutputProps{
		Value: role.RoleArn(),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	stack1 := NewOidcProviderStack(app, "GitHubOIDCStack", &awscdk.StackProps{
		StackName: jsii.String("github-oidc"),
		Env:       env(),
	})

	NewAppStack(app, "AppStack", &AppStackProps{
		awscdk.StackProps{
			StackName: jsii.String("demo-s3-cloudfront-assets"),
			Env:       env(),
		},
		stack1.OIDCProvider,
		"demo-assets-ksds3f7s",
		"vorotech",
		"aws-cdk-s3-cloudfront-assets",
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}

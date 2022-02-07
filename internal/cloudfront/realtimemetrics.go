package cloudfront

import (
	_ "embed"
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/awslogs"
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

const (
	DistributionIDKey  = "DistributionId"
	RealtimeMetricsKey = "RealtimeMetrics"
)

//go:embed lambda/handler.js
var code string

// SetRealtimeMetrics enables/disables CloutFront distribution realtime metrics
// with a technic which involves CustomResourse + Lambda combination.
//
// NOTE: If realtimeMetrics first enabled amd later key removed from config,
// it won't disable the monitoring subscription status, but only delete
// the custom resources from CloudFormation stack
func SetRealtimeMetrics(scope constructs.Construct, distr awscloudfront.Distribution, isEnabled bool) {
	lambda := awslambda.NewFunction(scope, jsii.String("MonitoringSubscriptionLambda"), &awslambda.FunctionProps{
		Handler:      jsii.String("index.handler"),
		Runtime:      awslambda.Runtime_NODEJS_14_X(),
		Description:  jsii.String("Sets or disables the CloudFront realtime monitoring subscription"),
		LogRetention: awslogs.RetentionDays_ONE_DAY,
		MemorySize:   jsii.Number(128), // minimum
		Timeout:      awscdk.Duration_Seconds(jsii.Number(8)),
		Code:         awslambda.Code_FromInline(&code),
	})

	lambda.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:    jsii.String("AllowSetRealtimeMonitoringSubscription"),
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"cloudfront:GetMonitoringSubscription",
			"cloudfront:CreateMonitoringSubscription",
			"cloudfront:DeleteMonitoringSubscription",
		),
		Resources: jsii.Strings("*"),
	}))

	lambda.Node().AddDependency(distr)

	// Logical ID of the custom resource has to be linked to passed props (similar to hash func)
	// to make sure the lambda is executed during each CloudFormation Stack update whenever props are changed.
	logicalID := jsii.String(fmt.Sprintf("InvokeMonitoringSubscriptionLambda%d", btoi(isEnabled)))
	awscdk.NewCustomResource(scope, logicalID, &awscdk.CustomResourceProps{
		ServiceToken: lambda.FunctionArn(),
		Properties: &map[string]interface{}{
			DistributionIDKey:  distr.DistributionId(),
			RealtimeMetricsKey: isEnabled,
		},
	})
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

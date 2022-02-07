package github

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/jsii-runtime-go"
)

func NewDeploymentRole(stack awscdk.Stack, oidcProvider awsiam.CfnOIDCProvider, gitHubOrg, repoName string) awsiam.Role {
	principal := awsiam.NewFederatedPrincipal(
		oidcProvider.AttrArn(),
		&map[string]interface{}{
			"StringLike": map[string]string{
				"token.actions.githubusercontent.com:sub": fmt.Sprintf("repo:%s/%s:*", gitHubOrg, repoName),
			},
		},
		jsii.String("sts:AssumeRoleWithWebIdentity"),
	)

	role := awsiam.NewRole(stack, jsii.String("DeploymentRole"), &awsiam.RoleProps{
		AssumedBy: principal,
		Path:      jsii.String("/deployment-role/"),
	})

	return role
}

# aws-cdk-s3-cloudfront-assets

An application is a S3 bucket with a content automatically synced from the `assets/` folder.
The bucket don't allow a public access, instead the the CloudFront Distribution configured to use 
the bucket as a source of content origin.

This demo only shows the continuous deployment approach for the application itself,
leaving the infrastructure CD aside.

The application's infrastructure is deployed to AWS with CloudFormation via AWS Code Deployment Kit (CDK).

First, deploy AWS CDK stack from local machine. 

```sh
export AWS_PROFILE=...
export AWS_REGION=...
cdk diff
cdk deploy --all
```

The application stack have two outputs:

* `S3BucketName` is the bucket name containing assets files and configured as the CloudFront origin.
* `DistributionDomainName` is the CDN domain address which delivers bucket content 
  which in turn will be synced from `assets/` folder by GitHub workflow action after merging the Pull Request.
* `DistributionID` is the CloudFront Distribution identifier to invalidate.
* `DeploymentRoleArn` is the role ARN which is granted with least privilege to sync S3 bucket content and
  invalidate CloudFront Distribution on content changes (by default existing content is cached for 24h).

Next, configure the following GitHub secrets:

* `AWS_S3_BUCKET=${S3BucketName}`
* `AWS_ROLE_TO_ASSUME=${DistributionDomainName}`
* `AWS_DISTRIBUTION_ID=${DistributionID}`

Finally, make a Pull Request and add/edit image(s) to the `assets/heroes` folder.

After Pull Request will be merged, the GitHub `cd` workflow action should 
sync created/updated files and invalidated CDN (for changed files only).

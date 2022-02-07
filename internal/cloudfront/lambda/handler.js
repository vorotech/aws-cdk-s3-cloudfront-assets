var aws = require('aws-sdk');
var response = require('cfn-response');

var cloudfront = new aws.CloudFront();

exports.handler = function (event, context) {
    console.log(event)
    var distributionId = event.ResourceProperties.DistributionId
    var realtimeMetrics = event.ResourceProperties.RealtimeMetrics === 'true'

    if (event.RequestType == "Delete") {
        console.log("Response immediately on custom resource deletion")
        response.send(event, context, response.SUCCESS);
        return;
    }

    var strStatus = realtimeMetrics ? "Enabled" : "Disabled"
    var params = {
        DistributionId: distributionId,
        MonitoringSubscription: {
            RealtimeMetricsSubscriptionConfig: {
                RealtimeMetricsSubscriptionStatus: strStatus
            }
        }
    };
    console.log("Set realtime monitoring subscription status to ", strStatus)
    cloudfront.createMonitoringSubscription(params, function (err, data) {
        if (err) console.error(err); // an error occurred
        else console.log(data);      // successful response
        response.send(event, context, response.SUCCESS);
    });
};

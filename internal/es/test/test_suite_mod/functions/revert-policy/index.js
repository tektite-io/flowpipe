/*
SPDX-FileCopyrightText: 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
SPDX-License-Identifier: MIT-0
*/

// Mocked env
process.env = {
  "restrictedActions": "s3:DeleteBucket,s3:DeleteObject"
}

let restrictedActions = process.env.restrictedActions.split(",");
const AWS = require('aws-sdk')
var iam = new AWS.IAM({apiVersion: '2010-05-08'});
exports.handler = async(event, context) => {

    // Mocked event
    event = {
      policy: '{"Version":"2012-10-17","Statement":[{"Sid":"VisualEditor0","Effect":"Allow","Action":["s3:AddBucket","s3:AddObject"],"Resource":"*"}]}',
      policyMeta: {
        "arn": "arn:aws:iam::123456789012:policy/ExamplePolicy",
        "policyName": "ExamplePolicy",
        "defaultVersionId": "v1"
      }
    }

    /* The following command Create a new blank policy version as a placeholder*/
    var params = {
      PolicyArn: event.policyMeta.arn, /* required */
      PolicyDocument: '{"Version": "2012-10-17","Statement": [{ "Sid": "VisualEditor0","Effect": "Allow","Action": "logs:GetLogGroupFields", "Resource": "*"}] }',
      SetAsDefault: true
    };

    try {
      const res = await iam.createPolicyVersion(params).promise()
    }catch(err){
      console.error(err)
    }

    //Delete the restricted policy version
    var params = {
      PolicyArn: event.policyMeta.arn, /* required */
      VersionId: event.policyMeta.defaultVersionId /* required */
    };

    // TODO: Uncomment this block to delete the restricted policy version
    // try {
    //   const res = await iam.deletePolicyVersion(params).promise()
    // }catch(err){
    //   console.error(err)
    // }
    console.log("Deleted the restricted policy version")

    return {
      "message": `Policy ${event.policyMeta.policyName} Has been altered and contains restricted Actions: ${event.policy}, please approve or deny this change`,
      "action": "remedy"
    };

}

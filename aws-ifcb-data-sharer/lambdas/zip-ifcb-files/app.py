import json
import boto3
import os

# Initialize the SQS client
sqs_client = boto3.client("sqs")


def lambda_handler(event, context):

    try:
        print(event)
        # Process each record from the DynamoDB stream
        for record in event["Records"]:
            print(record)
            # Only handle MODIFY or INSERT events
            if record["eventName"] in ["INSERT", "MODIFY"]:
                # Extract the new image (updated item) from the stream
                new_image = record.get("dynamodb", {}).get("NewImage", {})
                print("new_image", new_image)

        return {
            "statusCode": 200,
            "body": json.dumps({"status": "Messages processed successfully."}),
        }
    except Exception as e:
        print(f"Error processing DynamoDB stream: {e}")
        return {"statusCode": 500, "body": json.dumps({"error": str(e)})}

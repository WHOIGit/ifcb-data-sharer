import json
import boto3
import os
import magic
import re
from pathlib import Path
from botocore.exceptions import ClientError

# import IFCB utilities from https://github.com/joefutrelle/pyifcb package
from ifcb.data.adc import AdcFile
from ifcb.data.hdr import parse_hdr_file
from ifcb.data.identifiers import parse

s3_client = boto3.client("s3")
valid_extensions = [".adc", ".hdr", ".roi", ".csv"]


def extract_year_month_and_prefix(filename):
    # Regular expression to match the date and extract year, month, and prefix
    match = re.search(r"D(\d{4})(\d{2})\d{2}T", filename)

    if match:
        year = int(match.group(1))
        # month = int(match.group(2))
        prefix = filename.split("T")[0]
        return year, prefix
    else:
        raise ValueError("Invalid filename format")


def lambda_handler(event, context):
    print(event)
    # parse the S3 file received
    try:
        s3_Bucket_Name = event["Records"][0]["s3"]["bucket"]["name"]
        s3_File_Name = event["Records"][0]["s3"]["object"]["key"]
        print(s3_File_Name)

        # check the file extension
        # Extract the file extension (returns a tuple)
        s3_Root, file_extension = os.path.splitext(s3_File_Name)
        username = s3_Root.split("/")[0]
        dataset = s3_Root.split("/")[1]

        print("file_extension", file_extension, username, dataset)
        print("user", username)
        print("dataset", dataset)

        if file_extension not in valid_extensions:
            # delete file from S3 if not in whitelist
            s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
            print("file deleted")
            return {
                "statusCode": 200,
                "body": json.dumps(f"{s3_File_Name} not valid file type. Deleted"),
            }

        # check if file already exists in final location, if not then
        # download file to tmp directory to test file contents.
        # get the Bin pid, pyifcb needs correct file name to parse
        valid_file = False
        bin_pid = Path(s3_File_Name).stem
        tmp_file = f"/tmp/{bin_pid}{file_extension}"
        year, prefix = extract_year_month_and_prefix(bin_pid)
        destination_key = (
            f"{username}/{dataset}/{year}/{prefix}/{bin_pid}{file_extension}"
        )

        try:
            s3_client.head_object(Bucket=s3_Bucket_Name, Key=destination_key)
            print("File already exists, delete new version")
            s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
            return {
                "statusCode": 200,
                "body": json.dumps(f"{s3_File_Name} already exists. Deleted"),
            }
        except ClientError as e:
            # no existing file, continue parsing
            if e.response["Error"]["Code"] == "404":
                pass
            else:
                print(e)
                return False

        result = s3_client.download_file(s3_Bucket_Name, s3_File_Name, tmp_file)

        # parse file with pyifcb package
        # does it return a valid instance of ADC, HDR or ROI file
        if file_extension == ".adc":
            try:
                adc = AdcFile(tmp_file, True)
                print("ADC length: ", adc.__len__())
                print("ADC lid: ", adc.lid)
                valid_file = True
                dynamo_field = "hasAdc"
            except Exception as e:
                # error parsing ADC, delete file
                print("validation error", e)
                s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
                print("file deleted")

        if file_extension == ".hdr":
            try:
                hdr = parse_hdr_file(tmp_file)
                print("HDR file: ", hdr)
                valid_file = True
                dynamo_field = "hasHdr"
            except Exception as e:
                # error parsing HDR, delete file
                print("validation error", e)
                s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
                print("file deleted")

        if file_extension == ".roi":
            # check if it's a binary data file
            mime_type = magic.from_file(tmp_file, mime=True)
            if mime_type == "application/octet-stream":
                # check if file PID is valid
                try:
                    resp = parse(bin_pid)
                    print("valid ROI file.")
                    valid_file = True
                    dynamo_field = "hasRoi"
                except Exception as e:
                    # invalid pid, delete file
                    print("validation error", e)
                    s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
                    print("file deleted")
            else:
                print("INVALID ROI file.")
                s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
                print("file deleted")

        if file_extension == ".csv":
            # check if it's a csv file
            mime_type = magic.from_file(tmp_file, mime=True)
            if mime_type == "text/csv":
                print("valid CSV file.")
                valid_file = True
                dynamo_field = "hasCsv"
            else:
                print("INVALID CSV file.")
                s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
                print("file deleted")
        # delete file from tmp dir
        print("remove tmp file")
        os.remove(tmp_file)

    except Exception as err:
        print(err)
        return None

    if valid_file:
        # move file to valid file directory structure

        # destination_key_2 = f"{username}/{dataset}/{bin_pid}{file_extension}"
        print("Destination key check:")
        print(destination_key, s3_File_Name)

        if destination_key != s3_File_Name:
            # Copy the object if it's in wrong place
            s3_client.copy_object(
                Bucket=s3_Bucket_Name,
                CopySource={"Bucket": s3_Bucket_Name, "Key": s3_File_Name},
                Key=destination_key,
            )

            # Delete the original object
            s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
            print(f"Moved {s3_File_Name} to {destination_key}")

        # save status to Dynamo
        dynamodb = boto3.resource("dynamodb")
        table_name = "ifcb-data-sharer-bins"
        table = dynamodb.Table(table_name)

        try:
            print("saving to Dynamo...")
            # Dynamo needs float converted to Decimal for N type
            response = table.update_item(
                Key={
                    "username": username,
                    "pid": bin_pid,
                },
                UpdateExpression=f"SET {dynamo_field} = :{dynamo_field}, dataset = :dataset, s3KeyRoot = :s3KeyRoot",
                ExpressionAttributeValues={
                    f":{dynamo_field}": True,
                    ":dataset": dataset,
                    ":s3KeyRoot": s3_Root,
                },
                ReturnValues="UPDATED_NEW",
            )
            print(f"{bin_pid} saved")
            print(response)
        except Exception as err:
            print(err)
            return None

        return {
            "statusCode": 200,
            "body": json.dumps(f"{s3_File_Name} validated"),
        }
    else:
        return {
            "statusCode": 200,
            "body": json.dumps(f"{s3_File_Name} failed parsing. Deleted"),
        }

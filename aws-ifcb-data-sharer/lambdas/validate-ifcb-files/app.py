import json
import boto3
import os
import magic
from pathlib import Path

# import IFCB utilities from https://github.com/joefutrelle/pyifcb package
from ifcb.data.adc import AdcFile
from ifcb.data.hdr import parse_hdr_file

s3_client = boto3.client("s3")
valid_extensions = [".adc", ".hdr", ".roi"]


def lambda_handler(event, context):
    print(event)
    # parse the S3 file received
    try:
        s3_Bucket_Name = event["Records"][0]["s3"]["bucket"]["name"]
        s3_File_Name = event["Records"][0]["s3"]["object"]["key"]
        print(s3_File_Name)

        # check the file extension
        # Extract the file extension (returns a tuple)
        _, file_extension = os.path.splitext(s3_File_Name)
        print("file_extension", file_extension)

        if file_extension not in valid_extensions:
            # delete file from S3 if not in whitelist
            s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
            print("file deleted")
            return {
                "statusCode": 200,
                "body": json.dumps(f"{s3_File_Name} not valid file type. Deleted"),
            }

        # download file to tmp directory to test file contents
        # get the Bin pid, pyifcb needs correct file name to parse
        valid_file = False
        bin_pid = Path(s3_File_Name).stem
        tmp_file = f"/tmp/{bin_pid}{file_extension}"
        print("tmp_file", tmp_file)
        result = s3_client.download_file(s3_Bucket_Name, s3_File_Name, tmp_file)

        # parse file with pyifcb package
        # does it return a valid instance of ADC, HDR or ROI file
        if file_extension == ".adc":
            try:
                adc = AdcFile(tmp_file, True)
                print("ADC length: ", adc.__len__())
                print("ADC lid: ", adc.lid)
                valid_file = True
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
            except Exception as e:
                # error parsing HDR, delete file
                print("validation error", e)
                s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
                print("file deleted")

        if file_extension == ".roi":
            # check if it's a binary data file
            mime_type = magic.from_file("D20230729T032906_IFCB110.roi", mime=True)
            if mime_type == "application/octet-stream":
                print("valid ROI file.")
                valid_file = True
            else:
                print("INVALID ROI file.")
                s3_client.delete_object(Bucket=s3_Bucket_Name, Key=s3_File_Name)
                print("file deleted")

        # delete file from tmp dir
        print("remove tmp file")
        os.remove(tmp_file)

    except Exception as err:
        print(err)
        return None

    if valid_file:
        return {
            "statusCode": 200,
            "body": json.dumps(f"{s3_File_Name} validated"),
        }
    else:
        return {
            "statusCode": 200,
            "body": json.dumps(f"{s3_File_Name} failed parsing. Deleted"),
        }

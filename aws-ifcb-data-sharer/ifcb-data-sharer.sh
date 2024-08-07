#!/bin/bash
export AWS_PROFILE=ifcb-data-sharer
# sync local directory with files in S3, delete any files that don't match as well
aws s3 sync s3://ifcb-data-sharer.files /opt/ifcbdb/ifcbdb/ifcb_data/primary/ifcb-data-sharer  \
    --delete
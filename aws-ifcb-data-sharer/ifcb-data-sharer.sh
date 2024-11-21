#!/bin/bash
export AWS_PROFILE=ifcb-data-sharer
# sync local directory with files in S3, delete any files that don't match as well
aws s3 sync s3://ifcb-data-sharer.files /opt/ifcbdb/ifcbdb/ifcb_data/primary/ifcb-data-sharer  \
    --delete

# get a list of all dataset names to run operations on 
datasets=$(find /opt/ifcbdb/ifcbdb/ifcb_data/primary/ifcb-data-sharer -mindepth 2 -maxdepth 2 -type d  \( ! -iname ".*" \))

# loop through datasets, split string to last element
for i in $datasets; do
    # split directory name into user and dataset title, concat to use as the unique id
    dataset_id=$(echo "$i" | awk '{split($0,a,"/"); new_var=a[8]"_"a[9]; print new_var}')
    # use last directory string elemement for title
    dataset_title=$(echo "$i" | awk -F\/ '{print $NF}')
    # replace any underscores with space to format title
    dataset_formatted_title=${dataset_title/"_"/" "}
    # get user from directory string, second to last
    user=$(echo "$i" | awk '{split($0,a,"/"); print a[8]}')
    echo $dataset_id
    echo $dataset_title
    echo $dataset_formatted_title
    echo $user
    # add dataset to ifcbdb
    echo "add dataset to ifcbdb"
    docker exec -it ifcbdb_ifcbdb_1 python manage.py createdataset -t $dataset_formatted_title $dataset_id

    # set its data directory
    echo "set dataset's data directory"
    docker exec -it ifcbdb_ifcbdb_1 python manage.py adddirectory -k raw /data/primary/ifcb-data-sharer/$user/$dataset_title $dataset_id

    # import metadata if exists
    echo "import metadata if exists"
    docker exec -it ifcbdb_ifcbdb_1 python manage.py importmetadata /data/primary/ifcb-data-sharer/$user/$dataset_title/metadatafile.csv

    # sync ifcb data
    echo "sync ifcb data"
    docker exec -it ifcbdb_ifcbdb_1 python manage.py syncdataset $dataset_id

done
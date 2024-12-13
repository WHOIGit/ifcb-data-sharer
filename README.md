# IFCB Data Sharer

The IFCB Data Sharer allows multiple end users to share their IFCB data to the WHOI HABON IFCB dashboard using an automated Linux script.

Once installed and executed, the `ifcb-file-watcher` script will continuosly monitor a specified data directory on the IFCB device. Any new files created by the IFCB will by automatically uploaded to an AWS data pipeline that will save them to habon-ifcb.whoi.edu

## How to install

1. Contact mbrosnahan@whoi.edu to request a user account and receive access credentials
2. Install the `ifcb-file-watcher` script on to your IFCB device:

```
cd /home/your-user
curl -OL https://github.com/WHOIGit/ifcb-data-sharer/raw/refs/heads/main/ifcb-file-watcher
chmod +x ifcb-file-watcher
```

3. Create a new `.env` file in the same directory and add your AWS Key/Secret that you received from WHOI. Copy the example code from the `.env.example`:

```
AWS_ACCESS_KEY_ID=your-key-here
AWS_SECRET_ACCESS_KEY=your-secret-here
```

4. Your directory structure should look like:

```
ifcb-file-watcher
.env
```

## How to use

The `ifcb-file-watcher` script requires the following arguments:

- Directory to watch - This is the absolute or relative path to the root of the data directory for the IFCB files: `/home/user/ifcb-data`
- User name - The user name provided to you by WHOI: `my-user-name`
- Dataset name - The name of the dataset you want to add these files to on the IFCB Dashboard: `my-dataset`

```
./ifcb-file-watcher
Usage: ifcb-file-watcher <directory_to_watch> <user_name> <dataset_name>
```

Once the script is executed, it will sync all existing files and then continue to monitor the specified data directory for any new files. To keep it running when you exit your terminal session you need to start it as a background process.

Use the following command to start it as a background process and save the process PID into a local file. You can use this PID to kill the ifcb-file-watcher process if needed. You can also monitor the script output in the `ifcb-file-watcher.log` file.

```
nohup ./ifcb-file-watcher data_directory user_name “dataset name” > ifcb-file-watcher.log 2>&1 & echo $! > save_pid.txt

```

## How to stop

1. Get the process PID from the `save_pid.txt` file.
2. Kill the process: `kill -9 <your-PID>`

## Optional "Sync Only" mode

If you just need to sync an existing group of data files in a directory, you can run the script with the optional `-sync-only` flag before your arguments. This operational mode will end the program after the sync is complete:

```
./ifcb-file-watcher -sync-only data_directory user_name “dataset name”
```

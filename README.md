# IFCB Data Sharer

The IFCB Data Sharer allows multiple end users to share their IFCB data to the WHOI HABON IFCB dashboard using an automated Linux script.

Once installed and executed, the `ifcb-sync` command will continuosly monitor a specified data directory on the IFCB device. Any new files created by the IFCB will by automatically uploaded to an AWS data pipeline that will save them to habon-ifcb.whoi.edu

## How to install on IFCB (Linux)

1. Contact mbrosnahan@whoi.edu to request a user account and receive access credentials
2. Install the `ifcb-sync` script on to your IFCB device (if Git is already installed, skip the first step. Directions assume that `/home/ifcb` is your user home directory. Replace this path if necessary.):

```
sudo apt install git
cd /home/ifcb
git clone https://github.com/WHOIGit/ifcb-data-sharer.git
cd ifcb-data-sharer
chmod +x ifcb-sync
sudo ln -s /home/ifcb/ifcb-data-sharer/ifcb-sync /usr/local/bin/
```

3. Create a new `.env` file in the same directory. Copy the example code from the `.env.example`:

```
cp .env.example .env
```

4. Update the .env variables to the AWS Key/AWS Secret/User Account that you received from WHOI.

```
AWS_ACCESS_KEY_ID=your-key-here
AWS_SECRET_ACCESS_KEY=your-secret-here
USER_ACCOUNT=your-user-account
```

## How to use

The `ifcb-sync` script main commands:

### ifcb-sync start <target_directory> <target_time_series>

- Start the IFCB file watcher as a background process. Once the script is started, it will sync all existing files and then continue to monitor the specified data directory for any new files. You can also monitor the script output in the `ifcb-file-watcher.log` file.

- <target_directory> - This is the absolute or relative path to the root of the data directory for the IFCB files: ex. `/home/ifcb/ifcbdata`

- <target_time_series> - The name of the time series you want to add these files to on the IFCB Dashboard: `my-dataset`

### ifcb-sync stop <target_directory|target_time_series>

- Stops running processes associated with the target directory or time series. You only need to supply one of the options.

### ifcb-sync list

- List all the existing Time Series in your account.

Optional "Sync Only" mode

### ifcb-sync sync <target_directory> <target_time_series>

If you just need to upload or sync an existing group of data files in a directory, you can run the script in "sync-only" mode. This operation will end the program after the sync is complete. It will not monitor the directory for new files.

### Notes on data syncing

The data sync is a one-way sync from your IFCB device to WHOI's cloud storage. IF you add new files to the IFCB that are not currently present in the cloud or update existing files, then those files will be uploaded and synced. However, if you delete files from the IFCB device, this WILL NOT delete those files from the cloud.

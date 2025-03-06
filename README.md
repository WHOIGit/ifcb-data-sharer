# IFCB Data Sharer

IFCB Data Sharer allows Imaging FlowCytobot ([IFCB](https://mclanelabs.com/imaging-flowcytobot/)) operator groups to share their data through an [IFCB dashboard](https://github.com/WHOIGit/ifcbdb.git) hosted at the Woods Hole Oceanographic Institution (https://habon-ifcb.whoi.edu). Depending on how it's invoked, the program either performs a one-time file synchronization or continuosly monitors a specified data directory (e.g., on either an IFCB or a Linux, Mac, or Windows OS server), uploading any new files created within the directory to habon-ifcb.whoi.edu via an AWS data pipeline.

## Installation procedure

1. Contact mbrosnahan@whoi.edu to request a user account and receive access credentials
2. Ensure that Git is installed on your host.   

#### Linux
In a terminal:
```
sudo apt install git
```
#### Mac
Download and install Xcode through the [Mac App store](https://apps.apple.com/us/app/xcode)

#### Windows
Download and install [Git for Windows](https://git-scm.com/download/win). During installation, be sure to enable symbolic links.

3. Install the `ifcb-sync` script.

#### IFCB host
In a terminal:
```
cd /home/ifcb
git clone https://github.com/WHOIGit/ifcb-data-sharer.git
cd ifcb-data-sharer
chmod +x ifcb-sync
sudo ln -s /home/ifcb/ifcb-data-sharer/ifcb-sync /usr/local/bin/
```

#### Linux or MacOS server
In a terminal:
```
INSTALLDIR=$(pwd)
git clone https://github.com/WHOIGit/ifcb-data-sharer.git
cd ifcb-data-sharer
chmod +x ifcb-sync
sudo ln -s "$INSTALLDIR/ifcb-sync" /usr/local/bin/
```

#### Windows server
Open a Git Bash terminal 'as an Administrator' - right click icon in start menu > 'More' > 'Run as administrator'. In terminal:
```
git clone https://github.com/WHOIGit/ifcb-data-sharer.git
cd ifcb-data-sharer
chmod +x ifcb-sync
mkdir -p /usr/local/bin
```
Create a Windows symlink for ifcb-sync. Open and run cmd.exe as an administrator. In new cmd terminal:
```
cd C:\Program Files\Git\usr\local\bin
mklink ifcb-sync C:\path\to\ifcb-data-sharer\ifcb-sync
```
where `C:\path\to\ifcb-data-sharer` is the location where this repo was cloned. Default is `C:\Users\USERNAME\ifcb-data-sharer`. 

4. Create a new `.env` file in the same directory. In a terminal, copy the example code from the `.env.example`. Use Git Bash terminal if installing on a Windows host.

```
cp .env.example .env
```

5. Update the .env variables to the AWS Key/AWS Secret/User Account that you received from WHOI using a text editor (e.g., `nano .env`).

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

- <target_time_series> - The name of the time series you want to add these files to on the IFCB Dashboard: `my-dataset`. The name should not include spaces and must be less than 54 characters long.

### ifcb-sync stop <target_directory|target_time_series>

- Stops running processes associated with the target directory or time series. You only need to supply one of the options.

### ifcb-sync list

- List all the existing Time Series in your account.

## One-time sync option

### ifcb-sync sync <target_directory> <target_time_series>

If you just need to upload or sync an existing group of data files in a directory, you can run the script in "sync-only" mode. This operation will end the program after the sync is complete. It will not monitor the directory for new files.

### Notes on data syncing

The data sync is a one-way sync from your IFCB device to WHOI's cloud storage. Files in the target directory that are not already present in the cloud will be uploaded and synced. However, if files are removed or deleted from the target directory, these chagnes are not propgated to the time series on https://habon-ifcb.whoi.edu. Updates to the published time series need to be made through log into the IFCB dashboard website.

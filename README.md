# IFCB Data Sharer

The IFCB Data Sharer allows multiple end users to share their IFCB data to the WHOI HABON IFCB dashboard using an automated Linux script.

## How to use

1. Contact mbrosnahan@whoi.edu to request a user account and receive access credentials
2. Install the `ifcb-file-watcher` script on to your IFCB device:

```
cd /your-working-dir/
curl -OL https://github.com/WHOIGit/ifcb-data-sharer/raw/refs/heads/main/ifcb-file-watcher
chmod +x ifcb-file-watcher
```

3. Create a `.env` file in the same directory and add your AWS Key/Secret that you received from WHOI. Copy the example code from the `.env.example`:

```
AWS_ACCESS_KEY_ID=your-key-here
AWS_SECRET_ACCESS_KEY=your-secret-here
```

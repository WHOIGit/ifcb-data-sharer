package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/seqsense/s3sync"
)

// use godotenv package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}
	return os.Getenv(key)
}

func removeSpecialCharacters(str string) string {
	// replace any spaces with underscores
	newString := strings.ReplaceAll(str, " ", "_")

	// Define a regular expression that matches all characters except a-z, A-Z, 0-9, hyphen (-), and underscore (_)
	reg, err := regexp.Compile("[^a-zA-Z0-9-_]+")
	if err != nil {
		fmt.Println(err)
	}

	// Replace all occurrences of the pattern with an empty string
	cleanString := reg.ReplaceAllString(newString, "")
	return cleanString
}

// UploadFileToS3 uploads a file to an S3 bucket
func UploadFileToS3(awsRegion, bucketName, filePath string, dirToWatch string, userName string, datasetName string) error {
	// Create a new session using the default AWS profile or environment variables
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get the file info (to get file size, etc.)
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Check if it's a file or something else
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("It's not a regular file (could be a directory or something else)")
	}

	// set S3 key name using full file path except for the dirToWatch parent directories
	// handle dirToWatch that uses relative pathname in same pwd
	dirToWatch, _ = strings.CutPrefix(dirToWatch, "./")

	// check if dirToWatch arg included a end / or not to create clean S3 key name
	if strings.HasSuffix(dirToWatch, "/") {
		fmt.Println("The string ends with a '/', slice it off")
		dirToWatch = dirToWatch[:len(dirToWatch)-1]
	}
	fmt.Println("dirToWatch:", dirToWatch)
	bucketPath := userName + "/" + datasetName
	keyName := strings.Replace(filePath, dirToWatch, bucketPath, 1)
	fmt.Println("S3 keyName:", keyName)

	// Create S3 service client
	svc := s3.New(sess)

	// Upload the file to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(keyName),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// UploadFileToS3 uploads a file to an S3 bucket
func CheckTimeSeriesExists(awsRegion, bucketName, userName string, datasetName string) bool {
	// Create a new session using the default AWS profile or environment variables
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		fmt.Println(err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucketName), Prefix: aws.String(userName + "/"), Delimiter: aws.String("/")})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.CommonPrefixes)
	fmt.Println("Found", len(resp.Contents), "items in bucket", bucketName)

	for index, value := range resp.CommonPrefixes {
		fmt.Println("dataset:", datasetName)
		fmt.Println("Index:", index, "Value:", value.String())
		tsExists := strings.Contains(value.String(), datasetName)
		fmt.Println("TS exists:", tsExists)
		// return and break the loop if dataset found
		if tsExists {
			return tsExists
		}
	}
	return false
}

// askForConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func main() {
	// set optional sync-only flag
	syncOnly := flag.Bool("sync-only", false, "One time operation to only run the Sync operation on existing files")
	// set optional check for existing times series name
	checkTimeSeries := flag.Bool("check-time-series", false, "Whether to run a confirmation check on time series name")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("Usage: ifcb-file-watcher <directory_to_watch> <dataset_name>")
		os.Exit(1)
	}
	awsRegion := "us-east-1"               // Replace with your AWS region
	bucketName := "ifcb-data-sharer.files" // Replace with your S3 bucket name

	dirToWatch := flag.Arg(0)
	fullDatasetName := flag.Arg(1)
	datasetName := removeSpecialCharacters(fullDatasetName)

	// load .env file
	err := godotenv.Load(".env")
	aws_key := os.Getenv("AWS_ACCESS_KEY_ID")
	userName := os.Getenv("USER_ACCOUNT")
	fmt.Println("userName", userName)
	fmt.Println("AWS Key", aws_key)

	if err != nil {
		fmt.Println("Error loading .env file. You need to put your AWS access key/secret key in .env file")
		os.Exit(1)
	}

	// optional check if times series exists
	if *checkTimeSeries {
		res := CheckTimeSeriesExists(awsRegion, bucketName, userName, datasetName)
		fmt.Println("Check response", res)

		confirm := askForConfirmation("This time series already exists in your account. Please confirm that you want to use an existing account.")
		if !confirm {
			fmt.Println("Request canceled.")
			return
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// fmt.Println("Event:", event)

				if event.Op&fsnotify.Create == fsnotify.Create {
					fi, err := os.Stat(event.Name)
					if err == nil && fi.IsDir() {
						err = watcher.Add(event.Name)
						if err != nil {
							log.Println("Error adding directory:", err)
						}
					}
					fmt.Println("File created:", event.Name)

					// new file added, upload to AWS
					err = UploadFileToS3(awsRegion, bucketName, event.Name, dirToWatch, userName, datasetName)
					if err != nil {
						fmt.Println("Error uploading file:", err)
					} else {
						fmt.Println("Successfully uploaded file to S3!")
					}
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("File modified:", event.Name)
					// file modified, upload to AWS
					err = UploadFileToS3(awsRegion, bucketName, event.Name, dirToWatch, userName, datasetName)
					if err != nil {
						fmt.Println("Error uploading file:", err)
					} else {
						fmt.Println("Successfully uploaded file to S3!")
					}
				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Println("File removed:", event.Name)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					fmt.Println("File renamed:", event.Name)
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					fmt.Println("File permissions changed:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	// Walk the directory tree and add each directory to the watcher
	if !*syncOnly {
		err = filepath.Walk(dirToWatch, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				err = watcher.Add(path)
				if err != nil {
					return err
				}
				fmt.Println("Watching directory:", path)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Watching directory tree: %s\n", dirToWatch)
	}

	// Sync any existing files to AWS
	//
	// Create a new session using the default AWS profile or environment variables
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		fmt.Println("error creating session:", err)
	}

	syncManager := s3sync.New(sess)

	// Sync from local to s3
	if strings.HasSuffix(dirToWatch, "/") {
		fmt.Println("The string ends with a '/', slice it off")
		dirToWatch = dirToWatch[:len(dirToWatch)-1]
	}

	bucketSyncPath := "s3://" + bucketName + "/" + userName + "/" + datasetName
	fmt.Println("Sync to Bucket:", bucketSyncPath)
	syncManager.Sync(dirToWatch, bucketSyncPath)
	fmt.Println("Sync Complete", bucketSyncPath)

	if *syncOnly {
		// exit the program if only syncing
		os.Exit(0)
	}
	<-done
}

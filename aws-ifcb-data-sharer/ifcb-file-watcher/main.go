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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"github.com/seqsense/s3sync"
)

func main() {
	// set optional sync-only flag
	syncOnly := flag.Bool("sync-only", false, "One time operation to only run the Sync operation on existing files")
	// set optional check for existing times series name
	checkTimeSeries := flag.Bool("check-time-series", false, "Whether to run a confirmation check on time series name")
	// return a list of existing time series
	listTimeSeries := flag.Bool("list", false, "List existing time series for this user")

	// stampFile := flag.String("stamp", filepath.Join(os.Getenv("PWD"), ".last_upload"), "path to stamp file")
	flag.Parse()

	dirToWatch := flag.Arg(0)
	fullDatasetName := flag.Arg(1)
	datasetName := removeSpecialCharacters(fullDatasetName)
	stampFile := ".last_upload_" + datasetName

	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file. You need to put your AWS access key/secret key in .env file")
		os.Exit(1)
	}

	userName := os.Getenv("USER_ACCOUNT")
	awsRegion := "us-east-1" // Replace with your AWS region
	//bucketName := "ifcb-data-sharer.files" // Replace with your S3 bucket name
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		bucketName = "ifcb-data-sharer.files"
	}

	// handle list function, return results and exit
	if *listTimeSeries {
		res := getDataSeriesList(awsRegion, bucketName, userName)
		fmt.Println("Existing Time Series for user:", userName)
		for _, value := range res {
			fmt.Println(value)
		}
		os.Exit(0)
	}

	// optional check if times series exists
	if *checkTimeSeries {
		// returns true is time series exists
		res := checkTimeSeriesExists(awsRegion, bucketName, userName, datasetName)
		// fmt.Println("Check response", res)

		// if this time series is new, confirm that user want to continue
		if !res {
			confirm := askForConfirmation("You are creating a new Time Series. Please confirm that you want to set up a new Time Series")
			if confirm {
				fmt.Println("Request confirmed.")
				os.Exit(0)
			} else {
				fmt.Println("Request canceled.")
				os.Exit(1)
			}
		}
		fmt.Println("Existing Time Series. Start process")
		os.Exit(0)
	}

	interval := 60 * time.Second       // X seconds
	ticker := time.NewTicker(interval) // create a ticker
	defer ticker.Stop()                // ensure ticker is stopped when main exits

	//  run the loop in its own goroutine:
	done := make(chan bool)
	if !*syncOnly {
		go func() {
			for t := range ticker.C {
				fmt.Println("received tick at", t)
				uploadNewFiles(stampFile, awsRegion, bucketName, dirToWatch, userName, datasetName)
			}
		}()
	}

	if *syncOnly {
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
			//fmt.Println("The string ends with a '/', slice it off")
			dirToWatch = dirToWatch[:len(dirToWatch)-1]
		}

		bucketSyncPath := "s3://" + bucketName + "/" + userName + "/" + datasetName
		fmt.Println("Sync from Dir:", dirToWatch)
		fmt.Println("Sync to Bucket:", bucketSyncPath)
		err = syncManager.Sync(dirToWatch, bucketSyncPath)
		if err != nil {
			panic(err)
		}
		fmt.Println("Sync Complete", bucketSyncPath)
		// exit the program if only syncing
		os.Exit(0)
	}
	<-done
}

// main file watcher function to run in a timed loop
func uploadNewFiles(stampFile string, awsRegion, bucketName, dirToWatch string, userName string, datasetName string) {
	// ensure stamp file exists
	if _, err := os.Stat(stampFile); os.IsNotExist(err) {
		// create with zero time or current time. we'll treat as current to avoid bulk first run.
		log.Printf("creating stamp file.")
		f, err := os.Create(stampFile)
		if err != nil {
			log.Fatalf("creating stamp file: %v", err)
		}
		f.Close()
		now := time.Now()
		os.Chtimes(stampFile, now, now)
	}

	// read stamp file modtime
	info, err := os.Stat(stampFile)
	if err != nil {
		log.Fatalf("stat stamp file: %v", err)
	}
	lastRun := info.ModTime()
	log.Printf("Last upload stamp: %v\n", lastRun)

	// collect files newer than lastRun
	var toUpload []string
	err = filepath.Walk(dirToWatch, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsRegular() && fi.ModTime().After(lastRun) {
			toUpload = append(toUpload, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("walking source dir: %v", err)
	}

	if len(toUpload) == 0 {
		log.Println("No new files to upload.")
	} else {
		log.Printf("Uploading %d file(s):\n", len(toUpload))
		nowUTC := time.Now().UTC()

		for _, file := range toUpload {
			fmt.Println(" →", file)
			// new file added, upload to AWS
			err = UploadFileToS3(awsRegion, bucketName, file, dirToWatch, userName, datasetName)
			if err != nil {
				fmt.Println(nowUTC.Format("2006-01-02 15:04:05"), "Error uploading file:", err)
			} else {
				fmt.Println(nowUTC.Format("2006-01-02 15:04:05"), "Successfully uploaded file to S3!")
			}
		}
		// bump stamp to now
		now := time.Now()
		if err := os.Chtimes(stampFile, now, now); err != nil {
			log.Printf("WARNING: failed to update stamp mtime: %v", err)
		} else {
			log.Printf("Updated stamp to %v", now)
		}
	}
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
		return fmt.Errorf("not a regular file (could be a directory or something else)")
	}

	// set S3 key name using full file path except for the dirToWatch parent directories
	// replace any Windows file paths
	// handle dirToWatch that uses relative pathname in same pwd
	dirToWatch = filepath.ToSlash(dirToWatch)
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

// Check if given Time Series name exists
func checkTimeSeriesExists(awsRegion, bucketName, userName string, datasetName string) bool {
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

	//fmt.Println(resp.CommonPrefixes)
	//fmt.Println("Found", len(resp.Contents), "items in bucket", bucketName)

	for _, value := range resp.CommonPrefixes {
		//fmt.Println("dataset:", datasetName)
		tsExists := strings.Contains(aws.StringValue(value.Prefix), datasetName)
		//fmt.Println("TS exists:", tsExists)
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

// Get list of existing Time Series
func getDataSeriesList(awsRegion, bucketName, userName string) []string {
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

	//fmt.Println(resp.CommonPrefixes)
	//fmt.Println("Found", len(resp.Contents), "items in bucket", bucketName)

	var datasetsSlice []string

	for _, value := range resp.CommonPrefixes {
		arrayOfString := strings.Split(aws.StringValue(value.Prefix), "/")
		//fmt.Println("Array:", arrayOfString)
		//fmt.Println(arrayOfString[1])
		datasetsSlice = append(datasetsSlice, arrayOfString[1])

	}
	//datasetString := strings.Join(datasetsSlice, " ")
	//fmt.Println(datasetString)
	return datasetsSlice
}

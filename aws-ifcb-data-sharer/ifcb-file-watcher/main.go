package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fsnotify/fsnotify"
)

// UploadFileToS3 uploads a file to an S3 bucket
func UploadFileToS3(awsRegion, bucketName, filePath string, dirToWatch string, dashboardName string) error {
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

	// set S3 key name using full file path except for the dirToWatch parent directories
	// check if dirToWatch arg included a end / or not to create clean S3 key name
	if strings.HasSuffix(dirToWatch, "/") {
		fmt.Println("The string ends with a '/', slice it off")
		dirToWatch = dirToWatch[:len(dirToWatch)-1]
	}

	keyName := strings.Replace(filePath, dirToWatch, dashboardName, 1)
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

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: file-watcher <directory_to_watch> <dashboard_name>")
		os.Exit(1)
	}
	awsRegion := "us-east-1"               // Replace with your AWS region
	bucketName := "ifcb-data-sharer.files" // Replace with your S3 bucket name

	dirToWatch := os.Args[1]
	dashboardName := os.Args[2]

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
				fmt.Println("Event:", event)
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
					err = UploadFileToS3(awsRegion, bucketName, event.Name, dirToWatch, dashboardName)
					if err != nil {
						fmt.Println("Error uploading file:", err)
					} else {
						fmt.Println("Successfully uploaded file to S3!")
					}
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("File modified:", event.Name)
					// file modified, upload to AWS
					err = UploadFileToS3(awsRegion, bucketName, event.Name, dirToWatch, dashboardName)
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
	<-done
}

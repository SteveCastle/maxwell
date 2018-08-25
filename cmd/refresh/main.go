package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/SteveCastle/maxwell"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

type OutputConfig struct {
	Type string
	Size int
}

func main() {
	//settings
	bucket := flag.String("bucket", "maxwell-go", "Bucket to connect to.")
	region := flag.String("region", "us-east-1", "Region of the bucket to connect to.")
	sourcePath := flag.String("source", "images/", "Prefix to get input objects.")
	cachePath := flag.String("cache", "maxwell-cache/", "Prefix to write output objects.")
	config := flag.String("config", "config.json", "Config file for output")

	// Open the config file and parse it to an array of OutputConfig structs.
	// We will iterate over this to create the final output files in the cache.

	file, err := ioutil.ReadFile(*config)
	if err != nil {
		fmt.Println("Error reading config file.")
		return
	}
	var outputs []OutputConfig

	err = json.Unmarshal(file, &outputs)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Initialize session, s3, uploader, and downloader.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*region)})
	if err != nil {
		fmt.Println(err)
	}
	svc := s3.New(sess)
	downloader := s3manager.NewDownloader(sess)
	uploader := s3manager.NewUploader(sess)

	// Get list of objects in bucket.
	params := &s3.ListObjectsInput{
		Bucket: aws.String(*bucket),
		Prefix: aws.String(*sourcePath),
	}
	resp, err := svc.ListObjects(params)
	if err != nil {
		fmt.Println(err)
	}
	// Iterate over objects in bucket with prefix.
	fmt.Printf("Number of objects in set: %d\n", len(resp.Contents))
	fmt.Println(resp)
	var wg sync.WaitGroup
	for _, record := range resp.Contents {
		wg.Add(1)
		go processFile(&wg, record, bucket, cachePath, outputs, downloader, uploader)
	}
	wg.Wait()
}
func processFile(wg *sync.WaitGroup, record *s3.Object, bucket *string, cachePath *string, outputs []OutputConfig, downloader *s3manager.Downloader, uploader *s3manager.Uploader) {
	defer wg.Done()
	fmt.Printf("Working on: %s\n", *record.Key)

	//Create a file to hold original image data.
	fName := fmt.Sprintf("/tmp/%s", maxwell.Basename(*record.Key))
	file, err := os.Create(fName)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}
	defer file.Close()

	// Download file and write to file.
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(*bucket),
			Key:    aws.String(*record.Key),
		})
	if err != nil {
		exitErrorf("Unable to download item %q, %v", *record.Key, err)
	}

	fmt.Println("Downloaded", numBytes, "bytes")

	// process and upload defined file sizes.
	// TODO: seperate processing and uploading and make concurrent.
	for _, o := range outputs {
		if o.Type == "square" {
			maxwell.UploadToS3(maxwell.SquareResize(file, o.Size),
				*bucket, fmt.Sprintf("%s/%s_%dw.jpg", *cachePath, maxwell.Basename(*record.Key), o.Size), uploader)
		}
		if o.Type == "svg" {
			// Create a blurred svg.
			svg := maxwell.ConvertToSVG(fName)

			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String(*bucket),
				Key: aws.String((fmt.Sprintf("%s/%s.svg",
					*cachePath,
					maxwell.Basename(*record.Key)))),
				Body:        strings.NewReader(svg),
				ContentType: aws.String("image/svg+xml"),
			})
			if err != nil {
				// Print the error and exit.
				exitErrorf("Unable to upload %q to %q, %v", *record.Key, *bucket, err)
			}
		}
	}

	fmt.Printf("Successfully uploaded %q to %q\n", *record.Key, *bucket)
}

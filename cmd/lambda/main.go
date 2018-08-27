package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SteveCastle/maxwell"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

var config = `
[
  {"type": "square", "size": 100},
  {"type": "square", "size": 400},
  {"type": "square", "size": 1200},
  {"type": "svg"}
]
`

func handler(ctx context.Context, s3Event events.S3Event) {
	var outputs []OutputConfig

	err := json.Unmarshal([]byte(config), &outputs)
	if err != nil {
		fmt.Println(err)
		return
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")})
	if err != nil {
		fmt.Println(err)
	}
	downloader := s3manager.NewDownloader(sess)
	uploader := s3manager.NewUploader(sess)
	for _, record := range s3Event.Records {

		// Parse bucket and key from record.
		r := record.S3
		bucket := r.Bucket.Name
		key := r.Object.Key
		cachePath := "/public/maxwell-cache"
		fName := "/tmp/file.jpg"
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, bucket, key)

		//Create a file for temporary access by writers.
		file, err := os.Create("/tmp/file.jpg")
		if err != nil {
			exitErrorf("Unable to open file %q, %v", err)
		}

		defer file.Close()
		numBytes, err := downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			})
		if err != nil {
			exitErrorf("Unable to download item %q, %v", key, err)
		}

		fmt.Println("Downloaded", numBytes, "bytes")

		// Upload a lit size image.
		for _, o := range outputs {
			if o.Type == "square" {
				maxwell.UploadToS3(maxwell.SquareResize(file, o.Size),
					bucket, fmt.Sprintf("%s/%s_%dw.jpg", cachePath, maxwell.Basename(key), o.Size), uploader)
			}
			if o.Type == "svg" {
				// Create a blurred svg with ConvertToSVG this is a wrapper around the awesome primitive library by fogleman.
				svg := maxwell.ConvertToSVG(fName)
				// ConvertToSvg returns an svg string so you need to create a reader from it to upload.
				_, err = uploader.Upload(&s3manager.UploadInput{
					Bucket: aws.String(bucket),
					Key: aws.String((fmt.Sprintf("%s/%s.svg",
						cachePath,
						maxwell.Basename(key)))),
					Body:        strings.NewReader(svg),
					ContentType: aws.String("image/svg+xml"),
				})
				if err != nil {
					// Print the error and exit.
					exitErrorf("Unable to upload %q to %q, %v", key, bucket, err)
				}
			}
		}

		fmt.Printf("Successfully uploaded %q to %q\n", key, bucket)
	}
}
func main() {
	lambda.Start(handler)
}

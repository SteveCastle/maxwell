package main

import (
	"context"
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

func handler(ctx context.Context, s3Event events.S3Event) {

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
		maxwell.SquareResize(file,
			1200,
			bucket,
			fmt.Sprintf("public/maxwell-cache/%s_1200w.jpg",
				maxwell.Basename(key)),
			uploader)
		// Upload a lit size image.
		maxwell.SquareResize(file,
			400,
			bucket,
			fmt.Sprintf("public/maxwell-cache/%s_400w.jpg",
				maxwell.Basename(key)),
			uploader)

		maxwell.SquareResize(file,
			100,
			bucket,
			fmt.Sprintf("public/maxwell-cache/%s_100w.jpg",
				maxwell.Basename(key)),
			uploader)
		// Upload a gallery thumbnail.
		// Upload the SVG
		svg := maxwell.ConvertToSVG("/tmp/file.jpg")
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String((fmt.Sprintf("public/maxwell-cache/%s.svg", maxwell.Basename(key)))),
			Body:        strings.NewReader(svg),
			ContentType: aws.String("image/svg+xml"),
		})
		if err != nil {
			// Print the error and exit.
			exitErrorf("Unable to upload %q to %q, %v", key, bucket, err)
		}

		fmt.Printf("Successfully uploaded %q to %q\n", key, bucket)
	}
}
func main() {
	lambda.Start(handler)
}

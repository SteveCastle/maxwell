package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

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

	for _, record := range s3Event.Records {

		recordS3 := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, recordS3.Bucket.Name, recordS3.Object.Key)

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1")})
		if err != nil {
			fmt.Println(err)
		}
		downloader := s3manager.NewDownloader(sess)
		buff := &aws.WriteAtBuffer{}
		numBytes, err := downloader.Download(buff,
			&s3.GetObjectInput{
				Bucket: aws.String(recordS3.Bucket.Name),
				Key:    aws.String(recordS3.Object.Key),
			})
		if err != nil {
			exitErrorf("Unable to download item %q, %v", recordS3.Object.Key, err)
		}

		fmt.Println("Downloaded", numBytes, "bytes")

		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(recordS3.Bucket.Name),
			Key:         aws.String(("test.jpg")),
			Body:        bytes.NewReader(buff.Bytes()),
			ContentType: aws.String("image/jpeg"),
		})
		if err != nil {
			// Print the error and exit.
			exitErrorf("Unable to upload %q to %q, %v", recordS3.Object.Key, recordS3.Bucket.Name, err)
		}

		fmt.Printf("Successfully uploaded %q to %q\n", recordS3.Object.Key, recordS3.Bucket.Name)
	}
}
func main() {
	lambda.Start(handler)
}

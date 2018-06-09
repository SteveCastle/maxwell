package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"path"
	"strings"

	"github.com/SteveCastle/maxwell"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/nfnt/resize"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
func Basename(s string) string {
	f := path.Base(s)
	n := strings.LastIndexByte(f, '.')
	if n >= 0 {
		return f[:n]
	}
	return s
}
func simpleResize(file *os.File, width uint, bucket string, k string, s *session.Session) {
	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	b := &bytes.Buffer{}
	m := resize.Resize(width, 0, img, resize.Lanczos3)
	jpeg.Encode(b, m, nil)

	uploader := s3manager.NewUploader(s)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(k),
		Body:        bytes.NewReader(b.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	file.Seek(0, 0)
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
		//Create a file for temporary access by writers.
		file, err := os.Create("/tmp/file.jpg")
		if err != nil {
			exitErrorf("Unable to open file %q, %v", err)
		}

		defer file.Close()
		numBytes, err := downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(recordS3.Bucket.Name),
				Key:    aws.String(recordS3.Object.Key),
			})
		if err != nil {
			exitErrorf("Unable to download item %q, %v", recordS3.Object.Key, err)
		}

		fmt.Println("Downloaded", numBytes, "bytes")
		// Upload a lit size image.
		simpleResize(file, 400, recordS3.Bucket.Name, fmt.Sprintf("/maxwell-cache/%s_400w.jpg", Basename(recordS3.Object.Key)), sess)

		simpleResize(file, 100, recordS3.Bucket.Name, fmt.Sprintf("/maxwell-cache/%s_100w.jpg", Basename(recordS3.Object.Key)), sess)
		// Upload a gallery thumbnail.
		// Upload the SVG
		svg := maxwell.ConvertToSVG("/tmp/file.jpg")
		uploader := s3manager.NewUploader(sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(recordS3.Bucket.Name),
			Key:         aws.String((fmt.Sprintf("/maxwell-cache/%s.svg", Basename(recordS3.Object.Key)))),
			Body:        strings.NewReader(svg),
			ContentType: aws.String("image/svg+xml"),
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

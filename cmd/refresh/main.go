package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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

func main() {

	//settings
	bucket := flag.String("bucket", "notable-vegas", "Bucket to connect to")
	sourcePath := flag.String("source", "public/media", "Prefix to get input objects.")
	cachePath := flag.String("cache", "public/maxwell-cache", "Prefix to write output objects.")

	// Initialize session, s3, uploader, and downloader.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")})
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
	resp, _ := svc.ListObjects(params)

	// Iterate over objects in bucket with prefix.
	fmt.Printf("Number of objects in set:, %d\n", len(resp.Contents))
	for _, record := range resp.Contents {
		fmt.Printf("Working on: %s\n", *record.Key)

		//Create a file to hold original image data.
		file, err := os.Create("/tmp/file.jpg")
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
		// TODO: iterate defined sizes from flags.
		maxwell.SquareResize(file,
			1200,
			*bucket,
			fmt.Sprintf("%s/%s_1200w.jpg",
				*cachePath,
				maxwell.Basename(*record.Key)),
			uploader)
		maxwell.SquareResize(file,
			400,
			*bucket,
			fmt.Sprintf("%s/%s_400w.jpg",
				*cachePath,
				maxwell.Basename(*record.Key)),
			uploader)

		maxwell.SquareResize(file,
			100,
			*bucket,
			fmt.Sprintf("%s/%s_100w.jpg",
				*cachePath,
				maxwell.Basename(*record.Key)),
			uploader)

		// Create a blurred svg.
		svg := maxwell.ConvertToSVG("/tmp/file.jpg")

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

		fmt.Printf("Successfully uploaded %q to %q\n", *record.Key, *bucket)
	}
}

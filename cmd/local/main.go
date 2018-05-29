package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	// Get bucket and filename from command line
	bucket := flag.String("bucket", "maxwell-go", "S3 bucket to perform operations on")
	filename := flag.String("filename", "/Users/tracer/Downloads/wbvxqpwfunz01.jpg", "File to upload")
	key := flag.String("key", "/images/target.jpg", "key for upload target on s3")
	mimeType := flag.String("mimeType", "images/jpeg", "mimeType to use for s3 upload")
	flag.Parse()
	//Open file to be uploaded
	file, err := os.Open(*filename)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}

	defer file.Close()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println(err)
	}
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(*bucket),
		Key:         aws.String(*key),
		Body:        file,
		ContentType: aws.String(*mimeType),
	})
	if err != nil {
		// Print the error and exit.
		exitErrorf("Unable to upload %q to %q, %v", *filename, *bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", *filename, *bucket)
}

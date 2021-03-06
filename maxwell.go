package maxwell

import (
	"bytes"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/SteveCastle/primitive/primitive"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/edwvee/exiffix"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

//Config
var inputSize = 102
var outputSize = 600
var workers = runtime.NumCPU()
var configs shapeConfigArray
var outputs []string
var nth = 1

type shapeConfig struct {
	Count  int
	Mode   int
	Alpha  int
	Repeat int
}

type shapeConfigArray []shapeConfig

var config = shapeConfig{
	Count:  6,
	Mode:   1,
	Alpha:  0,
	Repeat: 10,
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ConvertToSVG converts an image file on disk into a primitive SVG.
func ConvertToSVG(f string) string {

	configs = append(configs, config)
	// seed random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// determine worker count
	workers := runtime.NumCPU()

	// read input image
	primitive.Log(1, "reading %s\n", f)
	input, err := primitive.LoadImage(f)
	check(err)

	// scale down input image if needed
	size := uint(inputSize)
	if size > 0 {
		input = resize.Thumbnail(size, size, input, resize.Bilinear)
	}
	// determine background color
	bg := primitive.MakeColor(primitive.AverageImageColor(input))
	// Start Transformation
	model := primitive.NewModel(input, bg, outputSize, workers)
	primitive.Log(1, "%d: t=%.3f, score=%.6f\n", 0, 0.0, model.Score)
	start := time.Now()
	frame := 0
	primitive.Log(1, "count=%d, mode=%d, alpha=%d, repeat=%d\n",
		config.Count, config.Mode, config.Alpha, config.Repeat)

	for i := 0; i < config.Count; i++ {
		frame++

		// find optimal shape and add it to the model
		t := time.Now()
		n := model.Step(primitive.ShapeType(config.Mode), config.Alpha, config.Repeat)
		nps := primitive.NumberString(float64(n) / time.Since(t).Seconds())
		elapsed := time.Since(start).Seconds()
		primitive.Log(1, "%d: t=%.3f, score=%.6f, n=%d, n/s=%s\n", frame, elapsed, model.Score, n, nps)
		// write output image(s)
	}
	return model.SVG()
}

// SquareResize resizes an input image file as a square constrained by a maximum width.
func SquareResize(file *os.File, size int) bytes.Buffer {
	// decode jpeg into image.Image
	img, _, err := exiffix.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	// bytes to return with resized jpeg image.
	b := &bytes.Buffer{}

	// resize to size using Lanczos resampling
	m := resize.Resize(uint(size), 0, img, resize.Lanczos3)

	// and preserve aspect ratio
	croppedImg, err := cutter.Crop(m, cutter.Config{
		Width:   1,
		Height:  1,
		Mode:    cutter.Centered,
		Options: cutter.Ratio,
	})
	jpeg.Encode(b, croppedImg, nil)
	file.Seek(0, 0)
	return *b
}

// Resize resizes an input image file to a set height and width without stretching.
func Resize(file *os.File, height int, width int) bytes.Buffer {
	// decode jpeg into image.Image
	img, _, err := exiffix.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	// bytes to return with resized jpeg image.
	b := &bytes.Buffer{}

	// resize to size using Lanczos resampling
	m := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	// and preserve aspect ratio
	croppedImg, err := cutter.Crop(m, cutter.Config{
		Width:   width,
		Height:  height,
		Mode:    cutter.Centered,
		Options: cutter.Ratio,
	})
	jpeg.Encode(b, croppedImg, nil)
	file.Seek(0, 0)
	return *b
}

//UploadToS3 will upload an image buffer to s3 given a bucket, key, and an s3 uploader)
func UploadToS3(data bytes.Buffer, bucket string, k string, uploader *s3manager.Uploader) error {
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(k),
		Body:        bytes.NewReader(data.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	return err
}

// Basename returns a filename without the extension to use as the key in s3.
func Basename(s string) string {
	f := path.Base(s)
	n := strings.LastIndexByte(f, '.')
	if n >= 0 {
		return f[:n]
	}
	return s
}

//OutputConfig is a type defining a struct to contain output file configs.
type OutputConfig struct {
	Type string
	Size int
}

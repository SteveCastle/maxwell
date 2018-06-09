package maxwell

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/fogleman/primitive/primitive"
	"github.com/nfnt/resize"
)

//Config
var inputSize = 102
var outputSize = 102
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

// ConvertToSVG converts an image buffer into a primitive SVG.
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

package main

import (
	"errors"
	"bytes"
	"os/user"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"

	"github.com/disintegration/imaging"
	"github.com/esimov/pigo/core"
	"github.com/fogleman/gg"
)

var (
	usr, _ = user.Current()
	cascadeFile = usr.HomeDir + "/go/src/github.com/esimov/pigo/data/facefinder"
	minSize, maxSize int = 20, 1000
	shiftFactor, scaleFactor, iouThreshold float64 = 0.1, 1.1, 0.3
)

var dc *gg.Context

// FaceDetector struct contains Pigo face detector general settings.
type FaceDetector struct {
	classifier  *pigo.Pigo
	minSize      int
	maxSize      int
	shiftFactor  float64
	scaleFactor  float64
	iouThreshold float64
	hatImage 	 image.Image
}

// DetectionResult contains the coordinates of the detected faces and the base64 converted image.
type DetectionResult struct {
	Faces []image.Rectangle
}

// Setup and loads into *FaceDetector all files and suff needed 
func Setup() (*FaceDetector, error){
	cascadeFile, err := ioutil.ReadFile(cascadeFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	pigo := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err := pigo.Unpack(cascadeFile)
	if err != nil {
		log.Fatal("Error unpacking binary cascade file")
	}
	hat := "santa-hat-transparent-background.png"
	xmashat, err := gg.LoadImage(hat)
	if err != nil {
		log.Fatalf("Error loading xmash hat.png")
	}

	return NewFaceDetector(classifier, minSize, maxSize, shiftFactor, scaleFactor, iouThreshold, xmashat), nil
}

// XMassify adds christmas hats to input into xmas.png
func XMassify(input []byte, fd *FaceDetector, ext string) ([]byte, error){

	faces, err := fd.DetectFaces(input)
	if err != nil {
		log.Println("Detection error:", err.Error())
		return []byte{}, err
	}
	if len(faces) == 0{
		return []byte{}, errors.New("No faces detected")
	}

	return fd.DrawFaces(faces, ext), nil
}

// NewFaceDetector initialises the constructor function.
func NewFaceDetector(cf *pigo.Pigo, minSize, maxSize int, shf, scf, iou float64, hat image.Image) *FaceDetector {
	return &FaceDetector{
		classifier:  cf,
		minSize:      minSize,
		maxSize:      maxSize,
		shiftFactor:  shf,
		scaleFactor:  scf,
		iouThreshold: iou,
		hatImage:	  hat,
	}
}

// DetectFaces run the detection algorithm over the provided source image.
func (fd *FaceDetector) DetectFaces(source []byte) ([]pigo.Detection, error) {
	img, _, err := image.Decode(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}
	src := pigo.ImgToNRGBA(img)

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	dc = gg.NewContext(cols, rows)
	dc.DrawImage(src, 0, 0)

	cParams := pigo.CascadeParams{
		MinSize:     fd.minSize,
		MaxSize:     fd.maxSize,
		ShiftFactor: fd.shiftFactor,
		ScaleFactor: fd.scaleFactor,
	}

	imgParams := pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	faces := fd.classifier.RunCascade(imgParams, cParams)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = fd.classifier.ClusterDetections(faces, fd.iouThreshold)

	return faces, nil
}

// DrawFaces marks the detected faces with a circle in case isCircle is true, otherwise marks with a rectangle.
func (fd *FaceDetector) DrawFaces(faces []pigo.Detection, ext string) []byte {
	var (
		qThresh float32 = 5.0
	)

	xmashat := fd.hatImage

	for _, face := range faces {
		if face.Q > qThresh {

			changeFactor := 5
			changeScale := face.Scale / 10
			xmas := imaging.Resize(xmashat, (face.Scale*2 - changeFactor*changeScale), 0, imaging.Lanczos)
			dc.DrawImage(
				xmas,
				(face.Col-face.Scale/2)-changeFactor*changeScale,
				(face.Row - face.Scale - changeFactor*changeScale/2 -changeScale),
			)
		}
	}

	img := dc.Image()
	
	buf := new(bytes.Buffer)
	switch ext {
	case ".jpg", ".jpeg":
		jpeg.Encode(buf, img, &jpeg.Options{Quality: 100})
	case ".png":
		png.Encode(buf, img)
	}
	return buf.Bytes()
}

// InSlice check if a slice contains the string value.
func InSlice(ext string, types []string) bool {
	for _, t := range types {
		if t == ext {
			return true
		}
	}
	return false
}

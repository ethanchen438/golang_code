package main

import (
	"canvas"
	"image"
	"math"
)

// AnimateBoards takes a slice of Board objects along with a cell width parameter.
// It generates a slice of images corresponding to drawing each Board on a canvas with the given cell width.
func AnimateBoards(timePoints []Board, cellWidth int) []image.Image {

	images := make([]image.Image, len(timePoints))

	if len(timePoints) == 0 {
		panic("Error: no Board objects present in input to AnimateSystem.")
	}

	// for every universe, draw to canvas and grab the image
	for i := range timePoints {
		images[i] = timePoints[i].DrawToImage(cellWidth)
	}

	return images
}

// AnimateBoardsParallel takes a slice of Board objects along with a cell width parameter.
// It generates a slice of images by drawing each Board on a canvas with the given cell width, using parallel processing.
func AnimateBoardsParallel(timePoints []Board, cellWidth int, numProcs int) []image.Image {

	images := make([]image.Image, len(timePoints))
	finished := make(chan bool, numProcs)

	if len(timePoints) == 0 {
		panic("Error: no Board objects present in input to AnimateSystemParallel.")
	}

	numFrames := len(timePoints)
	chunkSize := numFrames / numProcs

	for i := 0; i < numProcs; i++ {
		startIndex := i * chunkSize
		endIndex := startIndex + chunkSize
		if i == numProcs-1 {
			endIndex = numFrames
		}
		go animateChunk(timePoints, startIndex, endIndex, cellWidth, images, finished)
	}

	for i := 0; i < numProcs; i++ {
		<-finished
	}
	return images
}

// animateFramesChunk draws a range of frames to the images slice
func animateChunk(timePoints []Board, start, end int, cellWidth int, images []image.Image, finished chan bool) {

	for i := start; i < end; i++ {
		images[i] = timePoints[i].DrawToImage(cellWidth)
	}
	finished <- true
}

// DrawToImage is a Board method.
// Input: an integer cellWidth
// Output: the image.Image object corresponding to drawing the board
// on a square canvas, where each cell has width cellWidth.
func (b Board) DrawToImage(cellWidth int) image.Image {
	if b == nil {
		panic("Can't Draw a nil board.")
	}
	numRows := len(b)
	canvasWidth := numRows * cellWidth

	// create a new square canvas
	c := canvas.CreateNewCanvas(canvasWidth, canvasWidth)

	darkGray := canvas.MakeColor(30, 30, 30)
	gray := canvas.MakeColor(95, 95, 95)
	lightGray := canvas.MakeColor(190, 190, 190)
	white := canvas.MakeColor(255, 255, 255)

	// create a black background
	c.SetFillColor(darkGray)
	c.ClearRect(0, 0, canvasWidth, canvasWidth)
	c.Fill()

	// range over all the bodies and draw them.
	for i := range b {
		for j := range b[i] {
			val := b[i][j]

			if val == 0 {
				c.SetFillColor(darkGray)
			} else if val == 1 {
				c.SetFillColor(gray)
			} else if val == 2 {
				c.SetFillColor(lightGray)
			} else if val == 3 {
				c.SetFillColor(white)
			} else {
				c.SetFillColor(canvas.MakeColor(255, uint8(255-math.Min(255, 40*math.Log2(float64(val-3)))), uint8(255-math.Min(255, 80*math.Log2(float64(val-3))))))
			}

			// set central coordinates
			x := float64(j*cellWidth) + float64(cellWidth)/2
			y := float64(i*cellWidth) + float64(cellWidth)/2

			scalingFactor := 0.8 // to make circle smaller

			if val > 0 {
				c.Circle(x, y, scalingFactor*float64(cellWidth)/2)
				c.Fill()
			}
		}
	}
	// we want to return an image!
	//c.SaveToPNG("sandpile.png")
	return c.GetImage()
}

// Name: Ethan Chen
// Date: 09/29/2025
 
package main

import (
	"fmt"
	"gifhelper"
	"math"
	"math/rand"
	"os"
	"strconv"
)

func main() {

	fmt.Println("Let's simulate boids!")

	// Checks to make sure we have all the arguments we need to run boids
	if len(os.Args) != 13 {
		panic("Error: incorrect number of command line arguments.\n" +
			"Usage: ./boid numBoids skyWidth initialSpeed maxBoidSpeed numGens proximity separationFactor alignmentFactor cohesionFactor timeStep canvasWidth drawingFrequency")
	}

	// Converts all the input arguments to the correct type
	numBoids, err1 := strconv.Atoi(os.Args[1])
	Check(err1)

	skyWidth, err2 := strconv.ParseFloat(os.Args[2], 64)
	Check(err2)

	initialSpeed, err3 := strconv.ParseFloat(os.Args[3], 64)
	Check(err3)

	maxBoidSpeed, err4 := strconv.ParseFloat(os.Args[4], 64)
	Check(err4)

	numGens, err5 := strconv.Atoi(os.Args[5])
	Check(err5)

	proximity, err6 := strconv.ParseFloat(os.Args[6], 64)
	Check(err6)

	separationFactor, err7 := strconv.ParseFloat(os.Args[7], 64)
	Check(err7)

	alignmentFactor, err8 := strconv.ParseFloat(os.Args[8], 64)
	Check(err8)

	cohesionFactor, err9 := strconv.ParseFloat(os.Args[9], 64)
	Check(err9)

	timeStep, err10 := strconv.ParseFloat(os.Args[10], 64)
	Check(err10)

	canvasWidth, err11 := strconv.Atoi(os.Args[11])
	Check(err11)

	drawingFrequency, err12 := strconv.Atoi(os.Args[12])
	Check(err12)

	if drawingFrequency <= 0 {
		panic("Error: nonpositive number as drawingFrequency")
	}

	// Initializes initial sky values
	var initialSky Sky
	initialSky.width = skyWidth
	initialSky.maxBoidSpeed = maxBoidSpeed
	initialSky.proximity = proximity
	initialSky.separationFactor = separationFactor
	initialSky.alignmentFactor = alignmentFactor
	initialSky.cohesionFactor = cohesionFactor

	// Initializes all the boids with a random direction based on initial velocity, a random position, and no acceleration
	for i := 0; i < numBoids; i++ {

		theta := rand.Float64() * 2 * math.Pi
		vx := initialSpeed * math.Cos(theta)
		vy := initialSpeed * math.Sin(theta)

		b := Boid{
			position:     OrderedPair{x: rand.Float64() * skyWidth, y: rand.Float64() * skyWidth},
			velocity:     OrderedPair{x: vx, y: vy},
			acceleration: OrderedPair{x: 0, y: 0},
		}
		initialSky.boids = append(initialSky.boids, b)
	}

	// Used for drawing the canvas for the gif
	config := Config{
		CanvasWidth:     canvasWidth,
		BoidSize:        5.0, 
		BoidColor:       Color{R: 255, G: 255, B: 255, A: 255},
		BackgroundColor: Color{R: 173, G: 216, B: 230},
	}

	outputFile := "output/500_boids_1_cohesion"

	fmt.Println("Command line arguments read")
	fmt.Println("Simulating boids...")
	timePoints := SimulateBoids(initialSky, numGens+1, timeStep)
	fmt.Println("Simulation complete")
	fmt.Println("Drawing boids...")
	images := AnimateSystem(timePoints, config, drawingFrequency)
	fmt.Println("Images drawn")
	fmt.Println("Making GIF...")
	gifhelper.ImagesToGIF(images, outputFile)
	fmt.Println("GIF complete!")
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}


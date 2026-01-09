// Name: Ethan Chen
// Date: 10/26/25

package main

import (
	"fmt"
	"gifhelper"
	"os"
)

func main() {

	command := os.Args[1]

	var initialUniverse *Universe
	var numGens int
	var time, theta float64
	var width float64
	var scalingFactor float64

	switch command {

	case "jupiter":
		fmt.Println("Running Jupiter moon simulation...")
		filename := "jupiterMoons.txt"
		u, err := ReadUniverse(filename)
		if err != nil {
			panic(fmt.Sprintf("Failed to read jupiter: %v", err))
		}
		initialUniverse = &u

		numGens = 50000
		time = 7
		theta = 0.5
		scalingFactor = 5

	case "galaxy":
		fmt.Println("Running galaxy simulation...")
		g := InitializeGalaxy(500, 4e21, 5e22, 5e22)
		width = 1.0e23
		initialUniverse = InitializeUniverse([]Galaxy{g}, width)
		numGens = 100000
		time = 2e14
		theta = 0.5
		scalingFactor = 1e11

	case "collision":
		fmt.Println("Running galaxy collision simulation...")

		g0 := InitializeGalaxy(500, 4e21, 2e22, 5e22)   
		g1 := InitializeGalaxy(500, 4e21, 4e22, 5.2e22)
		Push(g0, +1e3, 0) 
		Push(g1, -1e3, 0) 
		width = 1.0e23
		initialUniverse = InitializeUniverse([]Galaxy{g0, g1}, width)
		numGens = 100000
		time = 2e14
		theta = 0.5
		scalingFactor = 1e11

	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Valid options: jupiter, galaxy, collision")
		return
	}

	// Run the Barnes–Hut simulation
	fmt.Println("Simulating with Barnes–Hut algorithm...")
	timePoints := BarnesHut(initialUniverse, numGens, time, theta)

	// Draw and save as GIF
	fmt.Println("Simulation complete. Drawing frames...")
	canvasWidth := 1000
	frequency := 1000
	imageList := AnimateSystem(timePoints, canvasWidth, frequency, scalingFactor)
	gifhelper.ImagesToGIF(imageList, command)

	fmt.Printf("GIF generated successfully: %s.gif\n", command)
}

func Push(g Galaxy, vx, vy float64) {
	for _, s := range g {
		s.velocity.x += vx
		s.velocity.y += vy
	}
}

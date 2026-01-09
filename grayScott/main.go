// Name: Ethan Chen
// Date: 09/02/2025

package main

import (
	"fmt"
	"gifhelper"
)

func main() {
	numRows := 250
	numCols := 250

	initialBoard := InitializeBoard(numRows, numCols)

	frac := 0.05 // tells us what percentage of interior cells to color with predators

	// how many predator rows and columns are there?
	predRows := frac * float64(numRows)
	predCols := frac * float64(numCols)

	midRow := numRows / 2
	midCol := numCols / 2

	// a little for loop to fill predators
	for r := midRow - int(predRows/2); r < midRow+int(predRows/2); r++ {
		for c := midCol - int(predCols/2); c < midCol+int(predCols/2); c++ {
			initialBoard[r][c][1] = 1.0
		}
	}

	// make prey concentration 1 at every cell
	for i := range initialBoard {
		for j := range initialBoard[i] {
			initialBoard[i][j][0] = 1.0
		}
	}

	// let's set some parameters too
	numGens := 20000 // number of iterations
	feedRate := 0.042
	killRate := 0.101

	preyDiffusionRate := 0.2 // prey are twice as fast at running
	predatorDiffusionRate := 0.1

	// let's declare kernel
	var kernel [3][3]float64
	kernel[0][0] = .05
	kernel[0][1] = .2
	kernel[0][2] = .05
	kernel[1][0] = .2
	kernel[1][1] = -1.0
	kernel[1][2] = .2
	kernel[2][0] = .05
	kernel[2][1] = .2
	kernel[2][2] = .05

	// let's simulate Gray-Scott!
	// result will be a collection of Boards corresponding to each generation.
	boards := SimulateGrayScott(initialBoard, numGens, feedRate, killRate, preyDiffusionRate, predatorDiffusionRate, kernel)

	fmt.Println("Done with simulation!")

	// we will draw what we have generated.
	fmt.Println("Drawing boards to file.")

	//for the visualization, we are only going to draw every nth board to be more efficient
	n := 100

	cellWidth := 1 // each cell is 1 pixel

	imageList := DrawBoards(boards, cellWidth, n)
	fmt.Println("Boards drawn! Now draw GIF.")

	outFile := "Gray-Scott"
	gifhelper.ImagesToGIF(imageList, outFile) // code is given
	fmt.Println("GIF drawn!")
}

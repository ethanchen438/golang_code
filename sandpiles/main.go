// Name: Ethan Chen
// Date: 11/04/25

package main
 
import (
	"fmt"
	"gifhelper"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {
	
	if len(os.Args) != 5 {
		fmt.Println("Usage: ./sandpile boardWidth numCoins placement cellWidth")
		return
	}

	boardWidth, err1 := strconv.Atoi(os.Args[1])
	numCoins, err2 := strconv.Atoi(os.Args[2])
	placement := os.Args[3]
	cellWidth, err3 := strconv.Atoi(os.Args[4])
	numProcs := runtime.NumCPU()
	var filename string

	if err1 != nil || err2 != nil || err3 != nil || boardWidth <= 0 || numCoins <= 0 || cellWidth <= 0 {
		fmt.Println("Error: all inputs must be positive values")
		return
	}

	//Initialize the main board
	board := make(Board, boardWidth)
	for i := range board {
		board[i] = make([]int, boardWidth)
	}

	switch placement {

	case "central":
		
		filename = "sandpiles_central"
		center := boardWidth / 2
		board[center][center] = numCoins

	case "random":

		filename = "sandpiles_random"
		positions := make([][2]int, 100)
		for i := 0; i < 100; i++ {
			r := rand.Intn(boardWidth)
			c := rand.Intn(boardWidth)
			positions[i] = [2]int{r, c}
		}

		for i := 0; i < numCoins; i++ {
			position := positions[rand.Intn(100)]
			board[position[0]][position[1]]++
		}

	default:
		fmt.Println("Must be central or random")
		return
	}
	parallelBoard := copyBoard(board)

	//Serial simulation
	fmt.Printf("Running serial sandpile simulation (%s placement)\n", placement)
	startSerial := time.Now()
	timePoints := SimulateSandpiles(board)
	elapsedSerial := time.Since(startSerial)
	fmt.Printf("Serial simulation complete in %s seconds.\n", elapsedSerial)

	//Parallel simulation
	fmt.Printf("Running parallel sandpile simulation with %d cores\n", numProcs)
	startParallel := time.Now()
	parallelTimePoints := SimulateSandpilesParallel(parallelBoard, numProcs)
	elapsedParallel := time.Since(startParallel)
	fmt.Printf("Parallel simulation complete in %s seconds.\n", elapsedParallel)

	//Generate GIFs
	fmt.Println("Drawing and generating GIFs")

	serialImages := AnimateBoards(timePoints, cellWidth)
	startSerialGIF := time.Now()
	gifhelper.ImagesToGIF(serialImages, filename+"_serial")
	elapsedSerialGIF := time.Since(startSerialGIF)
	fmt.Printf("Serial GIF complete in %s seconds.\n", elapsedSerialGIF)

	parallelImages := AnimateBoardsParallel(parallelTimePoints, cellWidth, numProcs)
	startParallelGIF := time.Now()
	gifhelper.ImagesToGIF(parallelImages, filename+"_parallel")
	elapsedParallelGIF := time.Since(startParallelGIF)
	fmt.Printf("Parallel GIF complete in %s seconds.\n", elapsedParallelGIF)

	fmt.Println("GIFs generated successfully")
}


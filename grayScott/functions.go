// Name: Ethan Chen
// Date: 09/02/2025

package main

// Input: an initial Board, a number of generations, and several parameters.
// Return: a slice of boards of length numGens+1 to simulate the GrayScott model over numGens generations, using the initial board.
func SimulateGrayScott(initialBoard Board, numGens int, feedRate, killRate, preyDiffusionRate, predatorDiffusionRate float64, kernel [3][3]float64) []Board {

	boards := make([]Board, numGens+1)
	boards[0] = initialBoard

	for i := 1; i <= numGens; i++ {
		boards[i] = UpdateBoard(boards[i-1], feedRate, killRate, preyDiffusionRate, predatorDiffusionRate, kernel)
	}
	return boards
}

// Input: a Board, a number of generations, and several parameters.
// Return: the board from simulating the grayScott model for one generation according to the parameters passed.
func UpdateBoard(currentBoard Board, feedRate, killRate, preyDiffusionRate, predatorDiffusionRate float64, kernel [3][3]float64) Board {

	numRows := CountRows(currentBoard)
	numCols := CountCols(currentBoard)
	newBoard := InitializeBoard(numRows, numCols)

	for row := 0; row < numRows; row++ {
		for col := 0; col < numCols; col++ {
			newBoard[row][col] = UpdateCell(currentBoard, row, col, feedRate, killRate, preyDiffusionRate, predatorDiffusionRate, kernel)
		}
	}
	return newBoard
}

// Input: a Board with row/column values, and several parameters.
// Return: the state of the cell at this row and column in the next generation of the update cell state at given row and col values.
func UpdateCell(currentBoard Board, row, col int, feedRate, killRate, preyDiffusionRate, predatorDiffusionRate float64, kernel [3][3]float64) Cell {

	currentCell := currentBoard[row][col]
	diffusionValues := ChangeDueToDiffusion(currentBoard, row, col, preyDiffusionRate, predatorDiffusionRate, kernel)
	reactionValues := ChangeDueToReactions(currentCell, feedRate, killRate)

	return SumCells(currentCell, diffusionValues, reactionValues)
}

// Input: a series of Cells as inputs.
// Return: the total values from all cells passed.
func SumCells(cells ...Cell) Cell {

	var sum Cell = [2]float64{0, 0}

	for _, value := range cells {
		sum[0] += value[0]
		sum[1] += value[1]
	}
	return sum
}

// Input: a Cell and several parameters to simulate reaction rates.
// Return: the reaction rates from both feeding and killing.
func ChangeDueToReactions(currentCell Cell, feedRate, killRate float64) Cell {

	var change Cell = [2]float64{0, 0}
	diffusion := currentCell[0] * (currentCell[1] * currentCell[1])
	change[0] = feedRate*(1-currentCell[0]) - diffusion
	change[1] = diffusion - (killRate * currentCell[1])

	return change
}

// Input: a Board, row/column values, several parameters, and a kernel to simulate diffusion.
// Return: the diffusion rates for both prey and predator.
func ChangeDueToDiffusion(currentBoard Board, row, col int, preyDiffusionRate, predatorDiffusionRate float64, kernel [3][3]float64) Cell {

	var diffusion Cell = [2]float64{0, 0}

	for kernelRows := -1; kernelRows <= 1; kernelRows++ {
		for kernelCols := -1; kernelCols <= 1; kernelCols++ {
			finalRows := row + kernelRows
			finalCols := col + kernelCols
			if InField(currentBoard, finalRows, finalCols) {
				diffusion[0] += currentBoard[finalRows][finalCols][0] * kernel[kernelRows+1][kernelCols+1]
				diffusion[1] += currentBoard[finalRows][finalCols][1] * kernel[kernelRows+1][kernelCols+1]
			}
		}
	}
	diffusion[0] *= preyDiffusionRate
	diffusion[1] *= predatorDiffusionRate
	return diffusion
}

// Input: a Board and row/col values (i,j).
// Return: true if board[i][j] is in the board else false.
func InField(currentBoard Board, i, j int) bool {

	rows := CountRows(currentBoard)
	cols := CountCols(currentBoard)

	if i < 0 || i >= rows || j < 0 || j >= cols {
		return false
	} else {
		return true
	}
}

// Input: a Board (assumes rectangular).
// Return: the number of rows in the board.
func CountRows(currentBoard Board) int {
	return len(currentBoard)
}

// Input: a Board (assumes rectangular).
// Return: the number of columns in the board.
func CountCols(currentBoard Board) int {
	if CountRows(currentBoard) == 0 {
		panic("Error: empty board given to CountCols")
	}
	return len(currentBoard[0])
}

// Input: a number of rows and a number of columns.
// Return: a numRows * numCols Board object with all values initialized to zero.
func InitializeBoard(numRows, numCols int) Board {
	currentBoard := make(Board, numRows)
	for r := range currentBoard {
		currentBoard[r] = make([]Cell, numCols)
	}
	return currentBoard
}

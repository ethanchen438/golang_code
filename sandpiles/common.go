// Name: Ethan Chen
// Date: 11/04/25

package main

// Input: a board
// Output: a copy of the board with all the values in its rows and columns 
func copyBoard(currentBoard Board) Board {

	newBoard := make(Board, len(currentBoard))
	
	for i := range currentBoard {
		row := make([]int, len(currentBoard[i]))
		for j := range currentBoard[i] {
			row[j] = currentBoard[i][j]
		}
		newBoard[i] = row
	}
	return newBoard
}

// Input: a board and its index value in the form of a row and col 
// Output: a board with the coins dispersed in each cardinal direction if the num of coins is greater than equal to 4 
func (b Board) Topple(row, col int) {

	if b[row][col] < 4 {
		return
	}

	b[row][col] -= 4
	numRows := len(b)
	numCols := len(b[0])

	// West 
	if row > 0 {
		b[row-1][col]++
	}
	// East
	if row < numRows-1 {
		b[row+1][col]++
	}
	// South
	if col > 0 {
		b[row][col-1]++
	}
	// North
	if col < numCols-1 {
		b[row][col+1]++
	}
}

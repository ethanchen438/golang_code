// Name: Ethan Chen
// Date: 11/04/25
// High Level Discussion: Vania Halim

package main

// SimulateSandpilesParallel takes as input a Board object and the number of processors.
// It returns a slice of Board objects, corresponding to repeated topples of the input
// board until we reach stability.
func SimulateSandpilesParallel(currentBoard Board, numProcs int) []Board {

	finalBoards := make([]Board, 0)
	finalBoards = append(finalBoards, copyBoard(currentBoard))
	numRows := len(currentBoard)
	finished := make(chan bool, numProcs)
	interval := 0

	for {
		// Creates chunks to pass through a channel by divding rows by processors in your computer 
		for i := 0; i < numProcs; i++ {
			chunkSize := numRows / numProcs
			startIndex := i * chunkSize
			endIndex := startIndex + chunkSize

			if i == numProcs-1 {
				endIndex = numRows
			}
			// Passes the chunk of one processor into a function with a start and end index of the row to topple 
			go toppleChunk(currentBoard, startIndex, endIndex, finished)
		}
		// Indicates the channel and its processor is finished 
		for i := 0; i < numProcs; i++ {
			<-finished
		}
		
		interval++
		stable := true
		// Loops over each value in the board and checks if a topple needs to occur 
		// After toppling, board is possibly unstable so sets stable to false so that it runs one more iteration 
		for r := range currentBoard {
			for c := range currentBoard[r] {
				if currentBoard[r][c] >= 4 {
					stable = false
					break
				}
			}
		}
		// Only passes every 500th iteration of the board to save memory when making the gif
		if interval%500 == 0 {
			finalBoards = append(finalBoards, copyBoard(currentBoard))
		}
		// Breaks out of the loop once the board is fully stable 
		if stable {
			break
		}
	}
	finalBoards = append(finalBoards,copyBoard(currentBoard))
	return finalBoards
}

// Input: a board, the start and end index of the row we want to topple and a channel to indicate the process has finished
// Output: a board after all the topples are complete for the chunk passed through from its processor 
func toppleChunk(board Board, start, end int, finished chan bool) {

	for r := start; r < end; r++ {
		for c := range board[r] {
			if board[r][c] >= 4 {
				board.Topple(r, c)
			}
		}
	}
	finished <- true
}

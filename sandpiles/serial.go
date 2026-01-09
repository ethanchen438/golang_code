// Name: Ethan Chen
// Date: 11/04/25

package main

func SimulateSandpiles(currentBoard Board) []Board {

	finalBoards := make([]Board, 0)
	finalBoards = append(finalBoards, copyBoard(currentBoard))
	interval := 0
	// Loops over each value in the board and checks if a topple needs to occur
	// After toppling, board is possibly unstable so sets stable to false so that it runs one more iteration
	for {
		
		stable := true
		for r := range currentBoard {
			for c := range currentBoard[r] {
				if currentBoard[r][c] >= 4 {
					currentBoard.Topple(r, c)
					stable = false
				}
			}
		}
		interval++
		// Only passes every 500th iteration of the board to save memory when making the gif
		if interval%500 == 0 {
			finalBoards = append(finalBoards, copyBoard(currentBoard))
		}
		// Breaks out the loop once board is fully stable
		if stable {
			break
		}
	}
	finalBoards = append(finalBoards,copyBoard(currentBoard))
	return finalBoards
}

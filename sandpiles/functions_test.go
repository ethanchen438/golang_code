// Name: Ethan Chen
// Date: 11/04/25

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

type copyTest struct {
	board  Board
	result Board
}

type ToppleTest struct {
	i, j   int
	board  Board
	result Board
}
type simulateSandpile struct {
	board  Board
	result Board
}

func TestSerial(t *testing.T) {
	tests := readSimulateTests("Tests/simulateSerial")

	for i, test := range tests {
		finalBoards := SimulateSandpiles(test.board)
		got := finalBoards[len(finalBoards)-1]

		if !boardsEqual(got, test.result) {
			t.Errorf("Serial Simulation Test %d failed:\nGot:\n%v\nWant:\n%v",
				i, boardToString(got), boardToString(test.result))
		}
	}
}

func TestParallel(t *testing.T) {
	tests := readSimulateTests("Tests/simulateParallel")
	numProcs := runtime.NumCPU()
	for i, test := range tests {
		finalBoards := SimulateSandpilesParallel(test.board, numProcs)
		got := finalBoards[len(finalBoards)-1]

		if !boardsEqual(got, test.result) {
			t.Errorf("Parallel Simulation Test %d failed:\nGot:\n%v\nWant:\n%v",
				i, boardToString(got), boardToString(test.result))
		}
	}
}

func readSimulateTests(directory string) []simulateSandpile {
	inputFiles := readDirectory(filepath.Join(directory, "input"))
	outputFiles := readDirectory(filepath.Join(directory, "output"))

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in Serial tests")
	}

	tests := make([]simulateSandpile, len(inputFiles))
	for i := range inputFiles {
		inputPath := filepath.Join(directory, "input", inputFiles[i].Name())
		outputPath := filepath.Join(directory, "output", outputFiles[i].Name())

		tests[i] = simulateSandpile{
			board:  readBoardFromFile(inputPath),
			result: readBoardFromFile(outputPath),
		}
	}

	return tests
}

func TestCopy(t *testing.T) {

	tests := readCopyTests("Tests/copyBoard")

	for i, test := range tests {
		copied := copyBoard(test.board)

		if !boardsEqual(copied, test.result) {
			t.Errorf("Test %d failed:\nGot:\n%v\nWant:\n%v", i, boardToString(copied), boardToString(test.result))
		}

		if len(copied) > 0 && len(copied[0]) > 0 {
			copied[0][0]++
			if copied[0][0] == test.board[0][0] {
				t.Errorf("Test %d failed: copyBoard did not perform a deep copy", i)
			}
		}
	}
}

func readCopyTests(directory string) []copyTest {
	inputFiles := readDirectory(filepath.Join(directory, "input"))
	outputFiles := readDirectory(filepath.Join(directory, "output"))

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in copyBoard tests")
	}

	tests := make([]copyTest, len(inputFiles))
	for i := range inputFiles {
		inputPath := filepath.Join(directory, "input", inputFiles[i].Name())
		outputPath := filepath.Join(directory, "output", outputFiles[i].Name())

		tests[i] = copyTest{
			board:  readBoardFromFile(inputPath),
			result: readBoardFromFile(outputPath),
		}
	}
	return tests
}

func TestTopple(t *testing.T) {
	tests := readToppleTests("Tests/Topple")

	for i, test := range tests {
		test.board.Topple(test.i, test.j)
		if !boardsEqual(test.board, test.result) {
			t.Errorf("Topple Test %d failed:\nGot:\n%v\nWant:\n%v", i, boardToString(test.board), boardToString(test.result))
		}
	}
}

func TestToppleChunk(t *testing.T) {
	tests := readToppleTests("Tests/ToppleChunk")

	for i, test := range tests {
		board := copyBoard(test.board)
		finished := make(chan bool, 1)
		go toppleChunk(board, test.i, test.j, finished)
		<-finished

		if !boardsEqual(board, test.result) {
			t.Errorf("ToppleChunk Test %d failed:\nGot:\n%v\nWant:\n%v",
				i, boardToString(board), boardToString(test.result))
		}
	}
}

func readToppleTests(directory string) []ToppleTest {
	inputFiles := readDirectory(filepath.Join(directory, "input"))
	outputFiles := readDirectory(filepath.Join(directory, "output"))

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in Topple tests")
	}

	tests := make([]ToppleTest, len(inputFiles))
	for i := range inputFiles {
		tests[i] = readToppleInput(filepath.Join(directory, "input", inputFiles[i].Name()))
		tests[i].result = readBoardFromFile(filepath.Join(directory, "output", outputFiles[i].Name()))
	}

	return tests
}

func readToppleInput(path string) ToppleTest {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) < 3 {
		panic("Invalid Topple input file: must have row/col, separator, and board lines")
	}

	coords := strings.Fields(lines[0])
	if len(coords) != 2 {
		panic("Invalid coordinate line in Topple input: " + lines[0])
	}
	row, _ := strconv.Atoi(coords[0])
	col, _ := strconv.Atoi(coords[1])

	var boardLines []string
	for _, l := range lines[1:] {
		if l == "-" {
			continue
		}
		boardLines = append(boardLines, l)
	}

	board := parseBoard(boardLines)

	return ToppleTest{
		i:     row,
		j:     col,
		board: board,
	}
}

func readBoardFromFile(path string) Board {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return parseBoard(lines)
}

func parseBoard(lines []string) Board {
	board := make(Board, len(lines))
	for i, line := range lines {
		parts := strings.Fields(line)
		row := make([]int, len(parts))
		for j, val := range parts {
			num, err := strconv.Atoi(val)
			if err != nil {
				panic(err)
			}
			row[j] = num
		}
		board[i] = row
	}
	return board
}

func boardsEqual(a, b Board) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func boardToString(b Board) string {
	var sb strings.Builder
	for _, row := range b {
		for _, val := range row {
			sb.WriteString(strconv.Itoa(val))
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func readDirectory(dir string) []os.DirEntry {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	return files
}

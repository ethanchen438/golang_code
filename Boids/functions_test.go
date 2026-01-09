// Name: Ethan Chen
// Date: 09/29/2025

package main

import (
	"bufio"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Struct for computing all three forces, takes in boids, force factors, and returns a ordered pair as a result
type ForceTests struct {
	b, b2                                                       Boid
	separationFactor, cohesionFactor, alignmentFactor, distance float64
	result                                                      OrderedPair
}

// Struct for computing distance, takes in boids and returns a float64 result
type DistanceTests struct {
	b, b2  Boid
	result float64
}

// Calls readForceTests and reads force inputs relating to separation factor and outputs 
func TestSeparation(t *testing.T) {

	tests := readForceTests("Tests/Separation", "separationFactor")

	for i, test := range tests {
		result := ComputeSeparation(test.b, test.b2, test.separationFactor, test.distance)
		if result != test.result {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.result)
		}
	}
}

// Calls readForceTests and reads force inputs relating to alignment factor and outputs 
func TestAlignment(t *testing.T) {

	tests := readForceTests("Tests/Alignment", "alignmentFactor")

	for i, test := range tests {
		result := ComputeAlignment(test.b, test.alignmentFactor, test.distance)
		if result != test.result {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.result)
		}
	}
}

// Calls readForceTests and reads force inputs relating to cohesion factor and outputs 
func TestCohesion(t *testing.T) {

	tests := readForceTests("Tests/Cohesion", "cohesionFactor")

	for i, test := range tests {
		result := ComputeCohesion(test.b, test.b2, test.cohesionFactor, test.distance)
		if result != test.result {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.result)
		}
	}
}

// Calls readDistanceTests and reads distance inputs and outputs
func TestDistance(t *testing.T) {
	tests := readDistanceTests("Tests/Distance")

	for i, test := range tests {
		result := ComputeDistance(test.b, test.b2)
		if result != test.result {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.result)
		}
	}
}

func readDistanceTests(directory string) []DistanceTests {
	inputFiles := readDirectory(directory + "/input")
	outputFiles := readDirectory(directory + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch")
	}

	tests := make([]DistanceTests, len(inputFiles))
	for i := range inputFiles {
		// Read two boid positions from input
		tests[i] = readDistanceInput(directory + "/input/" + inputFiles[i].Name())
		// Read expected float64 result from output
		tests[i].result = readFloatFromFile(directory + "/output/" + outputFiles[i].Name())
	}

	return tests
}

// Reads a distance test input file with two lines: boid1 position, boid2 position
func readDistanceInput(path string) DistanceTests {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) < 2 {
		panic("Invalid distance input file: must have 2 lines")
	}

	bPos := parsePair(lines[0])
	b2Pos := parsePair(lines[1])

	return DistanceTests{
		b:  Boid{position: bPos},
		b2: Boid{position: b2Pos},
	}
}

// Read a single float from a file (for expected distance output)
func readFloatFromFile(path string) float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		panic("Invalid output file: must have one line with a float")
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(scanner.Text()), 64)
	if err != nil {
		panic(err)
	}
	return val
}

func readForceTests(directory, factorName string) []ForceTests {
	inputFiles := readDirectory(directory + "/input")
	outputFiles := readDirectory(directory + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch")
	}

	tests := make([]ForceTests, len(inputFiles))
	for i := range inputFiles {
		tests[i] = readForceInput(directory+"/input/"+inputFiles[i].Name(), factorName)
		tests[i].result = readOrderedPairFromFile(directory + "/output/" + outputFiles[i].Name())
	}
	return tests
}

func readForceInput(path, factorName string) ForceTests {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var ft ForceTests
	switch factorName {
	case "alignmentFactor":
		// Alignment: only one boid
		if len(lines) < 3 {
			panic("Invalid alignment input file: must have 3 lines")
		}
		velocity := parsePair(lines[0])
		factor, _ := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)
		distance, _ := strconv.ParseFloat(strings.TrimSpace(lines[2]), 64)

		ft = ForceTests{
			b:               Boid{velocity: velocity},
			alignmentFactor: factor,
			distance:        distance,
		}

	default:
		// Separation and cohesion for two boids
		if len(lines) < 4 {
			panic("Invalid input file: must have 4 lines")
		}
		bPos := parsePair(lines[0])
		b2Pos := parsePair(lines[1])
		factor, _ := strconv.ParseFloat(strings.TrimSpace(lines[2]), 64)
		distance, _ := strconv.ParseFloat(strings.TrimSpace(lines[3]), 64)

		ft = ForceTests{
			b:        Boid{position: bPos},
			b2:       Boid{position: b2Pos},
			distance: distance,
		}

		switch factorName {
		case "separationFactor":
			ft.separationFactor = factor
		case "cohesionFactor":
			ft.cohesionFactor = factor
		}
	}

	return ft
}

// ReadOrderedPairFromFile parses an output file into an OrderedPair
func readOrderedPairFromFile(path string) OrderedPair {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		panic("Invalid output file: must have one line with two floats")
	}

	return parsePair(scanner.Text())
}

func parsePair(line string) OrderedPair {
	parts := strings.Fields(line)
	if len(parts) != 2 {
		panic("Invalid OrderedPair line: " + line)
	}
	x, _ := strconv.ParseFloat(parts[0], 64)
	y, _ := strconv.ParseFloat(parts[1], 64)
	return OrderedPair{x: x, y: y}
}

// ReadDirectory reads in a directory and returns a slice of fs.DirEntry objects containing file info for the directory
func readDirectory(dir string) []fs.DirEntry {
	//read in all files in the given directory
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	return files
}

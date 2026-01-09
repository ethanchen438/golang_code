package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type calculateCenterOfMass struct {
	inputStars []Star
	expected   Star
}

type gForceTest struct {
	s, s2  Star
	dis    float64
	result OrderedPair
}

type DistanceTests struct {
	s, s2  Star
	result float64
}

type insertStarTest struct {
	inputStars []Star
	expected   Star
}

type velocityTest struct {
	star     Star
	oldAccel OrderedPair
	timeStep float64
	expected OrderedPair
}

type positionTest struct {
	star     Star
	oldAccel OrderedPair
	oldVel   OrderedPair
	timeStep float64
	expected OrderedPair
}

type inRangeTest struct {
	pos      OrderedPair
	q        Quadrant
	expected bool
}

type NetForceTest struct {
	root     Star
	children []Star
	sectors  []Quadrant
	currStar Star
	theta    float64
	result   OrderedPair
}

func TestCalculateNetForce(t *testing.T) {
	tests := readNetForceTests("Tests/calculateNetForce")

	for i, test := range tests {
		root := &Node{
			star: &test.root,
			sector: test.sectors[0],
		}

		for j, childStar := range test.children {
			child := &Node{
				star: &childStar,
				sector: test.sectors[j+1],
			}
			root.children = append(root.children, child)
		}
		got := CalculateNetForce(root, &test.currStar, test.theta)
		if !floatAlmostEqual(got.x, test.result.x) || !floatAlmostEqual(got.y, test.result.y) {
			t.Errorf("Test %d failed: got %+v, want %+v", i, got, test.result)
		}
	}
}

func TestComputeGravitationalForce(t *testing.T) {
	tests := readGravityTests("Tests/gForceTest")

	for i, test := range tests {
		result := computeGravitationalForce(&test.s, &test.s2, test.dis)

		if math.Abs(result.x-test.result.x) > 1e-12 ||
			math.Abs(result.y-test.result.y) > 1e-12 {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.result)
		}
	}
}

func TestCalculateCenterOfMass(t *testing.T) {

	tests := readCalculateCenterOfMassTests("Tests/calculateCenterOfMass")

	for i, test := range tests {
		node := &Node{}
		for _, s := range test.inputStars {
			child := &Node{star: &Star{
				position: OrderedPair{x: s.position.x, y: s.position.y},
				mass:     s.mass,
			}}
			node.children = append(node.children, child)

		}
		node.calculateCenterOfMass()

		if node.star == nil {
			if test.expected.mass <= 0 {
				continue
			}
			t.Errorf("Test %d failed: expected non-nil star, got nil", i)
			continue
		}
		got := node.star
		want := test.expected

		if !floatAlmostEqual(got.position.x, want.position.x) ||
			!floatAlmostEqual(got.position.y, want.position.y) ||
			!floatAlmostEqual(got.mass, want.mass) {
			t.Errorf("Test %d failed:\n  got:  (x=%.4f, y=%.4f, m=%.4f)\n  want: (x=%.4f, y=%.4f, m=%.4f)",
				i, got.position.x, got.position.y, got.mass,
				want.position.x, want.position.y, want.mass)
		}
	}
}

func TestDistance(t *testing.T) {

	tests := readDistanceTests("Tests/Distance")

	for i, test := range tests {
		result := computeDistance(&test.s, &test.s2)
		if result != test.result {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.result)
		}
	}
}

func TestInsertStar(t *testing.T) {
	tests := readInsertStarTests("Tests/insertStar")

	for i, test := range tests {
		node := &Node{
			sector: Quadrant{
				x:     0,
				y:     0,
				width: 10,
			},
		}

		for _, s := range test.inputStars {
			star := s
			node.insertStar(&star)
		}

		if node.star == nil {
			t.Errorf("Test %d failed: root node.star is nil after insertion", i)
			continue
		}

		got := node.star
		want := test.expected

		if !floatAlmostEqual(got.position.x, want.position.x) ||
			!floatAlmostEqual(got.position.y, want.position.y) ||
			!floatAlmostEqual(got.mass, want.mass) {
			t.Errorf("Test %d failed:\n  got:  (x=%.4f, y=%.4f, m=%.4f)\n  want: (x=%.4f, y=%.4f, m=%.4f)",
				i,
				got.position.x, got.position.y, got.mass,
				want.position.x, want.position.y, want.mass)
		}
	}
}

func TestUpdateVelocity(t *testing.T) {
	tests := readVelocityTests("Tests/UpdateVelocity")

	for i, test := range tests {
		result := UpdateVelocity(test.star, test.oldAccel, test.timeStep)
		if !floatAlmostEqual(result.x, test.expected.x) || !floatAlmostEqual(result.y, test.expected.y) {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.expected)
		}
	}
}

func TestInRange(t *testing.T) {
	tests := readInRangeTests("Tests/inRange")

	for i, test := range tests {
		got := inRange(test.pos, test.q)
		if got != test.expected {
			t.Errorf("Test %d failed: got %v, want %v", i, got, test.expected)
		}
	}
}

func TestUpdatePosition(t *testing.T) {
	tests := readPositionTests("Tests/UpdatePosition")

	for i, test := range tests {
		result := UpdatePosition(test.star, test.oldAccel, test.oldVel, test.timeStep)
		if !floatAlmostEqual(result.x, test.expected.x) || !floatAlmostEqual(result.y, test.expected.y) {
			t.Errorf("Test %d failed: got %+v, want %+v", i, result, test.expected)
		}
	}
}

func readNetForceInput(path string) NetForceTest {
	file, err := os.Open(path)
	if err != nil {
		// Panic is used here as per the original code's style for test setup errors
		panic(fmt.Sprintf("Error opening file %s: %v", path, err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) < 3 {
		panic("Invalid NetForce input file: must have at least 3 lines")
	}

	// Line 0: root star (x y mass) + sector info (x y width)
	rootParts := strings.Fields(lines[0])
	if len(rootParts) != 6 {
		panic("Root line must have 6 floats (x y mass x y width)")
	}
	rootStar := Star{
		position: OrderedPair{x: parseFloat(rootParts[0]), y: parseFloat(rootParts[1])},
		mass:     parseFloat(rootParts[2]),
	}
	// FIX: Parsing and saving the root sector info
	rootSector := Quadrant{x: parseFloat(rootParts[3]), y: parseFloat(rootParts[4]), width: parseFloat(rootParts[5])}

	// Last two lines: current star and theta
	currStarParts := strings.Fields(lines[len(lines)-2])
	if len(currStarParts) != 3 {
		panic("Current star line must have 3 floats (x y mass)")
	}
	currStar := Star{
		position: OrderedPair{x: parseFloat(currStarParts[0]), y: parseFloat(currStarParts[1])},
		mass:     parseFloat(currStarParts[2]),
	}

	theta, err := strconv.ParseFloat(lines[len(lines)-1], 64)
	if err != nil {
		panic(fmt.Sprintf("Error parsing theta value: %v", err))
	}

	// Middle lines: children stars + sector info
	var children []Star
	var sectors []Quadrant
	sectors = append(sectors, rootSector) // Add root sector first

	for _, line := range lines[1 : len(lines)-2] {
		parts := strings.Fields(line)
		if len(parts) != 6 {
			panic("Child line must have 6 floats (x y mass x y width)")
		}
		child := Star{
			position: OrderedPair{x: parseFloat(parts[0]), y: parseFloat(parts[1])},
			mass:     parseFloat(parts[2]),
		}
		// FIX: Parsing and saving the child sector info
		childSector := Quadrant{x: parseFloat(parts[3]), y: parseFloat(parts[4]), width: parseFloat(parts[5])}
		sectors = append(sectors, childSector)
		children = append(children, child)
	}

	return NetForceTest{
		root:     rootStar,
		children: children,
		currStar: currStar,
		theta:    theta,
		sectors:  sectors, // FIX: Return sectors
	}
}

func readNetForceTests(directory string) []NetForceTest {
    inputFiles := readDirectory(directory + "/input")
    outputFiles := readDirectory(directory + "/output")

    if len(inputFiles) != len(outputFiles) {
        panic("Input/output file count mismatch in CalculateNetForce tests")
    }

    tests := make([]NetForceTest, len(inputFiles))
    for i := range inputFiles {
        // readNetForceInput now correctly returns NetForceTest with sectors populated
        tests[i] = readNetForceInput(directory + "/input/" + inputFiles[i].Name())
        tests[i].result = readOrderedPairFromFile(directory + "/output/" + outputFiles[i].Name())
    }
    return tests
}

// Helper function
func parseFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return val
}

func readVelocityTests(dir string) []velocityTest {
	inputFiles := readDirectory(dir + "/input")
	outputFiles := readDirectory(dir + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in UpdateVelocity tests")
	}

	tests := make([]velocityTest, len(inputFiles))
	for i := range inputFiles {
		tests[i] = readVelocityInput(dir + "/input/" + inputFiles[i].Name())
		tests[i].expected = readOrderedPairFromFile(dir + "/output/" + outputFiles[i].Name())
	}
	return tests
}

func readPositionTests(dir string) []positionTest {
	inputFiles := readDirectory(dir + "/input")
	outputFiles := readDirectory(dir + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in UpdatePosition tests")
	}

	tests := make([]positionTest, len(inputFiles))
	for i := range inputFiles {
		tests[i] = readPositionInput(dir + "/input/" + inputFiles[i].Name())
		tests[i].expected = readOrderedPairFromFile(dir + "/output/" + outputFiles[i].Name())
	}
	return tests
}

func readPositionInput(path string) positionTest {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 4 {
		panic("Invalid position input file: must have 3 lines")
	}

	// Line 1: pos.x pos.y vel.x vel.y acc.x acc.y
	parts := strings.Fields(lines[0])
	if len(parts) != 6 {
		panic("Line 1 must have 6 floats: pos.x pos.y vel.x vel.y acc.x acc.y")
	}
	posX, _ := strconv.ParseFloat(parts[0], 64)
	posY, _ := strconv.ParseFloat(parts[1], 64)
	velX, _ := strconv.ParseFloat(parts[2], 64)
	velY, _ := strconv.ParseFloat(parts[3], 64)
	accX, _ := strconv.ParseFloat(parts[4], 64)
	accY, _ := strconv.ParseFloat(parts[5], 64)

	oldAcc := parsePair(lines[1])
	oldVel := parsePair(lines[2])
	timeStep, _ := strconv.ParseFloat(strings.TrimSpace(lines[2]), 64)

	return positionTest{
		star: Star{
			position:     OrderedPair{x: posX, y: posY},
			velocity:     OrderedPair{x: velX, y: velY},
			acceleration: OrderedPair{x: accX, y: accY},
		},
		oldAccel: oldAcc,
		oldVel:   oldVel,
		timeStep: timeStep,
	}
}

func readInsertStarTests(directory string) []insertStarTest {

	inputFiles := readDirectory(directory + "/input")
	outputFiles := readDirectory(directory + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in InsertStar tests")
	}

	tests := make([]insertStarTest, len(inputFiles))

	for i := range inputFiles {
		inputPath := directory + "/input/" + inputFiles[i].Name()
		outputPath := directory + "/output/" + outputFiles[i].Name()
		tests[i].inputStars = readStarArrayFromFile(inputPath)
		tests[i].expected = readStarFromFile(outputPath)
	}
	return tests
}

func readStarArrayFromFile(path string) []Star {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var stars []Star
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			panic("Invalid input line: must have 3 floats (x y mass)")
		}
		x, _ := strconv.ParseFloat(parts[0], 64)
		y, _ := strconv.ParseFloat(parts[1], 64)
		m, _ := strconv.ParseFloat(parts[2], 64)
		stars = append(stars, Star{position: OrderedPair{x: x, y: y}, mass: m})
	}

	if len(stars) == 0 {
		panic("Input file must have at least one star")
	}

	return stars
}

func readDistanceTests(directory string) []DistanceTests {
	inputFiles := readDirectory(directory + "/input")
	outputFiles := readDirectory(directory + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch")
	}

	tests := make([]DistanceTests, len(inputFiles))
	for i := range inputFiles {
		tests[i] = readDistanceInput(directory + "/input/" + inputFiles[i].Name())
		tests[i].result = readFloatFromFile(directory + "/output/" + outputFiles[i].Name())
	}

	return tests
}

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

	sPos := parsePair(lines[0])
	s2Pos := parsePair(lines[1])

	return DistanceTests{
		s:  Star{position: sPos},
		s2: Star{position: s2Pos},
	}
}

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

func readGravityTests(directory string) []gForceTest {
	inputFiles := readDirectory(directory + "/input")
	outputFiles := readDirectory(directory + "/output")

	sort.Slice(inputFiles, func(i, j int) bool { return inputFiles[i].Name() < inputFiles[j].Name() })
	sort.Slice(outputFiles, func(i, j int) bool { return outputFiles[i].Name() < outputFiles[j].Name() })

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in Gravity tests")
	}

	tests := make([]gForceTest, len(inputFiles))
	for i := range inputFiles {
		tests[i] = readGravityInput(directory + "/input/" + inputFiles[i].Name())
		tests[i].result = readOrderedPairFromFile(directory + "/output/" + outputFiles[i].Name())
	}

	return tests
}

func readGravityInput(path string) gForceTest {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) < 4 {
		panic("Invalid gravity input file: must have 4 lines")
	}

	sPos := parsePair(lines[0])
	s2Pos := parsePair(lines[1])
	dis, _ := strconv.ParseFloat(strings.TrimSpace(lines[2]), 64)
	forceMag, _ := strconv.ParseFloat(strings.TrimSpace(lines[3]), 64)

	s := Star{position: sPos, mass: 1.0}
	s2 := Star{position: s2Pos, mass: 1.0}

	dx := s2.position.x - s.position.x
	dy := s2.position.y - s.position.y
	if dis != 0 {
		forceVec := OrderedPair{
			x: forceMag * dx / dis,
			y: forceMag * dy / dis,
		}
		return gForceTest{s: s, s2: s2, dis: dis, result: forceVec}
	}

	return gForceTest{s: s, s2: s2, dis: dis, result: OrderedPair{0, 0}}
}

func readCalculateCenterOfMassTests(directory string) []calculateCenterOfMass {
	inputFiles := readDirectory(directory + "/input")
	outputFiles := readDirectory(directory + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in CenterOfMass tests")
	}

	tests := make([]calculateCenterOfMass, len(inputFiles))
	for i := range inputFiles {
		inputPath := directory + "/input/" + inputFiles[i].Name()
		outputPath := directory + "/output/" + outputFiles[i].Name()

		tests[i] = readCalculateCenterOfMassInput(inputPath)
		tests[i].expected = readStarFromFile(outputPath)
	}
	return tests
}

func readCalculateCenterOfMassInput(path string) calculateCenterOfMass {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var stars []Star
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			panic("Invalid input line: must have 3 floats (x y mass)")
		}
		x, _ := strconv.ParseFloat(parts[0], 64)
		y, _ := strconv.ParseFloat(parts[1], 64)
		m, _ := strconv.ParseFloat(parts[2], 64)
		stars = append(stars, Star{position: OrderedPair{x: x, y: y}, mass: m})
	}

	if len(stars) == 0 {
		panic("CenterOfMass input file must have at least one star")
	}

	return calculateCenterOfMass{inputStars: stars}
}

func readOrderedPairFromFile(path string) OrderedPair {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		panic("Invalid file: expected one line with two floats")
	}

	line := scanner.Text()
	parts := strings.Fields(line)
	if len(parts) != 2 {
		panic("Invalid line: expected two floats")
	}

	x, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		panic(err)
	}
	y, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		panic(err)
	}

	return OrderedPair{x: x, y: y}
}

func readDirectory(dir string) []fs.DirEntry {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	return files
}

func floatAlmostEqual(a, b float64) bool {
	const eps = 1e-4
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps
}

func parsePair(line string) OrderedPair {
	parts := strings.Fields(line)
	if len(parts) != 2 {
		panic("Invalid pair line: " + line)
	}
	x, _ := strconv.ParseFloat(parts[0], 64)
	y, _ := strconv.ParseFloat(parts[1], 64)
	return OrderedPair{x: x, y: y}
}

func readStarFromFile(path string) Star {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			panic("Invalid InsertStar output line: must have 3 floats (x y mass)")
		}
		x, err1 := strconv.ParseFloat(parts[0], 64)
		y, err2 := strconv.ParseFloat(parts[1], 64)
		m, err3 := strconv.ParseFloat(parts[2], 64)
		if err1 != nil || err2 != nil || err3 != nil {
			panic("Error parsing COM output values")
		}
		return Star{position: OrderedPair{x: x, y: y}, mass: m}
	}

	panic("Output file must contain one non-empty line")
}

// Reads all inRange tests from input/output folders
func readInRangeTests(directory string) []inRangeTest {
	inputFiles := readDirectory(directory + "/input")
	outputFiles := readDirectory(directory + "/output")

	if len(inputFiles) != len(outputFiles) {
		panic("Input and output file count mismatch in inRange tests")
	}

	tests := make([]inRangeTest, len(inputFiles))
	for i := range inputFiles {
		inputPath := directory + "/input/" + inputFiles[i].Name()
		outputPath := directory + "/output/" + outputFiles[i].Name()
		tests[i].pos, tests[i].q = readInRangeInput(inputPath)
		tests[i].expected = readBoolFromFile(outputPath)
	}

	return tests
}

// Reads a single inRange input file (pos x y, quadrant x y width)
func readInRangeInput(path string) (OrderedPair, Quadrant) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) < 2 {
		panic("Invalid inRange input file: must have 2 lines")
	}

	pos := parsePair(lines[0])
	quadParts := strings.Fields(lines[1])
	if len(quadParts) != 3 {
		panic("Invalid quadrant line: must have 3 floats (x y width)")
	}
	x, _ := strconv.ParseFloat(quadParts[0], 64)
	y, _ := strconv.ParseFloat(quadParts[1], 64)
	width, _ := strconv.ParseFloat(quadParts[2], 64)

	return pos, Quadrant{x: x, y: y, width: width}
}

// Reads expected boolean output from file
func readBoolFromFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		panic("Invalid output file: must have one line with true/false")
	}
	line := strings.TrimSpace(scanner.Text())
	return line == "true"
}

func readVelocityInput(path string) velocityTest {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 3 {
		panic("Invalid velocity input file: must have 3 lines")
	}

	parts := strings.Fields(lines[0])
	if len(parts) != 6 {
		panic("Line 1 must have 6 floats: pos.x pos.y vel.x vel.y acc.x acc.y")
	}
	posX, _ := strconv.ParseFloat(parts[0], 64)
	posY, _ := strconv.ParseFloat(parts[1], 64)
	velX, _ := strconv.ParseFloat(parts[2], 64)
	velY, _ := strconv.ParseFloat(parts[3], 64)
	accX, _ := strconv.ParseFloat(parts[4], 64)
	accY, _ := strconv.ParseFloat(parts[5], 64)

	oldAcc := parsePair(lines[1])

	timeStep, _ := strconv.ParseFloat(strings.TrimSpace(lines[2]), 64)

	return velocityTest{
		star: Star{
			position:     OrderedPair{x: posX, y: posY},
			velocity:     OrderedPair{x: velX, y: velY},
			acceleration: OrderedPair{x: accX, y: accY},
		},
		oldAccel: oldAcc,
		timeStep: timeStep,
	}
}


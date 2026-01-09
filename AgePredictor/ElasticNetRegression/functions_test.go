package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// All parsers made with AI //

// ----------------------------------------------------------------------
// AddColOfOnes()
// ----------------------------------------------------------------------

type AddColOfOnesTest struct {
	input  [][]float64
	output [][]float64
}

func TestAddColOfOnes(t *testing.T) {
	tests := ReadAddColOfOnesTests("Tests/AddColOfOnes/")

	for i, test := range tests {
		got := AddColOfOnes(test.input)

		if !matricesEqual(got, test.output) {
			t.Errorf("Test #%d failed:\nGot:  %v\nWant: %v", i, got, test.output)
		}
	}

	fmt.Println("AddColOfOnes tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadAddColOfOnesTests(basePath string) []AddColOfOnesTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, _ := os.ReadDir(inputDir)
	var tests []AddColOfOnesTest

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".txt") {
			inputPath := filepath.Join(inputDir, f.Name())
			outputPath := filepath.Join(outputDir, strings.Replace(f.Name(), "input", "output", 1))

			input := readMatrixFile(inputPath)
			output := readMatrixFile(outputPath)

			tests = append(tests, AddColOfOnesTest{
				input:  input,
				output: output,
			})
		}
	}

	return tests
}

func readMatrixFile(path string) [][]float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var matrix [][]float64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		row := make([]float64, len(parts))

		for i, p := range parts {
			val, _ := strconv.ParseFloat(p, 64)
			row[i] = val
		}

		matrix = append(matrix, row)
	}

	return matrix
}

// ----------------------------------------------------------------------
// applyNormalization()
// ----------------------------------------------------------------------

type ApplyNormalizationTest struct {
	X      [][]float64
	y      []float64
	params NormParams
	XWant  [][]float64
	yWant  []float64
}

func TestApplyNormalization(t *testing.T) {
	tests := ReadApplyNormalizationTests("Tests/applyNormalization/")

	for i, test := range tests {

		gotX, gotY := applyNormalization(test.X, test.y, test.params)

		// --- NEW: round computed outputs before comparison ---
		for r := range gotX {
			for c := range gotX[r] {
				gotX[r][c] = roundFloat(gotX[r][c], 6)
			}
		}
		for r := range gotY {
			gotY[r] = roundFloat(gotY[r], 6)
		}
		// ------------------------------------------------------

		if !matricesEqual(gotX, test.XWant) {
			t.Errorf("Test #%d failed (X):\nGot:  %v\nWant: %v", i, gotX, test.XWant)
		}

		if !vectorsEqual(gotY, test.yWant) {
			t.Errorf("Test #%d failed (y):\nGot:  %v\nWant: %v", i, gotY, test.yWant)
		}
	}

	fmt.Println("applyNormalization tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadApplyNormalizationTests(basePath string) []ApplyNormalizationTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, _ := os.ReadDir(inputDir)
	var tests []ApplyNormalizationTest

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		index := strings.TrimPrefix(f.Name(), "input")
		index = strings.TrimSuffix(index, ".txt")

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, fmt.Sprintf("output%s.txt", index))

		X, y, params := parseInputFile(inputPath)
		XWant, yWant := parseOutputFile(outputPath)

		tests = append(tests, ApplyNormalizationTest{
			X:      X,
			y:      y,
			params: params,
			XWant:  XWant,
			yWant:  yWant,
		})
	}

	return tests
}

func parseInputFile(path string) ([][]float64, []float64, NormParams) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var X [][]float64
	var y []float64
	var params NormParams

	section := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "X:":
			section = "X"
			continue
		case "Y:":
			section = "Y"
			continue
		case "MEANS:":
			section = "MEANS"
			continue
		case "STDS:":
			section = "STDS"
			continue
		case "YMEAN:":
			section = "YMEAN"
			continue
		case "YSTD:":
			section = "YSTD"
			continue
		}

		fields := strings.Fields(line)

		switch section {
		case "X":
			X = append(X, parseFloatRow(fields))

		case "Y":
			for _, f := range fields {
				y = append(y, parseFloat(f))
			}

		case "MEANS":
			for _, f := range fields {
				params.Means = append(params.Means, parseFloat(f))
			}

		case "STDS":
			for _, f := range fields {
				params.Stds = append(params.Stds, parseFloat(f))
			}

		case "YMEAN":
			params.YMean = parseFloat(fields[0])

		case "YSTD":
			params.YStd = parseFloat(fields[0])
		}
	}

	return X, y, params
}

func parseOutputFile(path string) ([][]float64, []float64) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var X [][]float64
	var y []float64

	section := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "XOUT:":
			section = "X"
			continue
		case "YOUT:":
			section = "Y"
			continue
		}

		fields := strings.Fields(line)

		if section == "X" {
			X = append(X, parseFloatRow(fields))
		} else if section == "Y" {
			for _, f := range fields {
				y = append(y, parseFloat(f))
			}
		}
	}

	return X, y
}

// ----------------------------------------------------------------------
// denormalizePrediction()
// ----------------------------------------------------------------------

type DenormTest struct {
	yNorm  float64
	params NormParams
	want   float64
}

func TestDenormalizePrediction(t *testing.T) {
	tests := ReadDenormTests("Tests/denormalizePrediction/")

	for i, test := range tests {
		got := denormalizePrediction(test.yNorm, test.params)

		if got != test.want {
			t.Errorf("Test #%d failed:\nGot:  %v\nWant: %v", i, got, test.want)
		}
	}

	fmt.Println("denormalizePrediction tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadDenormTests(basePath string) []DenormTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, _ := os.ReadDir(inputDir)
	var tests []DenormTest

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		index := strings.TrimPrefix(f.Name(), "input")
		index = strings.TrimSuffix(index, ".txt")

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, fmt.Sprintf("output%s.txt", index))

		yNorm, params := parseDenormInput(inputPath)
		want := parseDenormOutput(outputPath)

		tests = append(tests, DenormTest{
			yNorm:  yNorm,
			params: params,
			want:   want,
		})
	}

	return tests
}

func parseDenormInput(path string) (float64, NormParams) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	section := ""
	var yNorm float64
	var params NormParams

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "YNORM:":
			section = "YNORM"
			continue
		case "YMEAN:":
			section = "YMEAN"
			continue
		case "YSTD:":
			section = "YSTD"
			continue
		}

		fields := strings.Fields(line)

		switch section {
		case "YNORM":
			yNorm = parseFloat(fields[0])

		case "YMEAN":
			params.YMean = parseFloat(fields[0])

		case "YSTD":
			params.YStd = parseFloat(fields[0])
		}
	}

	return yNorm, params
}

func parseDenormOutput(path string) float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "WANT:":
			section = "WANT"
			continue
		}

		if section == "WANT" {
			fields := strings.Fields(line)
			return parseFloat(fields[0])
		}
	}

	panic("No WANT section found in output file")
}

// ----------------------------------------------------------------------
// GetFolds()
// ----------------------------------------------------------------------

/*
This function was used to help generate outputs for testing GetFolds

	func main() {
	    nSamples := //insert number here
	    indices := make([]int, nSamples)
	    for i := range indices {
	        indices[i] = i
	    }

	    r := rand.New(rand.NewSource(42))
	    r.Shuffle(nSamples, func(i, j int) {
	        indices[i], indices[j] = indices[j], indices[i]
	    })

	    fmt.Println(indices)
	}
*/

type GetFoldsTest struct {
	nSamples int
	k        int
	want     [][]int
}

func TestGetFolds(t *testing.T) {
	tests := ReadGetFoldsTests("Tests/GetFolds/")

	for i, test := range tests {
		got := getFoldsDeterministic(test.nSamples, test.k)

		if !intMatricesEqual(got, test.want) {
			t.Errorf("Test #%d failed:\nGot:  %v\nWant: %v", i, got, test.want)
		}
	}

	fmt.Println("getFolds tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadGetFoldsTests(basePath string) []GetFoldsTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, err := os.ReadDir(inputDir)
	if err != nil {
		panic(err)
	}

	var tests []GetFoldsTest
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, strings.Replace(f.Name(), "input", "output", 1))

		nSamples, k := parseGetFoldsInput(inputPath)
		want := parseGetFoldsOutput(outputPath)

		tests = append(tests, GetFoldsTest{
			nSamples: nSamples,
			k:        k,
			want:     want,
		})
	}

	return tests
}

func parseGetFoldsInput(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var nSamples, k int
	section := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "NSAMPLES:":
			section = "NSAMPLES"
			continue
		case "K:":
			section = "K"
			continue
		}

		switch section {
		case "NSAMPLES":
			nSamples, _ = strconv.Atoi(line)
		case "K":
			k, _ = strconv.Atoi(line)
			section = ""
		}
	}

	return nSamples, k
}

func parseGetFoldsOutput(path string) [][]int {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var folds [][]int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		row := make([]int, len(parts))
		for i, p := range parts {
			val, _ := strconv.Atoi(p)
			row[i] = val
		}
		folds = append(folds, row)
	}

	return folds
}

// ----------------------------------------------------------------------
// getNormParams()
// ----------------------------------------------------------------------

type GetNormParamsTest struct {
	X      [][]float64
	y      []float64
	params NormParams
}

func TestGetNormParams(t *testing.T) {
	precision := uint(6)
	tests := ReadGetNormParamsTests("Tests/getNormParams/")

	for i, test := range tests {
		got := getNormParams(test.X, test.y)

		// Round results for stable comparison
		for j := range got.Means {
			got.Means[j] = roundFloat(got.Means[j], precision)
		}
		for j := range got.Stds {
			got.Stds[j] = roundFloat(got.Stds[j], precision)
		}
		got.YMean = roundFloat(got.YMean, precision)
		got.YStd = roundFloat(got.YStd, precision)

		if !normParamsEqual(got, test.params) {
			t.Errorf("Test #%d failed:\nGot: %+v\nWant: %+v", i, got, test.params)
		}
	}

	fmt.Println("getNormParams tested!")
}

// ----------------------------------------------------------------------
// Parsers

func normParamsEqual(a, b NormParams) bool {
	if len(a.Means) != len(b.Means) || len(a.Stds) != len(b.Stds) {
		return false
	}
	for i := range a.Means {
		if a.Means[i] != b.Means[i] || a.Stds[i] != b.Stds[i] {
			return false
		}
	}
	return a.YMean == b.YMean && a.YStd == b.YStd
}

func ReadGetNormParamsTests(basePath string) []GetNormParamsTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, _ := os.ReadDir(inputDir)
	var tests []GetNormParamsTest

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		index := strings.TrimPrefix(f.Name(), "input")
		index = strings.TrimSuffix(index, ".txt")

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, fmt.Sprintf("output%s.txt", index))

		X, y := parseGetNormParamsInput(inputPath)
		params := parseGetNormParamsOutput(outputPath)

		tests = append(tests, GetNormParamsTest{
			X:      X,
			y:      y,
			params: params,
		})
	}

	return tests
}

func parseGetNormParamsInput(path string) ([][]float64, []float64) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var X [][]float64
	var y []float64
	section := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "X:":
			section = "X"
			continue
		case "Y:":
			section = "Y"
			continue
		}

		fields := strings.Fields(line)
		switch section {
		case "X":
			X = append(X, parseFloatRow(fields))
		case "Y":
			for _, f := range fields {
				y = append(y, parseFloat(f))
			}
		}
	}

	return X, y
}

func parseGetNormParamsOutput(path string) NormParams {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var params NormParams
	section := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "MEANS:":
			section = "MEANS"
			continue
		case "STDS:":
			section = "STDS"
			continue
		case "YMEAN:":
			section = "YMEAN"
			continue
		case "YSTD:":
			section = "YSTD"
			continue
		}

		fields := strings.Fields(line)
		switch section {
		case "MEANS":
			for _, f := range fields {
				params.Means = append(params.Means, parseFloat(f))
			}
		case "STDS":
			for _, f := range fields {
				params.Stds = append(params.Stds, parseFloat(f))
			}
		case "YMEAN":
			params.YMean = parseFloat(fields[0])
		case "YSTD":
			params.YStd = parseFloat(fields[0])
		}
	}

	return params
}

// ----------------------------------------------------------------------
// meanSquaredError()
// ----------------------------------------------------------------------

type MeanSquaredErrorTest struct {
	yActual    []float64
	yPredicted []float64
	want       float64
}

func TestMeanSquaredError(t *testing.T) {
	tests := ReadMeanSquaredErrorTests("Tests/MeanSquaredError/")

	for i, test := range tests {
		got := meanSquaredError(test.yActual, test.yPredicted)

		// Round both expected and got values to avoid floating point precision issues
		gotRounded := roundFloat(got, 6)
		wantRounded := roundFloat(test.want, 6)

		if gotRounded != wantRounded {
			t.Errorf("Test #%d failed: Got %v, Want %v", i, gotRounded, wantRounded)
		}
	}

	fmt.Println("meanSquaredError tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadMeanSquaredErrorTests(basePath string) []MeanSquaredErrorTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, _ := os.ReadDir(inputDir)
	var tests []MeanSquaredErrorTest

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		index := strings.TrimPrefix(f.Name(), "input")
		index = strings.TrimSuffix(index, ".txt")

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, fmt.Sprintf("output%s.txt", index))

		yActual, yPredicted := parseMeanSquaredErrorInput(inputPath)
		want := parseMeanSquaredErrorOutput(outputPath)

		tests = append(tests, MeanSquaredErrorTest{
			yActual:    yActual,
			yPredicted: yPredicted,
			want:       want,
		})
	}

	return tests
}

func parseMeanSquaredErrorInput(path string) ([]float64, []float64) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var yActual, yPredicted []float64
	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "YACTUAL:":
			section = "YACTUAL"
			continue
		case "YPREDICTED:":
			section = "YPREDICTED"
			continue
		}

		fields := strings.Fields(line)
		switch section {
		case "YACTUAL":
			for _, f := range fields {
				yActual = append(yActual, parseFloat(f))
			}
		case "YPREDICTED":
			for _, f := range fields {
				yPredicted = append(yPredicted, parseFloat(f))
			}
		}
	}

	return yActual, yPredicted
}

func parseMeanSquaredErrorOutput(path string) float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var want float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.ToUpper(line) == "MSE:" {
			continue
		}

		want = parseFloat(line)
	}

	return want
}

// ----------------------------------------------------------------------
// precalculateXTS()
// ----------------------------------------------------------------------

type PrecalculateXTSTest struct {
	Input  [][]float64
	Output []float64
}

func TestPrecalculateXTS(t *testing.T) {
	tests := ReadPrecalculateXTSTests("Tests/precalculateXTS/")

	for i, test := range tests {
		got := precalculateXTS(test.Input)

		// Round for floating-point stability
		gotRounded := make([]float64, len(got))
		for j, val := range got {
			gotRounded[j] = roundFloat(val, 6)
		}

		wantRounded := make([]float64, len(test.Output))
		for j, val := range test.Output {
			wantRounded[j] = roundFloat(val, 6)
		}

		if !vectorsEqual(gotRounded, wantRounded) {
			t.Errorf("Test #%d failed: Got %v, Want %v", i, gotRounded, wantRounded)
		}
	}

	fmt.Println("precalculateXTS tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadPrecalculateXTSTests(basePath string) []PrecalculateXTSTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, _ := os.ReadDir(inputDir)
	var tests []PrecalculateXTSTest

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		index := strings.TrimPrefix(f.Name(), "input")
		index = strings.TrimSuffix(index, ".txt")

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, fmt.Sprintf("output%s.txt", index))

		input := parseMatrixFile(inputPath)
		output := parseVectorFile(outputPath)

		tests = append(tests, PrecalculateXTSTest{
			Input:  input,
			Output: output,
		})
	}

	return tests
}

func parseMatrixFile(path string) [][]float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var matrix [][]float64
	scanner := bufio.NewScanner(file)
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.ToUpper(line) == "X:" {
			section = "X"
			continue
		}

		if section == "X" {
			parts := strings.Fields(line)
			row := make([]float64, len(parts))
			for i, p := range parts {
				row[i] = parseFloat(p)
			}
			matrix = append(matrix, row)
		}
	}

	return matrix
}

func parseVectorFile(path string) []float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var vector []float64
	section := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.ToUpper(line) == "OUTPUT:" {
			section = "OUTPUT"
			continue
		}

		if section == "OUTPUT" {
			parts := strings.Fields(line)
			for _, p := range parts {
				vector = append(vector, parseFloat(p))
			}
		}
	}

	return vector
}

// ----------------------------------------------------------------------
// predict()
// ----------------------------------------------------------------------

type PredictTest struct {
	X       [][]float64
	Weights []float64
	Output  []float64
}

func TestPredict(t *testing.T) {
	tests := ReadPredictTests("Tests/predict/")

	for i, test := range tests {
		got := predict(test.X, test.Weights)

		// Round for floating-point stability
		gotRounded := make([]float64, len(got))
		for j, val := range got {
			gotRounded[j] = roundFloat(val, 6)
		}

		wantRounded := make([]float64, len(test.Output))
		for j, val := range test.Output {
			wantRounded[j] = roundFloat(val, 6)
		}

		if !vectorsEqual(gotRounded, wantRounded) {
			t.Errorf("Test #%d failed: Got %v, Want %v", i, gotRounded, wantRounded)
		}
	}

	fmt.Println("predict tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadPredictTests(basePath string) []PredictTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, _ := os.ReadDir(inputDir)
	var tests []PredictTest

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		index := strings.TrimPrefix(f.Name(), "input")
		index = strings.TrimSuffix(index, ".txt")

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, fmt.Sprintf("output%s.txt", index))

		X, weights := parsePredictInputFile(inputPath)
		output := parseVectorFile(outputPath)

		tests = append(tests, PredictTest{
			X:       X,
			Weights: weights,
			Output:  output,
		})
	}

	return tests
}

func parsePredictInputFile(path string) ([][]float64, []float64) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var X [][]float64
	var weights []float64
	section := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "X:":
			section = "X"
			continue
		case "WEIGHTS:":
			section = "WEIGHTS"
			continue
		}

		fields := strings.Fields(line)
		switch section {
		case "X":
			row := make([]float64, len(fields))
			for i, f := range fields {
				row[i] = parseFloat(f)
			}
			X = append(X, row)
		case "WEIGHTS":
			for _, f := range fields {
				weights = append(weights, parseFloat(f))
			}
		}
	}

	return X, weights
}

// ----------------------------------------------------------------------
// softThreshold()
// ----------------------------------------------------------------------

func TestSoftThreshold(t *testing.T) {
	inputDir := "Tests/SoftThreshold/Input"
	outputDir := "Tests/SoftThreshold/Output"

	files, _ := os.ReadDir(inputDir)
	for i, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, strings.Replace(f.Name(), "input", "output", 1))

		r, a := parseSoftThresholdInput(inputPath)
		want := parseSoftThresholdOutput(outputPath)

		got := softThreshold(r, a)

		if got != want {
			t.Errorf("Test #%d failed: Got %v, Want %v", i, got, want)
		}
	}

	fmt.Println("softThreshold tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseSoftThresholdInput(path string) (float64, float64) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var r, a float64
	section := ""
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "R:":
			section = "R"
			continue
		case "A:":
			section = "A"
			continue
		}

		val := parseFloat(line)

		if section == "R" {
			r = val
		} else if section == "A" {
			a = val
		}
	}

	return r, a
}

func parseSoftThresholdOutput(path string) float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var out float64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.ToUpper(line) != "OUTPUT:" {
			out = parseFloat(line)
		}
	}

	return out
}

// ----------------------------------------------------------------------
// elasticNetRegression()
// ----------------------------------------------------------------------

func TestElasticNetRegression(t *testing.T) {
	inputDir := "Tests/ElasticNetRegression/Input"
	outputDir := "Tests/ElasticNetRegression/Output"

	files, _ := os.ReadDir(inputDir)
	for i, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		inputPath := filepath.Join(inputDir, f.Name())
		outputPath := filepath.Join(outputDir, strings.Replace(f.Name(), "input", "output", 1))

		X, y, alpha, lambda, maxIter, tol, XTS := parseElasticNetRegressionInput(inputPath)
		want := parseElasticNetRegressionOutput(outputPath)

		got := elasticNetRegression(X, y, XTS, alpha, lambda, maxIter, tol)

		// Round both got and want to 3 decimal places before comparing
		for j := range got {
			got[j] = roundFloat(got[j], 3)
			if j < len(want) {
				want[j] = roundFloat(want[j], 3)
			}
		}

		if !vectorsEqual(got, want) {
			t.Errorf("Test #%d failed:\nGot:  %v\nWant: %v", i, got, want)
		}
	}

	fmt.Println("elasticNetRegression tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseElasticNetRegressionInput(path string) (X [][]float64, y []float64, alpha, lambda float64, maxIter int, tol float64, XTS []float64) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	section := ""
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		upper := strings.ToUpper(line)
		switch upper {
		case "X:":
			section = "X"
			continue
		case "Y:":
			section = "Y"
			continue
		case "ALPHA:":
			section = "ALPHA"
			continue
		case "LAMBDA:":
			section = "LAMBDA"
			continue
		case "MAXITER:":
			section = "MAXITER"
			continue
		case "TOL:":
			section = "TOL"
			continue
		case "XTS:":
			section = "XTS"
			continue
		}

		fields := strings.Fields(line)
		switch section {
		case "X":
			X = append(X, parseFloatRow(fields))
		case "Y":
			for _, f := range fields {
				y = append(y, parseFloat(f))
			}
		case "ALPHA":
			alpha = parseFloat(fields[0])
		case "LAMBDA":
			lambda = parseFloat(fields[0])
		case "MAXITER":
			maxIter, _ = strconv.Atoi(fields[0])
		case "TOL":
			tol = parseFloat(fields[0])
		case "XTS":
			for _, f := range fields {
				XTS = append(XTS, parseFloat(f))
			}
		}
	}

	return
}

func parseElasticNetRegressionOutput(path string) []float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var weights []float64
	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.ToUpper(line) == "WEIGHTS:" {
			section = "WEIGHTS"
			continue
		}

		if section == "WEIGHTS" {
			fields := strings.Fields(line)
			for _, f := range fields {
				weights = append(weights, parseFloat(f))
			}
		}
	}

	return weights
}

// ----------------------------------------------------------------------
// crossValidationElasticNet()
// ----------------------------------------------------------------------

func TestCrossValidationElasticNetDeterministic(t *testing.T) {
	tests := ReadCrossValidationElasticNetTests("Tests/CrossValidationElasticNet/")

	for i, test := range tests {
		got := crossValidationElasticNetDeterministic(
			test.X, test.Y,
			test.Alpha, test.Lambda,
			test.K, test.MaxIter, test.Tol,
		)

		got = roundFloat(got, 3)

		if got != test.Want {
			t.Errorf("Test #%d failed:\nGot:  %v\nWant: %v", i, got, test.Want)
		}
	}

	fmt.Println("crossValidationElasticNetDeterministic tested!")
}

type CrossValidationElasticNetTest struct {
	X       [][]float64
	Y       []float64
	Alpha   float64
	Lambda  float64
	K       int
	MaxIter int
	Tol     float64
	Want    float64
}

// ----------------------------------------------------------------------
// Parsers

func ReadCrossValidationElasticNetTests(basePath string) []CrossValidationElasticNetTest {
	inputDir := filepath.Join(basePath, "Input")
	outputDir := filepath.Join(basePath, "Output")

	files, err := os.ReadDir(inputDir)
	if err != nil {
		panic(err)
	}

	var tests []CrossValidationElasticNetTest

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		inputPath := filepath.Join(inputDir, f.Name())

		// Extract the index from "input0.txt" -> "0"
		index := strings.TrimPrefix(f.Name(), "input")
		index = strings.TrimSuffix(index, ".txt")

		outputPath := filepath.Join(outputDir, fmt.Sprintf("output%s.txt", index))

		X, Y, alpha, lambda, k, maxIter, tol := parseCrossValidationElasticNetInput(inputPath)
		want := parseCrossValidationOutput(outputPath)

		tests = append(tests, CrossValidationElasticNetTest{
			X:       X,
			Y:       Y,
			Alpha:   alpha,
			Lambda:  lambda,
			K:       k,
			MaxIter: maxIter,
			Tol:     tol,
			Want:    want,
		})
	}

	return tests
}

func parseCrossValidationElasticNetInput(path string) (X [][]float64, Y []float64, alpha, lambda float64, k, maxIter int, tol float64) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "X:":
			section = "X"
			continue
		case "Y:":
			section = "Y"
			continue
		case "ALPHA:":
			section = "ALPHA"
			continue
		case "LAMBDA:":
			section = "LAMBDA"
			continue
		case "K:":
			section = "K"
			continue
		case "MAXITER:":
			section = "MAXITER"
			continue
		case "TOL:":
			section = "TOL"
			continue
		}

		fields := strings.Fields(line)
		switch section {
		case "X":
			X = append(X, parseFloatRow(fields))
		case "Y":
			for _, f := range fields {
				Y = append(Y, parseFloat(f))
			}
		case "ALPHA":
			alpha = parseFloat(fields[0])
		case "LAMBDA":
			lambda = parseFloat(fields[0])
		case "K":
			k, _ = strconv.Atoi(fields[0])
		case "MAXITER":
			maxIter, _ = strconv.Atoi(fields[0])
		case "TOL":
			tol = parseFloat(fields[0])
		}
	}

	return
}

func parseCrossValidationOutput(path string) float64 {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	section := ""
	var mse float64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToUpper(line) {
		case "MSE:":
			section = "MSE"
			continue
		}

		if section == "MSE" {
			mse, _ = strconv.ParseFloat(line, 64)
			break
		}
	}

	return mse
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseFloatRow(fields []string) []float64 {
	row := make([]float64, len(fields))
	for i, f := range fields {
		row[i] = parseFloat(f)
	}
	return row
}

func matricesEqual(a, b [][]float64) bool {
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

func vectorsEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func intMatricesEqual(a, b [][]int) bool {
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

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// All parsers made with AI //

// ----------------------------------------------------------------------
// betaCDF()
// ----------------------------------------------------------------------

type BetaCDFTest struct {
	Inputs [][3]float64
	Want   []float64
	Name   string
}

func TestBetaCDF(t *testing.T) {
	tests := ReadBetaCDFTests("Tests/betaCDF")

	if len(tests) == 0 {
		t.Fatalf("No betaCDF tests found.")
	}

	tol := 1e-5

	for ti, test := range tests {
		if len(test.Inputs) != len(test.Want) {
			t.Fatalf("Test %d (%s): input/output mismatch: %d vs %d",
				ti, test.Name, len(test.Inputs), len(test.Want))
		}

		for i, row := range test.Inputs {
			alpha := row[0]
			beta := row[1]
			val := row[2]

			got := betaCDF(alpha, beta, val)
			want := test.Want[i]

			if math.Abs(got-want) > tol {
				t.Errorf("Test %d (%s) case %d failed: betaCDF(%.4f,%.4f,%.4f) = %.5f, want %.5f",
					ti, test.Name, i, alpha, beta, val, got, want)
			}
		}
	}

	fmt.Println("betaCDF tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadBetaCDFTests(baseDir string) []BetaCDFTest {
	inputDir := filepath.Join(baseDir, "Input")
	outputDir := filepath.Join(baseDir, "Output")

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil
	}

	re := regexp.MustCompile(`^input(\d+)\.txt$`)

	type fileEntry struct {
		idx  int
		path string
	}

	var inputs []fileEntry

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := re.FindStringSubmatch(e.Name())
		if len(m) == 2 {
			n, _ := strconv.Atoi(m[1])
			inputs = append(inputs, fileEntry{idx: n, path: filepath.Join(inputDir, e.Name())})
		}
	}

	sort.Slice(inputs, func(i, j int) bool { return inputs[i].idx < inputs[j].idx })

	var tests []BetaCDFTest

	for _, in := range inputs {
		outPath := filepath.Join(outputDir, fmt.Sprintf("output%d.txt", in.idx))

		inRows, err := parseBetaCDFInputs(in.path)
		if err != nil {
			panic(fmt.Sprintf("Failed to read %s: %v", in.path, err))
		}

		want, err := parseBetaCDFOutputs(outPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to read %s: %v", outPath, err))
		}

		tests = append(tests, BetaCDFTest{
			Inputs: inRows,
			Want:   want,
			Name:   fmt.Sprintf("input%d", in.idx),
		})
	}

	return tests
}

func parseBetaCDFInputs(path string) ([][3]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cases [][3]float64
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid input line: %s", line)
		}

		a, err1 := strconv.ParseFloat(strings.TrimSpace(fields[0]), 64)
		b, err2 := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		v, err3 := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, fmt.Errorf("parse error in line: %s", line)
		}

		cases = append(cases, [3]float64{a, b, v})
	}

	return cases, scanner.Err()
}

func parseBetaCDFOutputs(path string) ([]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []float64
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		v, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid expected output: %s", line)
		}
		out = append(out, v)
	}

	return out, scanner.Err()
}

// ----------------------------------------------------------------------
// betaDistEstimation()
// ----------------------------------------------------------------------

type BetaDistEstimationTest struct {
	Name      string
	Sample    []float64
	Resp      []float64
	Weights   []float64
	WantAlpha float64
	WantBeta  float64
	FileName  string
}

func TestBetaDistEstimation(t *testing.T) {
	for i := 0; i < 3; i++ {
		inputFile := fmt.Sprintf("Tests/betaDistEstimation/Input/input%d.txt", i)
		outputFile := fmt.Sprintf("Tests/betaDistEstimation/Output/output%d.txt", i)

		sampleData, responsibilities, weights, err := parseBetaDistEstimationInput(inputFile)
		if err != nil {
			t.Fatalf("Failed to parse input file %s: %v", inputFile, err)
		}

		gotAlpha, gotBeta := betaDistEstimation(sampleData, responsibilities, weights)

		wantValues, err := readOutput(outputFile)
		if err != nil {
			t.Fatalf("Failed to read output file %s: %v", outputFile, err)
		}
		if len(wantValues) != 2 {
			t.Fatalf("Output file %s does not have 2 lines", outputFile)
		}
		wantAlpha := wantValues[0]
		wantBeta := wantValues[1]

		if math.Abs(gotAlpha-wantAlpha) > 1e-4 || math.Abs(gotBeta-wantBeta) > 1e-4 {
			t.Errorf("Test %d failed: got (%.6f, %.6f), want (%.6f, %.6f)", i, gotAlpha, gotBeta, wantAlpha, wantBeta)
		}
	}

	fmt.Println("BetaDistEstimation tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseBetaDistEstimationInput(filename string) ([]float64, []float64, []float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	defer file.Close()

	var sampleData, responsibilities, weights []float64
	var current *[]float64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.Contains(line, "sampleData") {
				current = &sampleData
			} else if strings.Contains(line, "responsibilities") {
				current = &responsibilities
			} else if strings.Contains(line, "weights") {
				current = &weights
			}
			continue
		}

		// Parse floats and append
		parts := strings.Fields(line)
		for _, p := range parts {
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return nil, nil, nil, err
			}
			*current = append(*current, f)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, err
	}
	return sampleData, responsibilities, weights, nil
}

func readOutput(filename string) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		val, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// ----------------------------------------------------------------------
// betaLogPDF()
// ----------------------------------------------------------------------
func TestBetaLogPDF(t *testing.T) {
	numTests := 3 // Number of input/output files: input0.txt -> input2.txt
	for i := 0; i < numTests; i++ {
		inputFile := fmt.Sprintf("Tests/betaLogPDF/Input/input%d.txt", i)
		outputFile := fmt.Sprintf("Tests/betaLogPDF/Output/output%d.txt", i)

		input, err := parseBetaLogPDFInput(inputFile)
		if err != nil {
			t.Fatalf("Failed to load input: %v", err)
		}
		want, err := parseBetaLogPDFOutput(outputFile)
		if err != nil {
			t.Fatalf("Failed to load output: %v", err)
		}

		got := betaLogPDF(input[0], input[1], input[2])

		if math.Abs(got-want) > 1e-6 {
			t.Errorf("Test %d failed: got %.6f, want %.6f", i, got, want)
		}
	}
	fmt.Println("betaLogPDF tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseBetaLogPDFInput(filename string) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid input format in file %s", filename)
		}
		res := make([]float64, 3)
		for i, part := range parts {
			val, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil, err
			}
			res[i] = val
		}
		return res, nil
	}

	return nil, fmt.Errorf("no valid input found in file %s", filename)
}

func parseBetaLogPDFOutput(filename string) (float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		val, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return 0, err
		}
		return val, nil
	}

	return 0, fmt.Errorf("no valid output found in file %s", filename)
}

// ----------------------------------------------------------------------
// BMIQAllSamples()
// ----------------------------------------------------------------------
func TestBMIQAllSamples(t *testing.T) {
	for i := 0; i < 3; i++ {
		inputFile := fmt.Sprintf("Tests/BMIQAllSamples/Input/input%d.txt", i)
		outputFile := fmt.Sprintf("Tests/BMIQAllSamples/Output/output%d.txt", i)

		dataMatrix, designType, err := ParseBMIQAllSamplesInput(inputFile)
		if err != nil {
			t.Fatalf("Failed to load input file: %v", err)
		}
		want, err := ParseBMIQAllSamplesOutput(outputFile)
		if err != nil {
			t.Fatalf("Failed to load output file: %v", err)
		}

		got, err := BMIQAllSamples(dataMatrix, designType)
		if err != nil {
			t.Fatalf("BMIQAllSamples returned error: %v", err)
		}

		// Compare shapes
		if len(got) != len(want) {
			t.Errorf("Test %d: got %d rows, want %d rows", i, len(got), len(want))
			continue
		}
		for r := range got {
			if len(got[r]) != len(want[r]) {
				t.Errorf("Test %d row %d: got %d cols, want %d cols", i, r, len(got[r]), len(want[r]))
				continue
			}
			for c := range got[r] {
				if math.Abs(got[r][c]-want[r][c]) > 1e-6 {
					t.Errorf("Test %d row %d col %d: got %.6f, want %.6f", i, r, c, got[r][c], want[r][c])
				}
			}
		}

	}
	fmt.Println("BMIQAllSamples tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ParseBMIQAllSamplesInput(filename string) ([][]float64, []int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var dataMatrix [][]float64
	var designType []int
	scanner := bufio.NewScanner(file)
	mode := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			// switch mode
			if strings.Contains(line, "dataMatrix") {
				mode = "dataMatrix"
			} else if strings.Contains(line, "designType") {
				mode = "designType"
			}
			continue
		}

		switch mode {
		case "dataMatrix":
			tokens := strings.Fields(line)
			row := make([]float64, len(tokens))
			for i, token := range tokens {
				val, err := strconv.ParseFloat(token, 64)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid float: %v", token)
				}
				row[i] = val
			}
			dataMatrix = append(dataMatrix, row)
		case "designType":
			tokens := strings.Fields(line)
			designType = make([]int, len(tokens))
			for i, token := range tokens {
				val, err := strconv.Atoi(token)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid int: %v", token)
				}
				designType[i] = val
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return dataMatrix, designType, nil
}

func ParseBMIQAllSamplesOutput(filename string) ([][]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var output [][]float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tokens := strings.Fields(line)
		row := make([]float64, len(tokens))
		for i, token := range tokens {
			val, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float: %v", token)
			}
			row[i] = val
		}
		output = append(output, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return output, nil
}

// ----------------------------------------------------------------------
// divideByProbe()
// ----------------------------------------------------------------------
func TestDivideByProbe(t *testing.T) {
	for i := 0; i < 3; i++ {
		inputFile := fmt.Sprintf("Tests/divideByProbe/Input/input%d.txt", i)
		outputFile := fmt.Sprintf("Tests/divideByProbe/Output/output%d.txt", i)

		sampleData, designType, numType1, numType2, err := parseDivideByProbeInput(inputFile)
		if err != nil {
			t.Fatalf("Failed to parse input: %v", err)
		}

		gotType1Idx, gotType2Idx, gotType1Betas, gotType2Betas := divideByProbe(sampleData, designType, numType1, numType2)

		expectedType1Idx, expectedType2Idx, expectedType1Betas, expectedType2Betas, err := parseDivideByProbeOutput(outputFile)
		if err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		if !reflect.DeepEqual(gotType1Idx, expectedType1Idx) || !reflect.DeepEqual(gotType2Idx, expectedType2Idx) ||
			!reflect.DeepEqual(gotType1Betas, expectedType1Betas) || !reflect.DeepEqual(gotType2Betas, expectedType2Betas) {
			t.Errorf("Test %d failed:\ngot:  %v %v %v %v\nwant: %v %v %v %v", i,
				gotType1Idx, gotType2Idx, gotType1Betas, gotType2Betas,
				expectedType1Idx, expectedType2Idx, expectedType1Betas, expectedType2Betas)
		}
	}

	fmt.Println("divideByProbe tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseDivideByProbeInput(filename string) ([]float64, []int, int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sampleData []float64
	var designType []int
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.Contains(line, "sampleData") {
				section = "sampleData"
			} else if strings.Contains(line, "designType") {
				section = "designType"
			}
			continue
		}
		fields := strings.Fields(line)
		switch section {
		case "sampleData":
			for _, f := range fields {
				val, err := strconv.ParseFloat(f, 64)
				if err != nil {
					return nil, nil, 0, 0, err
				}
				sampleData = append(sampleData, val)
			}
		case "designType":
			for _, f := range fields {
				val, err := strconv.Atoi(f)
				if err != nil {
					return nil, nil, 0, 0, err
				}
				designType = append(designType, val)
			}
		}
	}

	if len(sampleData) != len(designType) {
		return nil, nil, 0, 0, fmt.Errorf("sampleData length %d does not match designType length %d", len(sampleData), len(designType))
	}

	numType1 := 0
	numType2 := 0
	for _, dt := range designType {
		if dt == 1 {
			numType1++
		} else if dt == 2 {
			numType2++
		}
	}

	return sampleData, designType, numType1, numType2, nil
}

func parseDivideByProbeOutput(filename string) ([]int, []int, []float64, []float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var type1Idx, type2Idx []int
	var type1Betas, type2Betas []float64
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.Contains(line, "type1indices") {
				section = "type1indices"
			} else if strings.Contains(line, "type2indices") {
				section = "type2indices"
			} else if strings.Contains(line, "type1betas") {
				section = "type1betas"
			} else if strings.Contains(line, "type2betas") {
				section = "type2betas"
			}
			continue
		}

		fields := strings.Fields(line)
		switch section {
		case "type1indices":
			for _, f := range fields {
				val, _ := strconv.Atoi(f)
				type1Idx = append(type1Idx, val)
			}
		case "type2indices":
			for _, f := range fields {
				val, _ := strconv.Atoi(f)
				type2Idx = append(type2Idx, val)
			}
		case "type1betas":
			for _, f := range fields {
				val, _ := strconv.ParseFloat(f, 64)
				type1Betas = append(type1Betas, val)
			}
		case "type2betas":
			for _, f := range fields {
				val, _ := strconv.ParseFloat(f, 64)
				type2Betas = append(type2Betas, val)
			}
		}
	}

	return type1Idx, type2Idx, type1Betas, type2Betas, nil
}

// ----------------------------------------------------------------------
// fitEMBetaMixture()
// ----------------------------------------------------------------------
func TestFitEMBetaMixture(t *testing.T) {
	testDir := "Tests/fitEMBetaMixture"
	inputDir := filepath.Join(testDir, "Input")
	outputDir := filepath.Join(testDir, "Output")

	files, err := os.ReadDir(inputDir)
	if err != nil {
		t.Fatalf("Failed to read input directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		inputPath := filepath.Join(inputDir, file.Name())

		// Map "input0.txt" → "output0.txt"
		outName := strings.Replace(file.Name(), "input", "output", 1)
		outputPath := filepath.Join(outputDir, outName)

		// Parse input
		sampleData, numStates, initResp, maxIter, tol, err := parseEMInput(inputPath)
		if err != nil {
			t.Errorf("Failed to parse input %s: %v", file.Name(), err)
			continue
		}

		// Parse expected output
		expected, err := parseEMOutput(outputPath)
		if err != nil {
			t.Errorf("Failed to parse output %s: %v", outName, err)
			continue
		}

		// Run the function
		result, err := fitEMBetaMixture(sampleData, numStates, initResp, maxIter, tol)
		if err != nil {
			t.Errorf("fitEMBetaMixture returned error for %s: %v", file.Name(), err)
			continue
		}

		// Compare alpha
		if len(result.Alpha) != len(expected.Alpha) {
			t.Errorf("Test %s alpha length mismatch: got %d, want %d", file.Name(), len(result.Alpha), len(expected.Alpha))
		} else {
			for i := range result.Alpha {
				if math.Abs(result.Alpha[i]-expected.Alpha[i]) > 1e-6 {
					t.Errorf("Test %s alpha[%d]: got %f, want %f", file.Name(), i, result.Alpha[i], expected.Alpha[i])
				}
			}
		}

		// Compare beta
		if len(result.Beta) != len(expected.Beta) {
			t.Errorf("Test %s beta length mismatch: got %d, want %d", file.Name(), len(result.Beta), len(expected.Beta))
		} else {
			for i := range result.Beta {
				if math.Abs(result.Beta[i]-expected.Beta[i]) > 1e-6 {
					t.Errorf("Test %s beta[%d]: got %f, want %f", file.Name(), i, result.Beta[i], expected.Beta[i])
				}
			}
		}

		// Compare stateProportion
		if len(result.StateProportion) != len(expected.StateProportion) {
			t.Errorf("Test %s StateProportion length mismatch: got %d, want %d", file.Name(), len(result.StateProportion), len(expected.StateProportion))
		} else {
			for i := range result.StateProportion {
				if math.Abs(result.StateProportion[i]-expected.StateProportion[i]) > 1e-6 {
					t.Errorf("Test %s StateProportion[%d]: got %f, want %f", file.Name(), i, result.StateProportion[i], expected.StateProportion[i])
				}
			}
		}

		// Compare responsibilities
		if len(result.Responsibilities) != len(expected.Responsibilities) {
			t.Errorf("Test %s Responsibilities row count mismatch: got %d, want %d", file.Name(), len(result.Responsibilities), len(expected.Responsibilities))
		} else {
			for i := range result.Responsibilities {
				if len(result.Responsibilities[i]) != len(expected.Responsibilities[i]) {
					t.Errorf("Test %s Responsibilities row %d length mismatch: got %d, want %d",
						file.Name(), i, len(result.Responsibilities[i]), len(expected.Responsibilities[i]))
					continue
				}
				for j := range result.Responsibilities[i] {
					if math.Abs(result.Responsibilities[i][j]-expected.Responsibilities[i][j]) > 1e-6 {
						t.Errorf("Test %s Responsibilities[%d][%d]: got %f, want %f",
							file.Name(), i, j, result.Responsibilities[i][j], expected.Responsibilities[i][j])
					}
				}
			}
		}
	}

	fmt.Println("fitEMBetaMixture tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseEMOutput(filePath string) (*EMResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := &EMResult{}
	state := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.Contains(line, "Alpha") {
				state = "Alpha"
			} else if strings.Contains(line, "Beta") {
				state = "Beta"
			} else if strings.Contains(line, "StateProportion") {
				state = "StateProportion"
			} else if strings.Contains(line, "Responsibilities") {
				state = "Responsibilities"
			}
			continue
		}
		switch state {
		case "Alpha":
			for _, s := range strings.Fields(line) {
				val, _ := strconv.ParseFloat(s, 64)
				result.Alpha = append(result.Alpha, val)
			}
		case "Beta":
			for _, s := range strings.Fields(line) {
				val, _ := strconv.ParseFloat(s, 64)
				result.Beta = append(result.Beta, val)
			}
		case "StateProportion":
			for _, s := range strings.Fields(line) {
				val, _ := strconv.ParseFloat(s, 64)
				result.StateProportion = append(result.StateProportion, val)
			}
		case "Responsibilities":
			fields := strings.Fields(line)
			row := make([]float64, len(fields))
			for i, f := range fields {
				row[i], _ = strconv.ParseFloat(f, 64)
			}
			result.Responsibilities = append(result.Responsibilities, row)
		}
	}
	return result, nil
}

func parseEMInput(filePath string) ([]float64, int, [][]float64, int, float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, nil, 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sampleData []float64
	var numStates int
	var initResponsibilities [][]float64
	var maxIter int
	var tolerance float64
	state := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.Contains(line, "sampleData") {
				state = "sampleData"
			} else if strings.Contains(line, "numStates") {
				state = "numStates"
			} else if strings.Contains(line, "initResponsibilities") {
				state = "initResponsibilities"
			} else if strings.Contains(line, "maxIter") {
				state = "maxIter"
			} else if strings.Contains(line, "tolerance") {
				state = "tolerance"
			}
			continue
		}

		switch state {
		case "sampleData":
			for _, s := range strings.Fields(line) {
				val, _ := strconv.ParseFloat(s, 64)
				sampleData = append(sampleData, val)
			}
		case "numStates":
			numStates, _ = strconv.Atoi(line)
		case "initResponsibilities":
			if line == "nil" {
				initResponsibilities = nil
			} else {
				fields := strings.Fields(line)
				row := make([]float64, len(fields))
				for i, f := range fields {
					row[i], _ = strconv.ParseFloat(f, 64)
				}
				initResponsibilities = append(initResponsibilities, row)
			}
		case "maxIter":
			maxIter, _ = strconv.Atoi(line)
		case "tolerance":
			tolerance, _ = strconv.ParseFloat(line, 64)
		}
	}
	return sampleData, numStates, initResponsibilities, maxIter, tolerance, nil
}

// ----------------------------------------------------------------------
// initalizeResponsibiltiies()
// ----------------------------------------------------------------------
func TestInitalizeResponsibilities(t *testing.T) {
	baseDir := "Tests/initalizeResponsibilities"

	inputFiles, err := filepath.Glob(filepath.Join(baseDir, "Input", "input*.txt"))
	if err != nil {
		t.Fatalf("failed globbing input files: %v", err)
	}
	outputFiles, err := filepath.Glob(filepath.Join(baseDir, "Output", "output*.txt"))
	if err != nil {
		t.Fatalf("failed globbing output files: %v", err)
	}

	if len(inputFiles) != len(outputFiles) {
		t.Fatalf("mismatched input/output file counts: %d vs %d", len(inputFiles), len(outputFiles))
	}

	sort.Strings(inputFiles)
	sort.Strings(outputFiles)

	for i := range inputFiles {
		inPath := inputFiles[i]
		outPath := outputFiles[i]

		t.Run(filepath.Base(inPath), func(t *testing.T) {
			length, numStates, initResp, err := parseInitalizeResponsibilitiesInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseInitalizeResponsibilitiesOutput(outPath, numStates)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := initalizeResponsibiltiies(length, numStates, initResp)

			for r := range got {
				for c := range got[r] {
					got[r][c] = roundFloat(got[r][c], 6)
					expected[r][c] = roundFloat(expected[r][c], 6)
				}
			}

			if !matricesEqual(got, expected) {
				t.Errorf("responsibilities mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}
	fmt.Println("initalizeResponsibiltiies tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseInitalizeResponsibilitiesOutput(path string, numStates int) ([][]float64, error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return nil, err
	}

	var mode string
	var dataLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			if strings.Contains(line, "Responsibilities") {
				mode = "matrix"
			}
			continue
		}

		if mode == "matrix" {
			dataLines = append(dataLines, line)
		}
	}

	matrix := make([][]float64, len(dataLines))
	for i, row := range dataLines {
		parts := strings.Fields(row)
		if len(parts) != numStates {
			return nil, fmt.Errorf("output row %d: expected %d values, got %d",
				i, numStates, len(parts))
		}

		matrix[i] = make([]float64, numStates)
		for j, p := range parts {
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float %q: %v", p, err)
			}
			matrix[i][j] = f
		}
	}

	return matrix, nil
}

func parseInitalizeResponsibilitiesInput(path string) (length int, numStates int, init [][]float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return 0, 0, nil, err
	}

	var mode string
	var initLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Comments tell us what the next line represents
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "length"):
				mode = "length"

			case strings.Contains(line, "numStates"):
				mode = "numStates"

			case strings.Contains(line, "initResponsibilities"):
				mode = "initResponsibilities"

			default:
				// ignore unknown comments
				mode = ""
			}
			continue
		}

		// Non-comment lines interpret based on mode
		switch mode {
		case "length":
			length, err = strconv.Atoi(line)
			if err != nil {
				return 0, 0, nil, fmt.Errorf("invalid length %q: %v", line, err)
			}

		case "numStates":
			numStates, err = strconv.Atoi(line)
			if err != nil {
				return 0, 0, nil, fmt.Errorf("invalid numStates %q: %v", line, err)
			}

		case "initResponsibilities":
			if line == "nil" {
				init = nil
			} else {
				initLines = append(initLines, line)
			}
		}
	}

	// Parse matrix if present
	if initLines != nil {
		matrix := make([][]float64, len(initLines))
		for i, row := range initLines {
			parts := strings.Fields(row)
			if len(parts) != numStates {
				return 0, 0, nil, fmt.Errorf("init row %d: expected %d values, got %d",
					i, numStates, len(parts))
			}
			matrix[i] = make([]float64, numStates)
			for j, p := range parts {
				f, err := strconv.ParseFloat(p, 64)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("invalid float %q: %v", p, err)
				}
				matrix[i][j] = f
			}
		}
		init = matrix
	}

	return length, numStates, init, nil
}

// ----------------------------------------------------------------------
// logLikelihood()
// ----------------------------------------------------------------------
func TestLogLikelihood(t *testing.T) {
	baseDir := "Tests/logLikelihood"

	inputFiles, err := filepath.Glob(filepath.Join(baseDir, "Input", "input*.txt"))
	if err != nil {
		t.Fatalf("failed globbing input files: %v", err)
	}
	outputFiles, err := filepath.Glob(filepath.Join(baseDir, "Output", "output*.txt"))
	if err != nil {
		t.Fatalf("failed globbing output files: %v", err)
	}

	if len(inputFiles) != len(outputFiles) {
		t.Fatalf("mismatched input/output file counts: %d vs %d", len(inputFiles), len(outputFiles))
	}

	sort.Strings(inputFiles)
	sort.Strings(outputFiles)

	for i := range inputFiles {
		inPath := inputFiles[i]
		outPath := outputFiles[i]

		t.Run(filepath.Base(inPath), func(t *testing.T) {
			alpha, beta, sampleData, weights, responsibilities, err := parseLogLikelihoodInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseLogLikelihoodOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := logLikelihood(alpha, beta, sampleData, weights, responsibilities)

			got = roundFloat(got, 6)
			expected = roundFloat(expected, 6)

			if got != expected {
				t.Errorf("logLikelihood mismatch\nExpected: %.6f\nGot: %.6f", expected, got)
			}
		})
	}

	fmt.Println("logLikelihood tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseLogLikelihoodInput(path string) (alpha, beta float64, sampleData, weights, responsibilities []float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return 0, 0, nil, nil, nil, err
	}

	var mode string

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "alpha"):
				mode = "alpha"
			case strings.Contains(line, "beta"):
				mode = "beta"
			case strings.Contains(line, "sampleData"):
				mode = "sampleData"
			case strings.Contains(line, "weights"):
				mode = "weights"
			case strings.Contains(line, "responsibilities"):
				mode = "responsibilities"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "alpha":
			alpha, err = strconv.ParseFloat(line, 64)
			if err != nil {
				return
			}
		case "beta":
			beta, err = strconv.ParseFloat(line, 64)
			if err != nil {
				return
			}
		case "sampleData":
			sampleData, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "weights":
			if line == "nil" {
				weights = nil
			} else {
				weights, err = parseFloatSlice(line)
				if err != nil {
					return
				}
			}
		case "responsibilities":
			if line == "nil" {
				responsibilities = nil
			} else {
				responsibilities, err = parseFloatSlice(line)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func parseLogLikelihoodOutput(path string) (float64, error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return 0, err
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		return strconv.ParseFloat(line, 64)
	}

	return 0, fmt.Errorf("no value found in output file")
}

// ----------------------------------------------------------------------
// methodOfMoments()
// ----------------------------------------------------------------------
func TestMethodOfMoments(t *testing.T) {
	baseDir := "Tests/methodOfMoments"

	inputFiles, err := filepath.Glob(filepath.Join(baseDir, "Input", "input*.txt"))
	if err != nil {
		t.Fatalf("failed globbing input files: %v", err)
	}
	outputFiles, err := filepath.Glob(filepath.Join(baseDir, "Output", "output*.txt"))
	if err != nil {
		t.Fatalf("failed globbing output files: %v", err)
	}

	if len(inputFiles) != len(outputFiles) {
		t.Fatalf("mismatched input/output file counts: %d vs %d", len(inputFiles), len(outputFiles))
	}

	sort.Strings(inputFiles)
	sort.Strings(outputFiles)

	for i := range inputFiles {
		inPath := inputFiles[i]
		outPath := outputFiles[i]

		t.Run(filepath.Base(inPath), func(t *testing.T) {
			data, responsibilities, weights, err := parseMethodOfMomentsInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseMethodOfMomentsOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := methodOfMoments(data, responsibilities, weights)

			// Round each value for comparison
			for j := range got {
				got[j] = roundFloat(got[j], 6)
				if j < len(expected) {
					expected[j] = roundFloat(expected[j], 6)
				}
			}

			if !matricesEqual([][]float64{got}, [][]float64{expected}) {
				t.Errorf("methodOfMoments mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}

	fmt.Println("methodOfMoments tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseMethodOfMomentsInput(path string) (data, responsibilities, weights []float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return nil, nil, nil, err
	}

	var mode string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "data"):
				mode = "data"
			case strings.Contains(line, "responsibilities"):
				mode = "responsibilities"
			case strings.Contains(line, "weights"):
				mode = "weights"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "data":
			data, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "responsibilities":
			if line == "nil" {
				responsibilities = nil
			} else {
				responsibilities, err = parseFloatSlice(line)
				if err != nil {
					return
				}
			}
		case "weights":
			if line == "nil" {
				weights = nil
			} else {
				weights, err = parseFloatSlice(line)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func parseMethodOfMomentsOutput(path string) ([]float64, error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return nil, err
	}

	var output []float64
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		slice, err := parseFloatSlice(line)
		if err != nil {
			return nil, err
		}
		output = append(output, slice...)
	}

	return output, nil
}

// ----------------------------------------------------------------------
// preprocessingClamp()
// ----------------------------------------------------------------------
func TestPreprocessingClamp(t *testing.T) {
	baseDir := "Tests/preprocessingClamp"

	inputFiles, err := filepath.Glob(filepath.Join(baseDir, "Input", "input*.txt"))
	if err != nil {
		t.Fatalf("failed globbing input files: %v", err)
	}
	outputFiles, err := filepath.Glob(filepath.Join(baseDir, "Output", "output*.txt"))
	if err != nil {
		t.Fatalf("failed globbing output files: %v", err)
	}

	if len(inputFiles) != len(outputFiles) {
		t.Fatalf("mismatched input/output file counts: %d vs %d", len(inputFiles), len(outputFiles))
	}

	sort.Strings(inputFiles)
	sort.Strings(outputFiles)

	for i := range inputFiles {
		inPath := inputFiles[i]
		outPath := outputFiles[i]

		t.Run(filepath.Base(inPath), func(t *testing.T) {
			value, adjustment, err := parsePreprocessingClampInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parsePreprocessingClampOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := preprocessingClamp(value, adjustment)
			got = roundFloat(got, 6)
			expected = roundFloat(expected, 6)

			if got != expected {
				t.Errorf("preprocessingClamp mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}

	fmt.Println("preprocessingClamp tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parsePreprocessingClampInput(path string) (value, adjustment float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return 0, 0, err
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("expected 2 values in line, got %d", len(parts))
		}
		value, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return
		}
		adjustment, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return
		}
		break
	}

	return
}

func parsePreprocessingClampOutput(path string) (float64, error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return 0, err
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		return strconv.ParseFloat(strings.TrimSpace(line), 64)
	}

	return 0, fmt.Errorf("no output found")
}

// ----------------------------------------------------------------------
// updateResponsibilities()
// ----------------------------------------------------------------------
func TestUpdateResponsibilities(t *testing.T) {
	baseDir := "Tests/updateResponsibilities"

	inputFiles, err := filepath.Glob(filepath.Join(baseDir, "Input", "input*.txt"))
	if err != nil {
		t.Fatalf("failed globbing input files: %v", err)
	}
	outputFiles, err := filepath.Glob(filepath.Join(baseDir, "Output", "output*.txt"))
	if err != nil {
		t.Fatalf("failed globbing output files: %v", err)
	}

	if len(inputFiles) != len(outputFiles) {
		t.Fatalf("mismatched input/output file counts: %d vs %d", len(inputFiles), len(outputFiles))
	}

	sort.Strings(inputFiles)
	sort.Strings(outputFiles)

	for i := range inputFiles {
		inPath := inputFiles[i]
		outPath := outputFiles[i]

		t.Run(filepath.Base(inPath), func(t *testing.T) {
			length, numStates, sampleData, weights, logProbs, stateProportion, alpha, beta, responsibilities, err := parseUpdateResponsibilitiesInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseUpdateResponsibilitiesOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			updateResponsibilities(length, numStates, sampleData, weights, logProbs, stateProportion, alpha, beta, responsibilities)

			// Round for comparison
			for i := range responsibilities {
				for j := range responsibilities[i] {
					responsibilities[i][j] = roundFloat(responsibilities[i][j], 6)
					if i < len(expected) && j < len(expected[i]) {
						expected[i][j] = roundFloat(expected[i][j], 6)
					}
				}
			}

			if !matricesEqual(responsibilities, expected) {
				t.Errorf("updateResponsibilities mismatch\nExpected: %v\nGot: %v", expected, responsibilities)
			}
		})
	}

	fmt.Println("updateResponsibilities tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseUpdateResponsibilitiesInput(path string) (length, numStates int, sampleData, weights, logProbs, stateProportion, alpha, beta []float64, responsibilities [][]float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return
	}

	mode := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "length"):
				mode = "length"
			case strings.Contains(line, "numStates"):
				mode = "numStates"
			case strings.Contains(line, "sampleData"):
				mode = "sampleData"
			case strings.Contains(line, "weights"):
				mode = "weights"
			case strings.Contains(line, "logProbs"):
				mode = "logProbs"
			case strings.Contains(line, "stateProportion"):
				mode = "stateProportion"
			case strings.Contains(line, "alpha"):
				mode = "alpha"
			case strings.Contains(line, "beta"):
				mode = "beta"
			case strings.Contains(line, "responsibilities"):
				mode = "responsibilities"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "length":
			length64, err2 := strconv.ParseInt(line, 10, 0)
			if err2 != nil {
				err = err2
				return
			}
			length = int(length64)
		case "numStates":
			numStates64, err2 := strconv.ParseInt(line, 10, 0)
			if err2 != nil {
				err = err2
				return
			}
			numStates = int(numStates64)
		case "sampleData":
			sampleData, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "weights":
			weights, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "logProbs":
			logProbs, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "stateProportion":
			stateProportion, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "alpha":
			alpha, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "beta":
			beta, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "responsibilities":
			linesplit := strings.Fields(line)
			row := make([]float64, len(linesplit))
			for i, s := range linesplit {
				row[i], err = strconv.ParseFloat(s, 64)
				if err != nil {
					return
				}
			}
			responsibilities = append(responsibilities, row)
		}
	}
	return
}

func parseUpdateResponsibilitiesOutput(path string) ([][]float64, error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return nil, err
	}

	var out [][]float64
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		row := make([]float64, len(parts))
		for i, s := range parts {
			row[i], err = strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
		}
		out = append(out, row)
	}
	return out, nil
}

// ----------------------------------------------------------------------
// weightedMean()
// ----------------------------------------------------------------------
func TestWeightedMean(t *testing.T) {
	baseDir := "Tests/weightedMean"

	inputFiles, err := filepath.Glob(filepath.Join(baseDir, "Input", "input*.txt"))
	if err != nil {
		t.Fatalf("failed globbing input files: %v", err)
	}
	outputFiles, err := filepath.Glob(filepath.Join(baseDir, "Output", "output*.txt"))
	if err != nil {
		t.Fatalf("failed globbing output files: %v", err)
	}

	if len(inputFiles) != len(outputFiles) {
		t.Fatalf("mismatched input/output file counts: %d vs %d", len(inputFiles), len(outputFiles))
	}

	sort.Strings(inputFiles)
	sort.Strings(outputFiles)

	for i := range inputFiles {
		inPath := inputFiles[i]
		outPath := outputFiles[i]

		t.Run(filepath.Base(inPath), func(t *testing.T) {
			data, responsibilities, weights, err := parseWeightedMeanInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseWeightedMeanOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := weightedMean(data, responsibilities, weights)
			got = roundFloat(got, 6)
			expected = roundFloat(expected, 6)

			if got != expected {
				t.Errorf("weightedMean mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}

	fmt.Println("weightedMean tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseWeightedMeanInput(path string) (data, responsibilities, weights []float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return
	}

	mode := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "data"):
				mode = "data"
			case strings.Contains(line, "responsibilities"):
				mode = "responsibilities"
			case strings.Contains(line, "weights"):
				mode = "weights"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "data":
			data, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "responsibilities":
			responsibilities, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "weights":
			weights, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		}
	}
	return
}

func parseWeightedMeanOutput(path string) (float64, error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return 0, err
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		return strconv.ParseFloat(strings.TrimSpace(line), 64)
	}
	return 0, fmt.Errorf("no output found")
}

// ----------------------------------------------------------------------
// weightedSampleVariance()
// ----------------------------------------------------------------------
func TestWeightedSampleVariance(t *testing.T) {
	baseDir := "Tests/weightedSampleVariance"

	inputFiles, err := filepath.Glob(filepath.Join(baseDir, "Input", "input*.txt"))
	if err != nil {
		t.Fatalf("failed globbing input files: %v", err)
	}
	outputFiles, err := filepath.Glob(filepath.Join(baseDir, "Output", "output*.txt"))
	if err != nil {
		t.Fatalf("failed globbing output files: %v", err)
	}

	if len(inputFiles) != len(outputFiles) {
		t.Fatalf("mismatched input/output file counts: %d vs %d", len(inputFiles), len(outputFiles))
	}

	sort.Strings(inputFiles)
	sort.Strings(outputFiles)

	for i := range inputFiles {
		inPath := inputFiles[i]
		outPath := outputFiles[i]

		t.Run(filepath.Base(inPath), func(t *testing.T) {
			data, responsibilities, weights, mean, err := parseWeightedSampleVarianceInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseWeightedSampleVarianceOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := weightedSampleVariance(data, responsibilities, weights, mean)
			got = roundFloat(got, 6)
			expected = roundFloat(expected, 6)

			if got != expected {
				t.Errorf("weightedSampleVariance mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}

	fmt.Println("weightedSampleVariance tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseWeightedSampleVarianceInput(path string) (data, responsibilities, weights []float64, mean float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return
	}

	mode := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "data"):
				mode = "data"
			case strings.Contains(line, "responsibilities"):
				mode = "responsibilities"
			case strings.Contains(line, "weights"):
				mode = "weights"
			case strings.Contains(line, "mean"):
				mode = "mean"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "data":
			data, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "responsibilities":
			responsibilities, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "weights":
			weights, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "mean":
			mean, err = strconv.ParseFloat(strings.TrimSpace(line), 64)
			if err != nil {
				return
			}
		}
	}
	return
}

func parseWeightedSampleVarianceOutput(path string) (float64, error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return 0, err
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		return strconv.ParseFloat(strings.TrimSpace(line), 64)
	}
	return 0, fmt.Errorf("no output found")
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

func readCleanLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func roundFloat(x float64, places int) float64 {
	if places < 0 {
		return x
	}
	pow := math.Pow(10, float64(places))
	return math.Round(x*pow) / pow
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

func parseFloatSlice(line string) ([]float64, error) {
	parts := strings.Fields(line)
	slice := make([]float64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float %q: %v", p, err)
		}
		slice[i] = v
	}
	return slice, nil
}

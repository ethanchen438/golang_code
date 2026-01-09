package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------
// TestPredictAgeFromWeights()
// ----------------------------------------------------------------------

func TestPredictAgeFromWeights(t *testing.T) {
	tests := ReadPredictAgeTests("Tests/predictAgeFromWeights")

	for i, test := range tests {
		got, err := predictAgeFromWeights(test.FeatureNames, test.Rows, test.Model)
		if err != nil {
			t.Fatalf("Test #%d failed with error: %v", i, err)
		}

		if len(got) != len(test.Want) {
			t.Fatalf("Test #%d length mismatch: got %d, want %d", i, len(got), len(test.Want))
		}

		for j := range got {
			if roundFloat(got[j], 4) != roundFloat(test.Want[j], 4) {
				t.Errorf("Test #%d sample %d failed: got %.4f, want %.4f", i, j, got[j], test.Want[j])
			}
		}
	}

	fmt.Println("predictAgeFromWeights tested!")
}

type PredictAgeTest struct {
	FeatureNames []string
	Rows         [][]string
	Model        *ModelConfig
	Want         []float64
}

// Parsers //
// Written using AI

func parsePredictAgeInput(filePath string) ([]string, [][]string, *ModelConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not open file '%s': %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var featureNames []string
	var rows [][]string
	model := &ModelConfig{
		FeatureMeans: make(map[string]float64),
		FeatureStds:  make(map[string]float64),
		Weights:      make(map[string]float64),
	}

	section := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(line, "FEATURE_NAMES:"):
			section = "features"
		case strings.HasPrefix(line, "ROWS:"):
			section = "rows"
		case strings.HasPrefix(line, "MODEL:"):
			section = "model"
		case strings.HasPrefix(line, "FEATURE_MEANS:"):
			section = "means"
		case strings.HasPrefix(line, "FEATURE_STDS:"):
			section = "stds"
		case strings.HasPrefix(line, "WEIGHTS:"):
			section = "weights"
		case strings.HasPrefix(line, "BIAS:"):
			section = "bias"
		default:
			switch section {
			case "features":
				featureNames = strings.Fields(line)
			case "rows":
				rows = append(rows, strings.Fields(line))
			case "model":
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					if err != nil {
						continue
					}
					if key == "YMean" {
						model.Norm.YMean = val
					} else if key == "YStd" {
						model.Norm.YStd = val
					}
				}
			case "means":
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					model.FeatureMeans[strings.TrimSpace(parts[0])] = val
				}
			case "stds":
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					model.FeatureStds[strings.TrimSpace(parts[0])] = val
				}
			case "weights":
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
					model.Weights[strings.TrimSpace(parts[0])] = val
				}
			case "bias":
				val, _ := strconv.ParseFloat(strings.TrimSpace(line), 64)
				model.Bias = val
			}
		}
	}

	return featureNames, rows, model, scanner.Err()
}

func parsePredictAgeOutput(filePath string) ([]float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open output file '%s': %w", filePath, err)
	}
	defer file.Close()

	var want []float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "Predicted_Ages:") {
			continue
		}
		val, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float in output file: %s", line)
		}
		want = append(want, val)
	}
	return want, scanner.Err()
}

func ReadPredictAgeTests(basePath string) []PredictAgeTest {
	tests := []PredictAgeTest{}
	for i := 0; i < 3; i++ {
		inputPath := fmt.Sprintf("%s/Input/input%d.txt", basePath, i)
		outputPath := fmt.Sprintf("%s/Output/output%d.txt", basePath, i)

		features, rows, model, err := parsePredictAgeInput(inputPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to read input file %s: %v", inputPath, err))
		}

		want, err := parsePredictAgeOutput(outputPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to read output file %s: %v", outputPath, err))
		}

		tests = append(tests, PredictAgeTest{
			FeatureNames: features,
			Rows:         rows,
			Model:        model,
			Want:         want,
		})
	}

	return tests
}

// ----------------------------------------------------------------------
// transposeAndImpute()
// ----------------------------------------------------------------------

func TestTransposeAndImpute(t *testing.T) {
	tests := ReadTransposeAndImputeTests("Tests/transposeAndImpute")

	for i, test := range tests {
		got, gotNumSamples, err := transposeAndImpute(test.Rows)
		if err != nil {
			t.Fatalf("Test #%d failed with error: %v", i, err)
		}

		if gotNumSamples != test.NumSamples {
			t.Errorf("Test #%d: numSamples mismatch: got %d, want %d", i, gotNumSamples, test.NumSamples)
		}

		if len(got) != len(test.Want) {
			t.Errorf("Test #%d: output row count mismatch: got %d, want %d", i, len(got), len(test.Want))
		}

		for r := range got {
			if len(got[r]) != len(test.Want[r]) {
				t.Errorf("Test #%d: output row %d length mismatch: got %d, want %d", i, r, len(got[r]), len(test.Want[r]))
				continue
			}
			for c := range got[r] {
				if math.Abs(got[r][c]-test.Want[r][c]) > 1e-6 {
					t.Errorf("Test #%d: output[%d][%d] mismatch: got %f, want %f", i, r, c, got[r][c], test.Want[r][c])
				}
			}
		}
	}

	fmt.Println("transposeAndImpute tested!")
}

type TransposeAndImputeTest struct {
	Rows       [][]string
	NumSamples int
	Want       [][]float64
}

// Parsers //
// Written using AI

func parseTransposeAndImputeInput(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open input file '%s': %w", filePath, err)
	}
	defer file.Close()

	var rows [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rows = append(rows, strings.Fields(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

func parseTransposeAndImputeOutput(filePath string) ([][]float64, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, fmt.Errorf("could not open output file '%s': %w", filePath, err)
	}
	defer file.Close()

	var want [][]float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		row := make([]float64, len(parts))
		for i, p := range parts {
			val, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid float in output file: %s", p)
			}
			row[i] = val
		}
		want = append(want, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	numSamples := len(want) // rows = samples
	return want, numSamples, nil
}

func ReadTransposeAndImputeTests(basePath string) []TransposeAndImputeTest {
	tests := []TransposeAndImputeTest{}

	for i := 0; i < 3; i++ { // adjust count based on number of test cases
		inputPath := fmt.Sprintf("%s/Input/input%d.txt", basePath, i)
		outputPath := fmt.Sprintf("%s/Output/output%d.txt", basePath, i)

		rows, err := parseTransposeAndImputeInput(inputPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to read input file %s: %v", inputPath, err))
		}

		want, numSamples, err := parseTransposeAndImputeOutput(outputPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to read output file %s: %v", outputPath, err))
		}

		tests = append(tests, TransposeAndImputeTest{
			Rows:       rows,
			NumSamples: numSamples,
			Want:       want,
		})
	}

	return tests
}

// ---------------------------
// Helpers
// ---------------------------

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

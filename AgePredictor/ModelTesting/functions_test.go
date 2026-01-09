package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// All parsers made with AI //

// ----------------------------------------------------------------------
// pearsonCorrelation()
// ----------------------------------------------------------------------
func TestPearsonCorrelation(t *testing.T) {
	baseDir := "Tests/pearsonCorrelation"

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
			predicted, chronological, err := parsePearsonInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parsePearsonOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := pearsonCorrelation(predicted, chronological)
			got = roundFloat(got, 6)
			expected = roundFloat(expected, 6)

			if got != expected {
				t.Errorf("pearsonCorrelation mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}

	fmt.Println("pearsonCorrelation tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parsePearsonInput(path string) (predicted, chronological []float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return
	}

	mode := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "predicted"):
				mode = "predicted"
			case strings.Contains(line, "chronological"):
				mode = "chronological"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "predicted":
			predicted, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "chronological":
			chronological, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		}
	}
	return
}

func parsePearsonOutput(path string) (float64, error) {
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
// medianAbsoluteError()
// ----------------------------------------------------------------------

func TestMedianAbsoluteError(t *testing.T) {
	baseDir := "Tests/medianAbsoluteError"

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
			predicted, chronological, err := parseMedianAbsoluteErrorInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseMedianAbsoluteErrorOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := medianAbsoluteError(predicted, chronological)
			got = roundFloat(got, 6)
			expected = roundFloat(expected, 6)

			if got != expected {
				t.Errorf("medianAbsoluteError mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}

	fmt.Println("medianAbsoluteError tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseMedianAbsoluteErrorInput(path string) (predicted, chronological []float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return
	}

	mode := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "predicted"):
				mode = "predicted"
			case strings.Contains(line, "chronological"):
				mode = "chronological"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "predicted":
			predicted, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "chronological":
			chronological, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		}
	}
	return
}

func parseMedianAbsoluteErrorOutput(path string) (float64, error) {
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
// meanError()
// ----------------------------------------------------------------------
func TestMeanError(t *testing.T) {
	baseDir := "Tests/meanError"

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
			predicted, chronological, err := parseMeanErrorInput(inPath)
			if err != nil {
				t.Fatalf("failed parsing input: %v", err)
			}

			expected, err := parseMeanErrorOutput(outPath)
			if err != nil {
				t.Fatalf("failed parsing expected output: %v", err)
			}

			got := meanError(predicted, chronological)
			got = roundFloat(got, 6)
			expected = roundFloat(expected, 6)

			if got != expected {
				t.Errorf("meanError mismatch\nExpected: %v\nGot: %v", expected, got)
			}
		})
	}

	fmt.Println("meanError tested!")
}

// ----------------------------------------------------------------------
// Parsers

func parseMeanErrorInput(path string) (predicted, chronological []float64, err error) {
	lines, err := readCleanLines(path)
	if err != nil {
		return
	}

	mode := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			switch {
			case strings.Contains(line, "predicted"):
				mode = "predicted"
			case strings.Contains(line, "chronological"):
				mode = "chronological"
			default:
				mode = ""
			}
			continue
		}

		switch mode {
		case "predicted":
			predicted, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		case "chronological":
			chronological, err = parseFloatSlice(line)
			if err != nil {
				return
			}
		}
	}
	return
}

func parseMeanErrorOutput(path string) (float64, error) {
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

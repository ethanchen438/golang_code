package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// All parsers made with AI //

// ----------------------------------------------------------------------
// parseAndNormalizeAge()
// ----------------------------------------------------------------------

type ParseNormalizeAgeTest struct {
	RawAges []string
	Want    []string
	Name    string
}

func TestParseAndNormalizeAge(t *testing.T) {
	tests := ReadParseNormalizeAgeTests("Tests/parseAndNormalizeAge")

	if len(tests) == 0 {
		t.Fatalf("No tests found in Tests/parseAndNormalizeAge (expected input/output pairs).")
	}

	for i, test := range tests {
		if len(test.RawAges) != len(test.Want) {
			t.Fatalf("Test #%d (%s): input/output length mismatch: %d vs %d", i, test.Name, len(test.RawAges), len(test.Want))
		}

		for j := range test.RawAges {
			got := parseAndNormalizeAge(test.RawAges[j])
			// Convert both got and want to float64 if possible
			gotF, err1 := strconv.ParseFloat(got, 64)
			wantF, err2 := strconv.ParseFloat(test.Want[j], 64)

			if err1 != nil || err2 != nil {
				// fallback to string comparison if parsing fails
				if got != test.Want[j] {
					t.Errorf("Test #%d sample %d failed: got %q, want %q", i, j, got, test.Want[j])
				}
			} else if math.Abs(gotF-wantF) > 1e-6 {
				t.Errorf("Test #%d sample %d failed: got %.2f, want %.2f", i, j, gotF, wantF)
			}
		}
	}

	fmt.Println("parseAndNormalizeAge tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadParseNormalizeAgeTests(baseDir string) []ParseNormalizeAgeTest {
	inputDir := filepath.Join(baseDir, "Input")
	outputDir := filepath.Join(baseDir, "Output")

	inputFiles, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("Warning: could not read input directory %s: %v\n", inputDir, err)
		return nil
	}

	re := regexp.MustCompile(`^input(\d+)\.txt$`)
	type fileEntry struct {
		idx  int
		path string
	}
	var inputs []fileEntry
	for _, e := range inputFiles {
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

	var tests []ParseNormalizeAgeTest
	for _, in := range inputs {
		outPath := filepath.Join(outputDir, fmt.Sprintf("output%d.txt", in.idx))
		rawAges, err := parseAgeInputFile(in.path)
		if err != nil {
			panic(fmt.Sprintf("Failed to read input file %s: %v", in.path, err))
		}
		want, err := parseAgeOutputFile(outPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to read output file %s: %v", outPath, err))
		}
		tests = append(tests, ParseNormalizeAgeTest{
			RawAges: rawAges,
			Want:    want,
			Name:    fmt.Sprintf("input%d", in.idx),
		})
	}
	return tests
}

func parseAgeInputFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open input file: %w", err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, sc.Err()
}

func parseAgeOutputFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open output file: %w", err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, sc.Err()
}

// ----------------------------------------------------------------------
// transpose()
// ----------------------------------------------------------------------

type TransposeTest struct {
	Input [][]string
	Want  [][]string
	Name  string
}

func TestTranspose(t *testing.T) {
	tests := ReadTransposeTests("Tests/transpose")
	if len(tests) == 0 {
		t.Fatalf("No tests found in Tests/transpose (check Input/Output folders).")
	}

	for ti, test := range tests {
		got := transpose(test.Input)
		if len(got) != len(test.Want) {
			t.Errorf("Test %d (%s): row count mismatch: got %d, want %d", ti, test.Name, len(got), len(test.Want))
			continue
		}
		for r := range got {
			if len(got[r]) != len(test.Want[r]) {
				t.Errorf("Test %d (%s): row %d length mismatch: got %d, want %d", ti, test.Name, r, len(got[r]), len(test.Want[r]))
				continue
			}
			for c := range got[r] {
				if got[r][c] != test.Want[r][c] {
					t.Errorf("Test %d (%s): output[%d][%d] mismatch: got %q, want %q", ti, test.Name, r, c, got[r][c], test.Want[r][c])
				}
			}
		}
	}

	fmt.Println("transpose tested!")
}

// ----------------------------------------------------------------------
// Parsers

func ReadTransposeTests(baseDir string) []TransposeTest {
	inputDir := filepath.Join(baseDir, "Input")
	outputDir := filepath.Join(baseDir, "Output")

	inputFiles, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("Warning: could not read input directory %s: %v\n", inputDir, err)
		return nil
	}

	re := regexp.MustCompile(`^input(\d+)\.txt$`)
	type fileEntry struct {
		idx  int
		path string
	}
	var inputs []fileEntry
	for _, e := range inputFiles {
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

	var tests []TransposeTest
	for _, in := range inputs {
		outPath := filepath.Join(outputDir, fmt.Sprintf("output%d.txt", in.idx))
		inRows, err := parseMatrixFile(in.path)
		if err != nil {
			panic(fmt.Sprintf("Failed to load input file %s: %v", in.path, err))
		}
		wantRows, err := parseMatrixFile(outPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to load output file %s: %v", outPath, err))
		}
		tests = append(tests, TransposeTest{
			Input: inRows,
			Want:  wantRows,
			Name:  fmt.Sprintf("input%d", in.idx),
		})
	}
	return tests
}

// parseMatrixFile reads a matrix (rows of strings) from a file
func parseMatrixFile(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rows [][]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rows = append(rows, strings.Fields(line))
	}
	return rows, sc.Err()
}

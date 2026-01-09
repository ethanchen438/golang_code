package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Regex to extract numerical age value
var ageRegex = regexp.MustCompile(`(\d+\.?\d*)`)

// Define the output and input directory path
const outputDirPath = "/content/drive/MyDrive/Programming for Scientists - Final Project/finalProcessedDatasets"
const inputDirPath = "/content/drive/MyDrive/Programming for Scientists - Final Project/finalUnprocessedDatasets"

// File of our 21k probes used in the model
const probeListFilename = "/content/drive/MyDrive/Programming for Scientists - Final Project/horvath_21k_probes.csv"

// loadRequiredProbes reads a CSV file, finds the "Name" column, and extracts all probe IDs
func loadRequiredProbes(listPath string) (map[string]bool, error) {
	
	probes := make(map[string]bool)
	file, err := os.Open(listPath)
	if err != nil {
		return nil, fmt.Errorf("could not open probe list file (%s). Make sure it exists in the input directory: %w", listPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header from probe list file: %w", err)
	}

	nameIndex := -1
	for i, colName := range header {
		if strings.EqualFold(strings.TrimSpace(colName), "Name") {
			nameIndex = i
			break
		}
	}

	if nameIndex == -1 {
		return nil, fmt.Errorf("probe list CSV must contain a column named 'Name'")
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading probe list records: %w", err)
		}

		if nameIndex < len(record) {
			probeID := strings.TrimSpace(record[nameIndex])
			if probeID != "" {
				probes[probeID] = true
			}
		}
	}

	if len(probes) == 0 {
		return nil, fmt.Errorf("no probes were successfully loaded from the 'Name' column in the list file")
	}

	return probes, nil
}

// parseAndNormalizeAge extracts the numerical age and converts it into years. If converting fails, it returns "NA".
// Input: a string representing the age
// Output: a normalized age string
func parseAndNormalizeAge(rawAge string) string {
	
	rawAge = strings.ToLower(rawAge)

	if strings.Contains(rawAge, "birth") || strings.Contains(rawAge,"Newborn") {
		fmt.Printf(" Assigned age: %s -> 0.00 (years)\n", rawAge)
		return "0.00"
	}

	matches := ageRegex.FindStringSubmatch(rawAge)
	if len(matches) < 2 {
		return "NA"
	}

	ageValStr := matches[1]
	isMonths := strings.Contains(rawAge, "month")
	ageValue, err := strconv.ParseFloat(ageValStr, 64)
	
	if err != nil {
		log.Printf("Error parsing age value '%s': %v", ageValStr, err)
		return "NA"
	}

	if isMonths {
		ageValue /= 12.0
		fmt.Printf(" Converted age: %s (months) -> %.2f (years)\n", rawAge, ageValue)
	}
	return fmt.Sprintf("%.2f", ageValue)
}

// Input: A matrix (2d slice of strings) containing the CpG values
// Output: The original matrix transposed with CpG values (still strings)
func transpose(rows [][]string) [][]string {
	
	if len(rows) == 0 || len(rows[0]) == 0 {
		return nil
	}
	numRows := len(rows)
	numCols := len(rows[0])

	// Create new matrix where dimensions are flipped
	transposed := make([][]string, numCols)
	for i := range transposed {
		transposed[i] = make([]string, numRows)
	}

	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			transposed[j][i] = rows[i][j]
		}
	}
	return transposed
}

// processFile extracts age and matrix data from a single GEO file using our filtered probes.
func processFile(filePath string, requiredProbes map[string]bool) error {
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var dataLines []string
	inTable := false
	var ageLine string

	// Read file to find age and the CpGIsland data matrix
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")

		// Detect start/end of data matrix
		if strings.HasPrefix(line, "!series_matrix_table_begin") {
			inTable = true
			continue
		}
		if strings.HasPrefix(line, "!series_matrix_table_end") {
			inTable = false
			continue
		}

		if inTable {
			dataLines = append(dataLines, line)
			continue
		}

		// Pull the full line that contains age values
		if strings.HasPrefix(line, "!Sample_characteristics_ch1") && strings.Contains(strings.ToLower(line), "age") {
			ageLine = line
		}
	}

	var ages []string
	if ageLine != "" {
		parts := strings.Split(ageLine, "\t")
		for _, part := range parts[1:] {
			val := strings.Trim(part, "\"")
			// Uses our helper function to parse and convert the age if needed
			normalizedAge := parseAndNormalizeAge(val)
			ages = append(ages, normalizedAge)
		}
		fmt.Printf("  Extracted and normalized %d ages from metadata.\n", len(ages))
	} else {
		fmt.Println("  No age line found in metadata. Age column will be 'NA'.")
	}

	// Parse the CpGIsland data matrix
	if len(dataLines) == 0 {
		return fmt.Errorf("no data matrix table found in file")
	}

	csvReader := csv.NewReader(strings.NewReader(strings.Join(dataLines, "\n")))
	csvReader.Comma = '\t'
	rawRows, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("error parsing data table: %w", err)
	}

	if len(rawRows) < 2 {
		return fmt.Errorf("not enough data rows in the table (header + data)")
	}

	header := rawRows[0]
	dataRows := rawRows[1:]

	fmt.Printf("  Raw data table has %d columns and %d data rows.\n", len(header), len(dataRows))

	// Filter rows to only include the required 21k probes
	filteredDataRows := [][]string{}
	probesFoundCount := 0
	filteredDataRows = append(filteredDataRows, header)

	for _, row := range dataRows {
		if len(row) > 0 {
			probeID := strings.TrimSpace(row[0])
			if requiredProbes[probeID] {
				filteredDataRows = append(filteredDataRows, row)
				probesFoundCount++
			}
		}
	}

	rows := filteredDataRows
	if probesFoundCount == 0 {
		return fmt.Errorf("input file contained zero matching probes out of %d required probes", len(requiredProbes))
	}
	fmt.Printf("  Filtered down to %d probes (required 21k subset).\n", probesFoundCount)

	//Check and replace missing matrix values with "0" to fill the matrix if needed to ensure square
	missingCount := 0
	for i := 1; i < len(rows); i++ {
		for j := 1; j < len(rows[i]); j++ {
			value := strings.TrimSpace(rows[i][j])

			isMissing := value == "" ||
				strings.EqualFold(value, "NA") ||
				strings.EqualFold(value, "NaN") ||
				strings.EqualFold(value, "NULL")

			if isMissing {
				rows[i][j] = "0"
				missingCount++
			}
		}
	}
	fmt.Printf("  Replaced %d missing values with '0' (Zero-imputation).\n", missingCount)

	ageRow := make([]string, len(header))
	ageRow[0] = "Age"
	for i := 1; i < len(header); i++ {
		if i-1 < len(ages) {
			ageRow[i] = ages[i-1]
		} else {
			ageRow[i] = "NA" 
		}
	}

	// Remake newRows with Header, AgeRow, and DataRows for transposing
	newRows := [][]string{rows[0], ageRow}
	newRows = append(newRows, rows[1:]...)

	// Transpose the matrix to be ready for elastic net regression (Rows=Samples, Cols=(CpGIsland Values)
	finalRows := transpose(newRows)

	// Write CSV Output to the specified directory
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	csvFilename := filepath.Join(outputDirPath, fmt.Sprintf("%s_matrix.csv", baseName))

	if err := writeCSV(finalRows, csvFilename); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	fmt.Printf("  Output file written: %s\n", csvFilename)
	return nil
}

// writeCSV writes the data in CSV format.
func writeCSV(rows [][]string, filename string) error {
	
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	writer.Comma = ','
	if err := writer.WriteAll(rows); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Main function to run the full age predictor workflow.
func main() {
	// Base directory containing subdirectories (melanoma, type II diabetes, random)
	const baseDir = "/content/drive/MyDrive/Programming for Scientists - Final Project/finalTestSets/"
	// Contains YMean, YStd, and Feature Means/Stds.
	const normsFilename = "/content/drive/MyDrive/Programming for Scientists - Final Project/finalOutputs/final_elastic_net_model.csv"
	// CpGIsland Weights file path
	const weightsFilename = "/content/drive/MyDrive/Programming for Scientists - Final Project/finalOutputs/elasticNetRegression_CpGs.csv"
	// Output directory
	const outputBaseDir = "/content/drive/MyDrive/Programming for Scientists - Final Project/finalOutputs/"

	fmt.Printf("--- Starting Directory-Based Age Prediction Workflow ---\n")

	// 2. Load Model Parameters (Norms and Metadata)
	model, err := loadNormsAndMetadata(normsFilename)
	if err != nil {
		fmt.Printf("Error loading norms and metadata: %v\n", err)
		return
	}

	// 2b. Load CpGIsland Weights 
	if err := loadWeights(weightsFilename, model); err != nil {
		fmt.Printf("Error loading feature weights: %v\n", err)
		return
	}

	// Set the Bias based on the value from elastic net regression 
	model.Bias = 0.0

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputBaseDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Process all files in the base directory
	if err := processDirectory(baseDir, model, outputBaseDir); err != nil {
		fmt.Printf("Fatal error during directory processing: %v\n", err)
	}

	fmt.Println("\n Workflow Complete")
}

// processDirectory iterates through all subdirectories in the base path and processes
// the GEO files found within.
func processDirectory(basePath string, model *ModelConfig, outputBaseDir string) error {
	// Parse through the base directory
	err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Process relevant .txt files 
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".txt") {
			// Extract information for naming: Folder name (ex: "melanoma") and GSE ID
			folderName := filepath.Base(filepath.Dir(path))
			gseID := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())) 

			fmt.Printf("\n--- Processing file: %s in folder: %s ---\n", d.Name(), folderName)

			// Data Processing 
			featureNames, enrichedRows, actualAges, err := extractAndEnrichGEOData(path)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", path, err)
				return nil 
			}

			// Extract Sample IDs for output
			if len(enrichedRows) < 1 {
				fmt.Printf("Warning: No sample data found in %s.\n", path)
				return nil
			}
			sampleIDs := enrichedRows[0][1:]
			if len(sampleIDs) == 0 {
				fmt.Printf("Warning: No sample IDs found in %s.\n", path)
				return nil
			}

			// Age Prediction
			predictions, err := predictAgeFromWeights(featureNames, enrichedRows, model)
			if err != nil {
				fmt.Printf("Error during age prediction for %s: %v\n", path, err)
				return nil 
			}

			// Consolidate results
			results := make([]AgePredictionResult, len(sampleIDs))
			for i := range sampleIDs {
				results[i] = AgePredictionResult{
					SampleID:     sampleIDs[i],
					PredictedAge: predictions[i],
					ActualAge:    actualAges[i],
				}
			}

			// Write Predictions to filename
			outputFilename := filepath.Join(outputBaseDir, fmt.Sprintf("%s_%s_predictions.csv", folderName, gseID))
			if err := writePredictions(results, outputFilename); err != nil {
				fmt.Printf("Error writing predictions for %s: %v\n", outputFilename, err)
			} else {
				fmt.Printf("-> Successfully written all %d predictions to %s\n", len(results), outputFilename)
			}
		}
		return nil
	})
	return err
}


// Parsers
// Written using AI

// extractAndEnrichGEOData reads the GEO matrix file, extracts age metadata, and returns the feature names, the augmented data table, and the actual ages.
func extractAndEnrichGEOData(filename string) (featureNames []string, augmentedRows [][]string, actualAges []string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var dataLines []string
	inTable := false
	var ageLine string
	var gseID string 

	// Separate metadata from data table.
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, nil, fmt.Errorf("error reading line: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")

		// Detect start and end of data table
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

		// Parses for the line that contains age values in the metadata section.
		if strings.HasPrefix(line, "!Sample_characteristics_ch1") && strings.Contains(line, "age") {
			if ageLine == "" {
				ageLine = line
			}
		}

		if strings.HasPrefix(line, "!Series_geo_accession") {
			parts := strings.Split(line, "\t")
			if len(parts) > 1 {
				gseID = strings.Trim(parts[1], " \"")
			}
		}
	}

	// Parse the age line from metadata.
	actualAges, err = parseAgeLine(ageLine)
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Printf("-> Extracted %d actual age values from metadata for %s.\n", len(actualAges), gseID)

	// Parse the main data table.
	if len(dataLines) == 0 {
		return nil, nil, nil, fmt.Errorf("no data table found in file")
	}

	csvReader := csv.NewReader(strings.NewReader(strings.Join(dataLines, "\n")))
	csvReader.Comma = '\t' 
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading data table: %w", err)
	}

	if len(rows) < 2 {
		return nil, nil, nil, fmt.Errorf("not enough data rows in the table")
	}

	header := rows[0]
	fmt.Printf("-> Data table dimensions: %d features (rows) and %d samples (columns, excluding first column).\n", len(rows)-1, len(header)-1)

	// Create and insert the new age row for the prediction step.
	ageRowForPrediction := createAgeRowForPrediction(header, actualAges)
	newRows := [][]string{header, ageRowForPrediction}
	newRows = append(newRows, rows[1:]...)

	// Extract feature names from the first column of data rows
	featureNames = make([]string, len(newRows)-2) 
	for i := 2; i < len(newRows); i++ {
		featureNames[i-2] = newRows[i][0] 
	}
	return featureNames, newRows, actualAges, nil
}

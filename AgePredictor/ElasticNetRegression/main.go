package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

func main() {

	inputDirPath := "/content/drive/MyDrive/Programming for Scientists - Final Project/finalProcessedDatasets"
	outputWeightsFilePath := "/content/drive/MyDrive/Programming for Scientists - Final Project/finalOutputs/elasticNetRegression_CpGs.csv"
	outputNormsFilePath := "/content/drive/MyDrive/Programming for Scientists - Final Project/finalOutputs/final_elastic_net_model.csv"

	fmt.Printf("Starting Elastic Net Hyperparameter Tuning and Regression.\n")
	fmt.Printf("Loading and aggregating data from directory: %s\n", inputDirPath)

	// Load and Aggregate Data from Directory
	data, err := loadAllDataFromDir(inputDirPath)
	if err != nil {
		log.Fatalf("Error loading data from directory: %v", err)
	}

	X := data.Matrix
	y := data.Ages

	fmt.Printf("\nData aggregation complete: %d total samples, %d features (Canonical CpG Islands)\n", len(X), len(X[0]))

	// Setup Cross-Validation Grid Search and Model Parameters
	alphaValues := []float64{0.1, 0.5, 1.0}
	lambdaValues := []float64{0.01, 0.1, 1.0}
	kFolds := 5
	maxIter := 10000
	tol := 1e-4

	fmt.Printf("\nStarting %d-Fold Cross-Validation Grid Search\n", kFolds)
	fmt.Printf("Testing %d parameter combinations (alpha: %.2f, %.2f, %.2f | lambda: %.2f, %.2f, %.2f)\n",
		len(alphaValues)*len(lambdaValues), alphaValues[0], alphaValues[1], alphaValues[2], lambdaValues[0], lambdaValues[1], lambdaValues[2])

	// The best values are determined using cross validation on various possible lambda/alpha combinations and will be used as the final model
	var bestAlpha float64
	var bestLambda float64
	minMSE := math.MaxFloat64

	// Perform Grid Search
	for _, alpha := range alphaValues {
		for _, lambda := range lambdaValues {
			
			avgMSE := crossValidationElasticNet(X, y, alpha, lambda, kFolds, maxIter, tol)
			fmt.Printf(" 	Alpha=%.2f, Lambda=%.2f: Avg MSE = %.4f\n", alpha, lambda, avgMSE)

			if avgMSE < minMSE {
				minMSE = avgMSE
				bestAlpha = alpha
				bestLambda = lambda
			}
		}
	}

	fmt.Printf("\nGrid Search Complete\n")
	fmt.Printf("Optimal Hyperparameters: Alpha = %.2f, Lambda = %.2f\n", bestAlpha, bestLambda)
	fmt.Printf("Best Average MSE: %.4f\n", minMSE)

	// Train final Model on full dataset with updated parameters
	X_final := AddColOfOnes(X)
	normParamsFinal := getNormParams(X_final, y)
	X_norm_final, y_norm_final := applyNormalization(X_final, y, normParamsFinal)
	X_norm_final_XTS := precalculateXTS(X_norm_final)

	fmt.Printf("\nTraining final model with Optimal Alpha=%.2f and Lambda=%.2f...\n", bestAlpha, bestLambda)
	finalWeightsNorm := elasticNetRegression(X_norm_final, y_norm_final, X_norm_final_XTS, bestAlpha, bestLambda, maxIter, tol)

	fmt.Println("\nFinal Regression Results (Normalized Weights)")
	if len(finalWeightsNorm) > 0 {
		fmt.Printf("Intercept (Bias): %.6f\n", finalWeightsNorm[0])
	}

	fmt.Println("Feature Weights:")
	nonZeroCount := 0
	const epsilon = 1e-9
	for i := 1; i < len(finalWeightsNorm); i++ {
		if math.Abs(finalWeightsNorm[i]) > epsilon {
			nonZeroCount++
			featureName := data.CPGIslands[i-1]
			fmt.Printf(" 	%s: %.6f\n", featureName, finalWeightsNorm[i])
		}
	}
	fmt.Printf("\nTotal Non-Zero Features Selected: %d\n", nonZeroCount)

	// Write CpGIsland Weights 
	fmt.Printf("Writing selected features to: %s\n", outputWeightsFilePath)
	if err := writeNonZeroWeightsToCSV(outputWeightsFilePath, data.CPGIslands, finalWeightsNorm); err != nil {
		log.Printf("Error writing weights CSV file: %v", err)
	} else {
		fmt.Println("Successfully saved selected features (weights) to CSV.")
	}

	// Write Normalization Parameters (YMean, YStd, Feature Norms)
	fmt.Printf("Writing normalization parameters to: %s\n", outputNormsFilePath)
	if err := writeNormsAndMetadataToCSV(outputNormsFilePath, normParamsFinal, data.CPGIslands); err != nil {
		log.Fatalf("Error writing norms/metadata CSV file: %v", err)
	} else {
		fmt.Println("Successfully saved normalization parameters (YMean, YStd, Feature Norms) to CSV.")
	}
}

// Parsers
// Written using AI

/*
How data should look:

| SampleID | Age | CpG1 | CpG2 | CpG3 | CpG4 |
| -------- | --- | ---- | ---- | ---- | ---- |
| S1       | 25  | 0.8  | 0.3  | 0.5  | 0.1  |
| S2       | 35  | 0.2  | 0.7  | 0.9  | 0.4  |
| S3       | 40  | 0.5  | 0.6  | 0.3  | 0.8  |

How data should be stored:

X := [][]float64{
    {0.8, 0.3, 0.5, 0.1}, // CpG features for S1
    {0.2, 0.7, 0.9, 0.4}, // CpG features for S2
    {0.5, 0.6, 0.3, 0.8}, // CpG features for S3
}

y := []float64{25, 35, 40} // Ages
*/

// readCSVFileRaw handles file opening, CSV reading, and parsing data from a single file.
// It returns the raw matrix, ages, and the CPG island names found in the header.
func readCSVFileRaw(filePath string) ([][]float64, []float64, []string, error) {
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, nil, nil, fmt.Errorf("file is empty")
		}
		return nil, nil, nil, fmt.Errorf("error reading header: %w", err)
	}

	if len(header) < 3 {
		return nil, nil, nil, fmt.Errorf("header must have at least 'SampleID', 'Age', and one feature column (got %d)", len(header))
	}

	cpGIsls := header[2:]
	var dataMatrix [][]float64
	var ages []float64

	// Read in relevant data rows
	for lineNum := 2; ; lineNum++ { 
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading record on line %d in %s: %v. Skipping.", lineNum, filepath.Base(filePath), err)
			continue
		}

		if len(record) != len(header) {
			log.Printf("Skipping line %d in %s due to inconsistent column count: expected %d, got %d\n", lineNum, filepath.Base(filePath), len(header), len(record))
			continue
		}

		// Parse Age
		age, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			// Checks for "NA" age values which we must skip as they lack a valid target. Often due to missing values from the original file
			log.Printf("Skipping line %d in %s due to invalid Age value: '%s' (Expected float, likely 'NA').\n", lineNum, filepath.Base(filePath), record[1])
			continue
		}

		// Parse CpGIsland weights data
		var rowFeatures []float64
		isValidRow := true
		for i := 2; i < len(record); i++ {
			value, err := strconv.ParseFloat(record[i], 64)
			if err != nil {
				log.Printf("Skipping line %d in %s due to invalid float value in feature '%s': '%s'\n", lineNum, filepath.Base(filePath), cpGIsls[i-2], record[i])
				isValidRow = false
				break
			}
			rowFeatures = append(rowFeatures, value)
		}

		if isValidRow {
			ages = append(ages, age)
			dataMatrix = append(dataMatrix, rowFeatures)
		}
	}

	return dataMatrix, ages, cpGIsls, nil
}

// processAndAlignFile is a worker function that reads, parses, and aligns data from a single file.
func processAndAlignFile(filePath string, canonicalCPGs []string, outChan chan<- AlignedFileData) {
	
	currentFileMatrix, currentFileAges, currentFileCPGs, err := readCSVFileRaw(filePath)

	if err != nil {
		outChan <- AlignedFileData{Error: err, FileName: filepath.Base(filePath)}
		return
	}

	if len(currentFileMatrix) == 0 {
		outChan <- AlignedFileData{Error: fmt.Errorf("contains no valid data rows"), FileName: filepath.Base(filePath)}
		return
	}

	// Map CpGIslands in the current file to their column index in the current file
	currentFileCPGSet := make(map[string]int)
	for i, cpg := range currentFileCPGs {
		currentFileCPGSet[cpg] = i
	}

	nCanonicalFeatures := len(canonicalCPGs)
	alignedMatrix := make([][]float64, len(currentFileMatrix))

	// Align each sample row
	for i, rawRow := range currentFileMatrix {
		alignedRow := make([]float64, nCanonicalFeatures)

		// Iterate over the CpGIsland weights
		for j, canonicalCPG := range canonicalCPGs {
			if currentFileIndex, ok := currentFileCPGSet[canonicalCPG]; ok {
				// Feature is present in current file, use its value
				alignedRow[j] = rawRow[currentFileIndex]
			} else {
				// Feature is missing in current file, replace with 0
				alignedRow[j] = 0.0
			}
		}
		alignedMatrix[i] = alignedRow
	}

	outChan <- AlignedFileData{AlignedMatrix: alignedMatrix, Ages: currentFileAges, FileName: filepath.Base(filePath)}
}

// loadAllDataFromDir scans a directory, reads all CSVs, and aggregates the data concurrently.
func loadAllDataFromDir(dirPath string) (MatrixData, error) {
	
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return MatrixData{}, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var combinedData MatrixData
	var canonicalCPGs []string
	filesToProcess := make([]string, 0)

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".csv" {
			continue
		}

		fullPath := filepath.Join(dirPath, file.Name())

		if canonicalCPGs == nil {
			// Read the first valid CSV file sequentially to establish the final data file.
			log.Printf("Sequentially reading first file to establish schema: %s", file.Name())

			currentFileMatrix, currentFileAges, currentFileCPGs, err := readCSVFileRaw(fullPath)
			if err != nil {
				log.Printf("Skipping file %s due to read error: %v", file.Name(), err)
				continue
			}
			if len(currentFileMatrix) == 0 {
				log.Printf("File %s contains no valid data rows. Skipping.", file.Name())
				continue
			}

			canonicalCPGs = currentFileCPGs
			combinedData.CPGIslands = canonicalCPGs
			combinedData.Matrix = append(combinedData.Matrix, currentFileMatrix...)
			combinedData.Ages = append(combinedData.Ages, currentFileAges...)
			log.Printf("Schema set by %s: %d canonical features.", file.Name(), len(canonicalCPGs))
		} else {
			// Collect all other files for parallel processing
			filesToProcess = append(filesToProcess, fullPath)
		}
	}

	if canonicalCPGs == nil {
		return MatrixData{}, fmt.Errorf("no valid CSV files found to establish canonical schema in %s", dirPath)
	}

	// Parallel Processing for remaining files
	if len(filesToProcess) > 0 {
		log.Printf("\nStarting parallel processing for %d remaining files...", len(filesToProcess))

		var wg sync.WaitGroup
		resultsChan := make(chan AlignedFileData, len(filesToProcess))

		for _, filePath := range filesToProcess {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				processAndAlignFile(path, canonicalCPGs, resultsChan)
			}(filePath)
		}

		wg.Wait()
		close(resultsChan)

		// Combine results from the channel
		for result := range resultsChan {
			if result.Error != nil {
				log.Printf("Skipping file %s due to processing error: %v", result.FileName, result.Error)
				continue
			}

			log.Printf("Successfully aligned %s: %d rows.", result.FileName, len(result.AlignedMatrix))
			combinedData.Matrix = append(combinedData.Matrix, result.AlignedMatrix...)
			combinedData.Ages = append(combinedData.Ages, result.Ages...)
		}
	}

	if len(combinedData.Matrix) == 0 {
		return MatrixData{}, fmt.Errorf("no valid data could be loaded from any CSV files in %s", dirPath)
	}

	return combinedData, nil
}

// writeNormsAndMetadataToCSV writes all our normalization parameters (YMean, YStd, CpGIsland Means/Stds)
func writeNormsAndMetadataToCSV(filePath string, params NormParams, featureNames []string) error {
	
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create norms/metadata output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write Headers (Type, Parameter, Value)
	if err := writer.Write([]string{"Type", "Parameter", "Value"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write Target Variable Norms (YMean, YStd)
	if err := writer.Write([]string{"Metadata", "YMean", fmt.Sprintf("%.6f", params.YMean)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"Metadata", "YStd", fmt.Sprintf("%.6f", params.YStd)}); err != nil {
		return err
	}

	// Write Feature Norms (Means and Stds)
	for i := 1; i < len(params.Means); i++ {
		cpgID := featureNames[i-1] // map index

		meanRecord := []string{"Feature_Norm", "Mean_" + cpgID, fmt.Sprintf("%.6f", params.Means[i])}
		if err := writer.Write(meanRecord); err != nil {
			return err
		}

		stdRecord := []string{"Feature_Norm", "Std_" + cpgID, fmt.Sprintf("%.6f", params.Stds[i])}
		if err := writer.Write(stdRecord); err != nil {
			return err
		}
	}

	return nil
}

// writeNonZeroWeightsToCSV filters non-zero weights and writes them to a CSV file.
func writeNonZeroWeightsToCSV(filePath string, featureNames []string, weights []float64) error {
	
	const epsilon = 1e-9 // Tolerance for treating a weight as zero

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"FeatureName", "NormalizedWeight"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	if len(weights) > 0 && math.Abs(weights[0]) > epsilon {
		if err := writer.Write([]string{"Intercept", fmt.Sprintf("%.6f", weights[0])}); err != nil {
			return fmt.Errorf("failed to write intercept: %w", err)
		}
	}

	// Write CpGIsland weights 
	for i := 1; i < len(weights); i++ {
		weight := weights[i]
		if math.Abs(weight) > epsilon {
			featureName := featureNames[i-1]
			record := []string{featureName, fmt.Sprintf("%.6f", weight)}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write record for %s: %w", featureName, err)
			}
		}
	}

	return nil
}

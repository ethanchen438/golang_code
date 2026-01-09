package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Regex to extract age from input matrix file
var ageRegex = regexp.MustCompile(`(\d+\.?\d*)`)

// Normailization values used from input normalization weight file
type NormParams struct {
	YMean float64
	YStd  float64
}

// All the values being taken into account when predicting age from our model
type ModelConfig struct {
	Norm         NormParams
	FeatureMeans map[string]float64
	FeatureStds  map[string]float64
	Weights      map[string]float64
	Bias         float64
}

// Stores the sample ID, predicted age, and actual age.
type AgePredictionResult struct {
	SampleID     string
	PredictedAge float64
	ActualAge    string
}

// Genereates an age row that will be used for comparison later
func createAgeRowForPrediction(header []string, actualAges []string) []string {

	ageRow := make([]string, len(header))
	ageRow[0] = "Age"

	for i := 1; i < len(header); i++ {
		ageIndex := i - 1
		if ageIndex < len(actualAges) {
			rawAgeStr := actualAges[ageIndex]
			// Check if the age string is a number. If not, use "0.0" as a placeholder.
			if _, err := strconv.ParseFloat(rawAgeStr, 64); err != nil {
				ageRow[i] = "0.0"
			} else {
				ageRow[i] = rawAgeStr
			}
		} else {
			ageRow[i] = "0.0"
		}
	}
	return ageRow
}

// Parsers for our model
// Written using AI

// parseAndNormalizeAge extracts the numerical age and converts it to years if the raw string
// contains the keyword "month". If conversion fails, it returns "NA".
func parseAndNormalizeAge(rawAge string) string {

	rawAge = strings.ToLower(rawAge)

	// Check for "birth" and "newborn" and assign age of 0
	if strings.Contains(rawAge, "birth") || strings.Contains(rawAge, "newborn") {
		return "0.00"
	}

	matches := ageRegex.FindStringSubmatch(rawAge)
	if len(matches) < 2 {
		return "NA"
	}

	ageValStr := matches[1]
	// Check the entire raw string for the keyword "month"
	isMonths := strings.Contains(rawAge, "month")
	ageValue, err := strconv.ParseFloat(ageValStr, 64)

	if err != nil {
		return "NA"
	}

	if isMonths {
		ageValue /= 12.0
	}

	return fmt.Sprintf("%.2f", ageValue)
}

// Further cleans our age entries by removing the label and accounts for any missing values
func parseAgeLine(ageLine string) ([]string, error) {

	if ageLine == "" {
		return nil, fmt.Errorf("age line was not found in metadata")
	}

	var ages []string
	parts := strings.Split(ageLine, "\t")

	for _, part := range parts[1:] {

		val := strings.Trim(part, "\"")
		val = strings.TrimSpace(strings.TrimPrefix(val, "age:"))
		normalizedAge := parseAndNormalizeAge(val)

		//Checks for any missing ages and adds them to our slice
		if normalizedAge == "NA" || normalizedAge == "" {
			ages = append(ages, "NA")
		} else {
			ages = append(ages, normalizedAge)
		}
	}
	return ages, nil
}

// loadNormsAndMetadata reads the normalization and model metadata from the specified CSV files (ex: "final_elastic_net_model.csv").
func loadNormsAndMetadata(filePath string) (*ModelConfig, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open norms/metadata file '%s': %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	model := &ModelConfig{
		FeatureMeans: make(map[string]float64),
		FeatureStds:  make(map[string]float64),
		Weights:      make(map[string]float64),
		Bias:         0.0,
	}

	var yMeanFound, yStdFound bool

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading norms/metadata: %w", err)
	}

	for _, record := range records {
		if len(record) < 3 {
			if len(record) > 0 {
			}
			continue
		}

		if record[0] == "Type" {
			continue
		}

		paramType := record[0]
		parameter := record[1]
		valueStr := record[2]
		value, err := strconv.ParseFloat(valueStr, 64)

		if err != nil {
			continue
		}

		if parameter == "YMean" {
			model.Norm.YMean = value
			yMeanFound = true
		} else if parameter == "YStd" {
			model.Norm.YStd = value
			yStdFound = true
		} else if paramType == "Feature_Norm" {
			parts := strings.SplitN(parameter, "_", 2)
			if len(parts) != 2 {
				continue
			}
			paramTag := parts[0]
			cpgID := parts[1]

			if paramTag == "Mean" {
				model.FeatureMeans[cpgID] = value
			} else if paramTag == "Std" {
				model.FeatureStds[cpgID] = value
			}
		}
	}

	if !yMeanFound {
		return nil, fmt.Errorf("YMean (Age Mean) is missing from the model file. Please ensure a line with 'YMean' in the second column exists")
	}
	if !yStdFound || model.Norm.YStd == 0 {
		return nil, fmt.Errorf("YStd (Age Standard Deviation) is missing or zero in the model file. Please ensure a line with 'YStd' in the second column exists and has a positive value")
	}

	fmt.Printf("-> Loaded Norms/Metadata: YMean=%.2f, YStd=%.2f. Feature Norms: %d.\n",
		model.Norm.YMean, model.Norm.YStd, len(model.FeatureMeans))

	return model, nil
}

// loadWeights reads the feature weights from the "crossvalidated_cpgs.csv" file and fills out the corresponding values in ModelConfig's Weights.
func loadWeights(filePath string, model *ModelConfig) error {

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open weights file '%s': %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading weights file: %w", err)
	}

	if len(records) < 2 || records[0][0] != "FeatureName" || records[0][1] != "NormalizedWeight" {
		return fmt.Errorf("weights file has unexpected format. Expected header: FeatureName,NormalizedWeight (CSV)")
	}

	for _, record := range records[1:] {
		if len(record) < 2 {
			continue
		}

		cpgID := record[0]
		weightStr := record[1]

		weight, err := strconv.ParseFloat(weightStr, 64)
		if err != nil {
			continue
		}

		model.Weights[cpgID] = weight
	}

	if len(model.Weights) == 0 {
		return fmt.Errorf("no feature weights found in the weights file")
	}

	fmt.Printf("-> Loaded Feature Weights: %d weights.\n", len(model.Weights))

	return nil
}

// Parsers
// Written using AI

// Input: Ordered list of CpG names that match up with the dataset columns, 2d slice containing raw CpG values (as strings),
// and list of weights found using the elastic net pipeline
// Output: A list of predicted ages for each of the samples
func predictAgeFromWeights(featureNames []string, rows [][]string, model *ModelConfig) ([]float64, error) {

	if len(rows) < 3 {
		return nil, fmt.Errorf("enriched data table is too small for prediction")
	}

	// rows[0] is Header, rows[1] is the placeholder Age row, rows[2:] are the CpG data.
	dataRows := rows[2:]

	// Transpose and Impute (from Feature x Sample to Sample x Feature)
	features, numSamples, err := transposeAndImpute(dataRows)
	if err != nil {
		return nil, err
	}

	numDataFeatures := len(featureNames)
	if numDataFeatures == 0 || numSamples == 0 {
		return nil, fmt.Errorf("no numerical data available for prediction")
	}

	// Alignment of Features, Normalization, and Prediction
	predictions := make([]float64, numSamples)
	modelFeatureCount := len(model.Weights)

	if modelFeatureCount == 0 {
		return nil, fmt.Errorf("model has no weights loaded; cannot predict age")
	}

	fmt.Printf("\n-> Prepared %d samples with %d features each from the data file.\n", numSamples, numDataFeatures)
	fmt.Printf("-> Using %d features from the model for prediction (Bias=%.4f + %d weights).\n", modelFeatureCount+1, model.Bias, modelFeatureCount)

	for i := 0; i < numSamples; i++ {
		predictedAgeNorm := model.Bias
		sampleFeatures := features[i]

		// Iterate over features in the data file
		for j := 0; j < numDataFeatures; j++ {
			cpgID := featureNames[j]
			dataValue := sampleFeatures[j]

			// Check if this feature is required by the model (has a non-zero weight)
			weight, weightExists := model.Weights[cpgID]
			mean, meanExists := model.FeatureMeans[cpgID]
			std, stdExists := model.FeatureStds[cpgID]

			if weightExists && meanExists && stdExists && std != 0 {
				normalizedValue := (dataValue - mean) / std
				predictedAgeNorm += normalizedValue * weight
			}
		}

		// Denormalize the predicted age back to the original scale
		predictedAge := (predictedAgeNorm * model.Norm.YStd) + model.Norm.YMean

		// Ensure age is non-negative
		if predictedAge < 0 {
			predictedAge = 0.0
		}

		predictions[i] = predictedAge
	}
	return predictions, nil
}

// Input: A matrix (2d slice of strings) containing the CpG values
// Output: The original matrix transposed with CpG values converted to floats
func transposeAndImpute(dataRows [][]string) ([][]float64, int, error) {

	numFeatures := len(dataRows)
	if numFeatures == 0 {
		return nil, 0, nil
	}
	numSamples := len(dataRows[0]) - 1

	// Step 1: Parse and add missing data (NA)
	floatMatrix := make([][]float64, numFeatures)
	featureSums := make([]float64, numFeatures)
	featureCounts := make([]int, numFeatures)

	for i, row := range dataRows {
		floatMatrix[i] = make([]float64, numSamples)
		for j := 0; j < numSamples; j++ {
			val, err := strconv.ParseFloat(row[j+1], 64)
			if err == nil {
				floatMatrix[i][j] = val
				featureSums[i] += val
				featureCounts[i]++
			} else {
				floatMatrix[i][j] = math.NaN()
			}
		}
	}

	// Calculate means and perform imputation
	for i := range floatMatrix {
		var featureMean float64
		if featureCounts[i] > 0 {
			featureMean = featureSums[i] / float64(featureCounts[i])
		} else {
			featureMean = 0.0 // Default to 0 if all values are missing
		}
		// Impute: replace NaNs with the calculated mean for that feature
		for j := 0; j < numSamples; j++ {
			if math.IsNaN(floatMatrix[i][j]) {
				floatMatrix[i][j] = featureMean
			}
		}
	}

	// Transpose the matrix
	outputFeatures := make([][]float64, numSamples)
	for j := 0; j < numSamples; j++ { // Iterate over samples (new rows)
		outputFeatures[j] = make([]float64, numFeatures)
		for i := 0; i < numFeatures; i++ { // Iterate over features (new columns)
			outputFeatures[j][i] = floatMatrix[i][j]
		}
	}

	return outputFeatures, numSamples, nil
}

// writePredictions writes the sample IDs, predicted ages, and actual ages to a CSV file.
func writePredictions(results []AgePredictionResult, filename string) error {

	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	writer.Comma = ','

	// Write header row
	if err := writer.Write([]string{"Sample_ID", "Predicted_Age", "Actual_Age"}); err != nil {
		return err
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.SampleID,
			fmt.Sprintf("%.4f", result.PredictedAge),
			result.ActualAge,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

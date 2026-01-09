package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// Copy and paste into main to use!!
// main() --> Reads in the predicted and chronological age from two csv files --> Calculates the testing measures Pearson, Median Absolute Error, and Mean Error between the two
func main() {

	//Read in predicted ages
	predicted, err := readFloatColumn("predicted.csv", true)
	if err != nil {
		panic(err)
	}

	//Read in chronological ages
	chronological, err := readFloatColumn("chronological.csv", true)
	if err != nil {
		panic(err)
	}

	//Print out the results
	fmt.Println("Pearson:", pearsonCorrelation(predicted, chronological))
	fmt.Println("Median AE:", medianAbsoluteError(predicted, chronological))
	fmt.Println("Mean Error:", meanError(predicted, chronological))
}

// Made using AI
// readColumnCSV() --> Reads a CSV file containing a single column of numbers. If the csv file has a header, the first row will be skipped.
// Input: A string, filename, containing the file name and/or path to the file containing the csv file to be read in.
// Output: A slice of float64s containing the read in numbers from the csv file. An error indicating if there was a problem.
func readFloatColumn(filename string, hasHeader bool) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read csv: %w", err)
	}

	startRow := 0
	if hasHeader {
		startRow = 1
	}

	values := make([]float64, 0, len(records)-startRow)

	for i := startRow; i < len(records); i++ {
		if len(records[i]) == 0 {
			continue
		}

		v, err := strconv.ParseFloat(records[i][0], 64)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid float on row %d (%q): %w",
				i+1,
				records[i][0],
				err,
			)
		}

		values = append(values, v)
	}

	return values, nil
}

package main

import (
	"fmt"
	"log"
)

// Copy and paste into main to use!!
// main() --> Reads in the processed dataset from a csv, normalizes all the samples inside , and saves the normalzied data into a new csv
func main() {
	// Load the probe design (columns = CpGs)
	design := []int{ /* 1 or 2 per CpG */ }

	// Read beta-values CSV (rows = samples, columns = CpGs)
	betaMatrix, err := readCSV("beta_values.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Normalize all samples
	normalizedMatrix, err := BMIQAllSamples(betaMatrix, design)
	if err != nil {
		log.Fatal(err)
	}

	//Save normalized matrix to CSV
	err = writeCSV("normalized_beta_values.csv", normalizedMatrix)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Normalization complete. Saved to normalized_beta_values.csv")
}

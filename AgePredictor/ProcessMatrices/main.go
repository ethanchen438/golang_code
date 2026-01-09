package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	
	fmt.Printf("Starting data processing from directory: %s\n", inputDirPath)
	fmt.Printf("Writing output to directory: %s\n", outputDirPath)

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDirPath, 0755); err != nil {
		log.Fatalf("Error creating output directory %s: %v", outputDirPath, err)
	}

	files, err := os.ReadDir(inputDirPath)
	if err != nil {
		log.Fatalf("Error reading directory %s: %v", inputDirPath, err)
	}

	requiredProbes, err := loadRequiredProbes(probeListFilename)
	if err != nil {
		log.Fatalf("Fatal error loading required probes: %v", err)
	}
	fmt.Printf("Successfully loaded %d required probes for filtering from %s.\n", len(requiredProbes), probeListFilename)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		fullPath := filepath.Join(inputDirPath, fileName)

		if !strings.HasSuffix(strings.ToLower(fileName), ".txt") {
			log.Printf("Skipping file %s: Not a .txt file.", fileName)
			continue
		}

		fmt.Printf("\n--- Processing file: %s ---\n", fileName)
		err := processFile(fullPath, requiredProbes)
		if err != nil {
			log.Printf("Failed to process %s: %v", fileName, err)
		} else {
			fmt.Printf("Successfully processed %s.\n", fileName)
		}
	}
}

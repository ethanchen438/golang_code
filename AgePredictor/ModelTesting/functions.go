package main

import (
	"math"
	"sort"
)

// pearsonCorrelation() --> Computes the Pearson Correlation Coefficient between predicted and chronological age from a series of samples
// Input: Two slices of float64s containing the predicted and chronological ages of a series of samples. The indices for the predicted and chronological ages must match samplewise
// Ouput: A float64 for the Pearson Correlation Coefficient
func pearsonCorrelation(predicted, chronological []float64) float64 {
	//If the length of the slices are not equal or one of them is 0 --> Panic
	if len(predicted) != len(chronological) || len(predicted) == 0 || len(chronological) == 0 {
		panic("Slices must be the same length and non-empty.")
	}

	n := float64(len(predicted))

	//Loop over each pair of values and calculate the following sums
	sumPredicted, sumChronological, sumProduct, sumPredicted2, sumChronological2 := 0.0, 0.0, 0.0, 0.0, 0.0
	for i := range predicted {
		x := predicted[i]
		y := chronological[i]

		sumPredicted += x          //Sum of predicted ages
		sumChronological += y      //Sum of chronological ages
		sumProduct += x * y        //Sum of the products of predicted and chronological
		sumPredicted2 += x * x     //Sum of predicted ages squared
		sumChronological2 += y * y //Sum of chronological age squared
	}

	//Calculate Pearson Correlation Coefficient --> Covariance / Product of Standard Deviations
	numerator := n*sumProduct - sumPredicted*sumChronological
	denominator := math.Sqrt((n*sumPredicted2 - sumPredicted*sumPredicted) * (n*sumChronological2 - sumChronological*sumChronological))

	//If the denominator is 0 --> Panic --> Cannot divide by 0
	if denominator == 0 {
		panic("Denominator zero, cannot compute correlation.")
	}

	return numerator / denominator
}

// medianAbsoluteError() --> Computes the median absolute error between predicted and chronological age from a series of samples
// Input: Two slices of float64s containing the predicted and chronological ages of a series of samples. The indices for the predicted and chronological ages must match samplewise
// Ouput: A float64 for the median absolute error
func medianAbsoluteError(predicted, chronological []float64) float64 {
	//If the length of the slices are not equal or one of them is 0 --> Panic
	if len(predicted) != len(chronological) || len(predicted) == 0 || len(chronological) == 0 {
		panic("Slices must be the same length and non-empty.")
	}

	//Calculate the absolute difference for each sample
	absDiffs := make([]float64, len(predicted))
	for i := range predicted {
		absDiffs[i] = math.Abs(predicted[i] - chronological[i])
	}

	//Sort through the absolute differences
	sort.Float64s(absDiffs)

	//Find the median's index
	mid := len(absDiffs) / 2

	//If the slices were even in legnth --> Calculate average of the two middle absolute differences
	if len(absDiffs)%2 == 0 {
		return (absDiffs[mid-1] + absDiffs[mid]) / 2.0

		//If the slices were odd --> Return median
	} else {
		return absDiffs[mid]
	}
}

// meanError() --> Computes the mean difference between predicted and chronological age from a series of samples
// Input: Two slices of float64s containing the predicted and chronological ages of a series of samples. The indices for the predicted and chronological ages must match samplewise
// Ouput: A float64 for the mean error
func meanError(predicted, chronological []float64) float64 {
	//If the length of the slices are not equal or one of them is 0 --> Panic
	if len(predicted) != len(chronological) || len(predicted) == 0 || len(chronological) == 0 {
		panic("Slices must be the same length and non-empty.")
	}

	//Calculate the total difference
	sum := 0.0
	for i := range predicted {
		sum += predicted[i] - chronological[i]
	}

	//Return the mean difference
	return sum / float64(len(predicted))
}

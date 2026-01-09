package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// Generates a random number for random fold assignment
func init() {
	rand.Seed(time.Now().UnixNano())
}

// NormParams stores the mean and standard deviation for normalization/denormalization.
type NormParams struct {
	// Stores all x means 
	Means []float64
	// Stores all x standard deviations 
	Stds []float64
	// Stores y mean (age)
	YMean float64
	// Stores y SD (age)
	YStd float64
}

// MatrixData holds the combined CpGIsland matrix, ages, and feature names.
type MatrixData struct {
	// Combined matrix including both CpG methylation and ages for each sample
	Matrix [][]float64
	// Slice containing list of all ages
	Ages []float64
	// Slice containing list of all CpGs
	CPGIslands []string
}

// AlignedFileData is used to return results from parallel file processing.
type AlignedFileData struct {
	AlignedMatrix [][]float64
	Ages          []float64
	Error         error
	FileName      string
}

// Input: a 2D-slice of floats of size n x m. In this case, it will be our CpG matrix
// Output: a 2D-slice of floats of size n x m+1, where first column is a column of 1.0s
func AddColOfOnes(mtx [][]float64) [][]float64 {

	rows := len(mtx)
	if rows == 0 {
		panic("matrix has no rows!")
	}

	cols := len(mtx[0])
	result := make([][]float64, rows)

	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols+1)
		result[i][0] = 1.0

		for j := 0; j < cols; j++ {
			result[i][j+1] = mtx[i][j]
		}
	}
	return result
}

// softThreshold performs the soft-thresholding operator used in the Coordinate Descent optimization.
// Input: a weight value, r, and a threshold value, a
// Output: 0, or a shurnken value of the weight
func softThreshold(r, a float64) float64 {
	// Shrinks weights that are greater than the threshold
	if r < -a {
		return r + a
	} else if r > a {
		return r - a
	}
	// Eliminates coefficient below the threshold
	return 0
}

// precalculateXTS computes the sum of squares (X^T * X diagonal) for each feature column.
// Input: a 2D-slice. In our case, this will be our matrix of CpG values
// Output: A slice of the sum of squares from the matrix
func precalculateXTS(X [][]float64) []float64 {

	nSamples := len(X)

	if nSamples == 0 || len(X[0]) == 0 {
		return nil
	}

	nFeatures := len(X[0])

	xts := make([]float64, nFeatures)

	// Loops through different columns and calculates sum of squares
	for j := 0; j < nFeatures; j++ {
		var S float64
		for i := 0; i < nSamples; i++ {
			S += X[i][j] * X[i][j]
		}
		xts[j] = S
	}

	return xts
}

// getNormParams calculates the mean and standard deviation for normalization.
// Input: A 2D-slice of floats (our CpG matrix) and a slice of floats (vector of ages)
// Output: The means and standard deviations of the inputted matrix and slice
func getNormParams(X [][]float64, y []float64) NormParams {

	nSamples := len(X)
	if nSamples == 0 {
		panic("no rows in matrix")
	}

	nFeatures := len(X[0])
	params := NormParams{
		Means: make([]float64, nFeatures),
		Stds:  make([]float64, nFeatures),
	}

	for i := 1; i < nFeatures; i++ {

		var sum, sumSquare float64
		for j := 0; j < nSamples; j++ {
			sum += X[j][i]
			sumSquare += X[j][i] * X[j][i]
		}

		mean := sum / float64(nSamples)
		variance := (sumSquare / float64(nSamples)) - (mean * mean)

		if variance < 0 {
			variance = 0
		}

		std := math.Sqrt(variance)

		if std == 0 {
			std = 1.0
		}

		params.Means[i] = mean
		params.Stds[i] = std
	}

	// Calculate parameters for target vector y (age)
	var ySum, ySumSquare float64
	for _, val := range y {
		ySum += val
		ySumSquare += val * val
	}

	params.YMean = ySum / float64(nSamples)
	yVariance := (ySumSquare / float64(nSamples)) - (params.YMean * params.YMean)
	params.YStd = math.Sqrt(yVariance)

	if params.YStd == 0 {
		params.YStd = 1.0
	}

	return params
}

// applyNormalization normalizes X and y using the provided parameters.
// Input: Matrix X and vector y and normalization parameters (from getNormParams)
// Output: A normalized matrix X and vector y
func applyNormalization(X [][]float64, y []float64, params NormParams) ([][]float64, []float64) {

	nSamples := len(X)
	nFeatures := len(X[0])

	XNorm := make([][]float64, nSamples)
	for i := range XNorm {
		XNorm[i] = make([]float64, nFeatures)
		copy(XNorm[i], X[i])
	}
	yNorm := make([]float64, nSamples)
	copy(yNorm, y)

	// Normalize features 
	for i := 1; i < nFeatures; i++ {
		for j := 0; j < nSamples; j++ {
			XNorm[j][i] = (XNorm[j][i] - params.Means[i]) / params.Stds[i]
		}
	}

	// Normalize vector y
	for i := range yNorm {
		yNorm[i] = (yNorm[i] - params.YMean) / params.YStd
	}

	return XNorm, yNorm
}

// denormalizePrediction scales a prediction back to the original age scale.
// Input: The yNorm after the age vector was normalized and the paramters from normParams
// Output: The original age value before normalization
func denormalizePrediction(yNorm float64, params NormParams) float64 {
	return yNorm*params.YStd + params.YMean
}

// elasticNetRegression performs Elastic Net using optimized Coordinate Descent.
// Input: feature matrix X and age vector y, optimal alpha and lambda values (from crossValidation),
// precalculated sum of squares, a tolerance value, and max iterations for the coordinate descent loop
// Output: The optimal weights (slice of floats) for our regression model
func elasticNetRegression(X [][]float64, y []float64, XTS []float64, alpha float64, lambda float64, maxIter int, tol float64) []float64 {

	nSamples := len(X)
	nFeatures := len(X[0])
	weights := make([]float64, nFeatures)

	// Our default residuals will be equal to the initial ages, which is the greatest possible value
	residuals := make([]float64, nSamples)
	copy(residuals, y)

	// Outer loop: run iterations of coordinate descent until changes to the residual are small enough
	// or we reach max iterations
	for n := 0; n < maxIter; n++ {
		var maxChange float64

		// Inner loop: loop over each weight (column) and update it
		for j := 0; j < nFeatures; j++ {

			oldW := weights[j]
			// Normalize  the sum of squares
			S := XTS[j] / float64(nSamples)

			// Computes the correlation between the current weight and current residuals
			var r float64
			for i := 0; i < nSamples; i++ {
				r += X[i][j] * residuals[i]
			}

			// Average squared magnitude of current feature (j)
			rPrime := r/float64(nSamples) + oldW*S

			var newW float64

			if j == 0 {
				// Intercept not penalized by L1/L2 penalty
				if S == 0 {
					newW = oldW
				} else {
					newW = rPrime / S
				}
			} else {
				// Features penalized using Elastic Net formula
				// L1 (Lasso) performed by soft thresholding, L2 (Ridge) shrinkage done in denominator
				newW = softThreshold(rPrime, lambda*alpha) / (S + lambda*(1-alpha))
			}

			// Calculate how much the residual changes with the current weight
			change := math.Abs(oldW - newW)
			if change > maxChange {
				maxChange = change
			}

			// Update the current weight
			weights[j] = newW

			// Update the residual
			if change != 0 {
				wDiff := newW - oldW
				for i := 0; i < nSamples; i++ {
					residuals[i] -= X[i][j] * wDiff
				}
			}
		}

		// If the change to the residuals is low enough, we can exit the loop
		if maxChange < tol {
			// At this points our weights should be highly optimized
			break
		}
	}

	return weights
}

// meanSquaredError calculates the Mean Squared Error between predicted and actual values.
// We use this function to determine the accuracy of our model
// Input: A slice of predicted ages and ground truth ages
// Output: The MSE (float64) between the two
func meanSquaredError(yActual, yPredicted []float64) float64 {

	//  Exit function if we are missing data or have no data
	if len(yActual) != len(yPredicted) || len(yActual) == 0 {
		return 0.0
	}

	var sumSquaredError float64

	for i := range yActual {
		diff := yActual[i] - yPredicted[i]
		sumSquaredError += diff * diff
	}

	return sumSquaredError / float64(len(yActual))
}

// getFolds splits the dataset indices into K equally sized folds, after shuffling.
// Input: The number of samples we have and how many folds we want to split it into
// Output: A 2D-slice of the folds
func getFolds(nSamples, k int) [][]int {

	foldSize := nSamples / k
	folds := make([][]int, k)

	indices := make([]int, nSamples)
	for i := range indices {
		indices[i] = i
	}

	// Shuffle the indices for valid K-Fold Cross-Validation
	rand.Shuffle(nSamples, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Create folds from the shuffled indices
	for i := 0; i < k; i++ {
		start := i * foldSize
		end := start + foldSize

		// Ensure the last fold takes all remaining samples
		if i == k-1 {
			end = nSamples
		}

		folds[i] = make([]int, len(indices[start:end]))
		copy(folds[i], indices[start:end])
	}

	return folds
}

// getFoldsDeterministic is the same as getFolds, but uses a deterministic
// shuffle instead of having it be fully random.
// This function is NOT used in the full pipeline and is only used for testing getFolds.
func getFoldsDeterministic(nSamples, k int) [][]int {
	
	foldSize := nSamples / k
	folds := make([][]int, k)

	indices := make([]int, nSamples)
	for i := range indices {
		indices[i] = i
	}

	// Use deterministic shuffle for testing
	r := rand.New(rand.NewSource(42))
	r.Shuffle(nSamples, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	for i := 0; i < k; i++ {
		start := i * foldSize
		end := start + foldSize
		if i == k-1 {
			end = nSamples
		}
		folds[i] = make([]int, len(indices[start:end]))
		copy(folds[i], indices[start:end])
	}

	return folds
}

// predict calculates predictions based on the features X (with intercept) and weights.
// This is the function that actually uses the model (weights) found using elastic net to find the biological age
// Input: A dataset containing (a) sample(s) and corresponding CpGs methylation value
// Output: A slice containing the predicted age for each sample
func predict(X [][]float64, weights []float64) []float64 {

	// Ensures we have intercept in case we don't
	// We will never have DNA methylation of 1.0, so this is safe for detecting presence of itercept
	if X[0][0] != 1.0 {
		AddColOfOnes(X)
	}

	nSamples := len(X)
	yPredicted := make([]float64, nSamples)

	for i := 0; i < nSamples; i++ {
		var prediction float64
		for j := range weights {
			prediction += X[i][j] * weights[j]
		}
		yPredicted[i] = prediction
	}
	return yPredicted
}

// crossValidationElasticNet runs K-fold CV concurrently for a single set of hyperparameters.
// Each cross validation runs the entire elastic net pipeline to find an MSE
// Input: feature matrix X and age vector y, optimal alpha and lambda values (from crossValidation),
// precalculated sum of squares, a tolerance value, and max iterations for the coordinate descent loop,
// and lastly, a k value to determine the amount of folds
// Output: The MSE for the trained model
func crossValidationElasticNet(X [][]float64, y []float64, alpha float64, lambda float64, k int, maxIter int, tol float64) float64 {
	
	nSamples := len(X)
	folds := getFolds(nSamples, k)

	// Add intercept column to the full dataset before cross-validation starts
	X_with_ones := AddColOfOnes(X)

	var wg sync.WaitGroup
	results := make(chan float64, k)

	// Loop over each fold (K iterations)
	for i := 0; i < k; i++ {
		wg.Add(1)

		go func(foldIndex int) {
			defer wg.Done()

			testIndices := folds[foldIndex]

			// Determine training indices
			trainIndices := make([]int, 0, nSamples-len(testIndices))
			allIndicesSet := make(map[int]bool)
			for _, idx := range testIndices {
				allIndicesSet[idx] = true
			}
			for idx := 0; idx < nSamples; idx++ {
				if !allIndicesSet[idx] {
					trainIndices = append(trainIndices, idx)
				}
			}

			// Create Train and Test splits
			XTrainRaw := make([][]float64, len(trainIndices))
			YTrainRaw := make([]float64, len(trainIndices))
			for j, idx := range trainIndices {
				XTrainRaw[j] = X_with_ones[idx]
				YTrainRaw[j] = y[idx]
			}

			XTestRaw := make([][]float64, len(testIndices))
			YTestRaw := make([]float64, len(testIndices))
			for j, idx := range testIndices {
				XTestRaw[j] = X_with_ones[idx]
				YTestRaw[j] = y[idx]
			}

			// Calculate normalization parameters based on the training set
			normParams := getNormParams(XTrainRaw, YTrainRaw)

			// Normalize training and test sets
			XTrainNorm, YTrainNorm := applyNormalization(XTrainRaw, YTrainRaw, normParams)
			XTestNorm, _ := applyNormalization(XTestRaw, YTestRaw, normParams)

			// Pre-calculate sum of squares for optimization
			XTrainNormXTS := precalculateXTS(XTrainNorm)

			// Train the model using elastic net regression
			weightsNorm := elasticNetRegression(XTrainNorm, YTrainNorm, XTrainNormXTS, alpha, lambda, maxIter, tol)

			// Predict age on normalized test set
			YTestPredNorm := predict(XTestNorm, weightsNorm)

			// Denormalize predictions
			YTestPredOriginal := make([]float64, len(YTestPredNorm))
			for j, predNorm := range YTestPredNorm {
				YTestPredOriginal[j] = denormalizePrediction(predNorm, normParams)
			}

			mse := meanSquaredError(YTestRaw, YTestPredOriginal)
			results <- mse
		}(i)
	}

	wg.Wait()
	close(results)

	var totalMSE float64
	for mse := range results {
		totalMSE += mse
	}

	return totalMSE / float64(k)
}

// crossValidationElasticNetDeterministic is the same as crossValidationElasticNet
// but it uses getFoldsDeterministic as a helper instead of getFolds
// This function is NOT used in the full pipeline and is only used for testing crossValidationElasticNet.
func crossValidationElasticNetDeterministic(X [][]float64, y []float64, alpha float64, lambda float64, k int, maxIter int, tol float64) float64 {
	
	nSamples := len(X)
	folds := getFoldsDeterministic(nSamples, k)

	// Add intercept column to the full dataset before CV starts
	X_with_ones := AddColOfOnes(X)

	var wg sync.WaitGroup
	results := make(chan float64, k)

	// Loop over each fold (K iterations)
	for i := 0; i < k; i++ {
		wg.Add(1)

		go func(foldIndex int) {
			defer wg.Done()

			testIndices := folds[foldIndex]

			// Determine training indices
			trainIndices := make([]int, 0, nSamples-len(testIndices))
			allIndicesSet := make(map[int]bool)
			for _, idx := range testIndices {
				allIndicesSet[idx] = true
			}
			for idx := 0; idx < nSamples; idx++ {
				if !allIndicesSet[idx] {
					trainIndices = append(trainIndices, idx)
				}
			}

			// Create Train and Test splits
			XTrainRaw := make([][]float64, len(trainIndices))
			YTrainRaw := make([]float64, len(trainIndices))
			for j, idx := range trainIndices {
				XTrainRaw[j] = X_with_ones[idx]
				YTrainRaw[j] = y[idx]
			}

			XTestRaw := make([][]float64, len(testIndices))
			YTestRaw := make([]float64, len(testIndices))
			for j, idx := range testIndices {
				XTestRaw[j] = X_with_ones[idx]
				YTestRaw[j] = y[idx]
			}

			// Calculate normalization parameters based on the training set
			normParams := getNormParams(XTrainRaw, YTrainRaw)

			// 2. Normalize training and test sets
			XTrainNorm, YTrainNorm := applyNormalization(XTrainRaw, YTrainRaw, normParams)
			XTestNorm, _ := applyNormalization(XTestRaw, YTestRaw, normParams)

			// Pre-calculate sum of squares for optimization
			XTrainNormXTS := precalculateXTS(XTrainNorm)

			// Train the model using elastic net regression
			weightsNorm := elasticNetRegression(XTrainNorm, YTrainNorm, XTrainNormXTS, alpha, lambda, maxIter, tol)

			// Predict age on normalized test set
			YTestPredNorm := predict(XTestNorm, weightsNorm)

			// Denormalize predictions
			YTestPredOriginal := make([]float64, len(YTestPredNorm))
			for j, predNorm := range YTestPredNorm {
				YTestPredOriginal[j] = denormalizePrediction(predNorm, normParams)
			}

			mse := meanSquaredError(YTestRaw, YTestPredOriginal)
			results <- mse
		}(i)
	}

	wg.Wait()
	close(results)

	var totalMSE float64
	for mse := range results {
		totalMSE += mse
	}

	return totalMSE / float64(k)
}

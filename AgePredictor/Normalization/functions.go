package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"

	"gonum.org/v1/gonum/optimize"
	"gonum.org/v1/gonum/stat/distuv"
)

type matrix [][]float64

type logLikelihoodMatrix struct {
	values  []float64
	weights []float64
}

// ---------- EM mixture for Beta distributions ----------
type EMResult struct {
	Alpha            []float64   // alpha parameters per component
	Beta             []float64   // beta parameters per component
	StateProportion  []float64   // mixture weights
	Responsibilities [][]float64 // n x K
}

//The following code starts with the most basic functions and ends with the "master" top-level functions

// preprocessingClamp() --> Clamps extreme zeroes and ones (values that are very close to 0 and 1) to our adjustment value. Important for preventing extreme zeroes and ones from breaking our calculations. Many of these values will be rounded to 0 or 1 due to floating point precision, and they are by nature below the measurement precision of the methylation assays. If unchanged, this will cause issues as we take a the log of them, leading to either NaN or -Inf values.
// Input: Two float64s, value is the number being clamped, and adjustment is the number being clamped to.
// Output: A float64 after a value has possibly been clamped
func preprocessingClamp(value, adjustment float64) float64 {
	//If value is extremely close to 0 --> Set to adjustment
	if value < adjustment {
		return adjustment
	}
	//If value is extremely close to 1 --> Set to 1 - adjustment
	if value > 1-adjustment {
		return 1 - adjustment
	}
	return value
}

// weightedMean() --> Calculates the weighted mean of a slice of sample data
// Input: Three slices of float64s. data is the sample data, responsibilities is the array of probabilities that each data value belongs to a state (U, M, H), and weights is the possible weighting each data value should be given (default to 1)
// Output: A float64 referring to the weighted mean of the sample data
func weightedMean(data, responsibiltiies, weights []float64) float64 {

	//Edge case --> Data is empty --> Return NaN
	if len(data) == 0 {
		return math.NaN()
	}

	numerator := 0.0
	denom := 0.0

	//Weighted Mean = sum(data * resp * weight) / sum(resp * weight)
	for i := range data {
		denom += responsibiltiies[i] * weights[i]
		numerator += data[i] * responsibiltiies[i] * weights[i]
	}

	return numerator / denom
}

// weightedSampleVariance() --> Calculates the weighted sample variance of a slice of sample data
// Input: Three slices of float64s. data is the sample data, responsibilities is the array of probabilities that each data value belongs to a state (U, M, H), and weights is the possible weighting each data value should be given (default to 1)
// Output: A float64 referring to the weighted variance of the sample data
func weightedSampleVariance(data, responsibilities, weights []float64, mean float64) float64 {

	//Edge case --> Data is empty --> Return 0
	if len(data) <= 1 {
		return 0
	}

	//Weigthed Sample Variance = E[X^2] - E[X]^2 = E[resp * weight * data^2] - mean^2
	numerator := 0.0
	denom := 0.0
	for i := range data {
		denom += weights[i] * responsibilities[i]
		numerator += weights[i] * responsibilities[i] * data[i] * data[i]
	}
	return numerator/denom - mean*mean

}

// methodOfMoments() --> Calculates the initial guesses for the logLikelihood() function's alpha and beta inputs. Uses the method of moments approach for beta distributions
// Input: Three slices of float64s. data is the sample data, responsibilities is the array of probabilities that each data value belongs to a state (U, M, H), and weights is the possible weighting each data value should be given (default to 1)
// Output: A slice of float64s containing two starting guesses for a beta-distribution that is being expectation maximized. Specifically, the log(alpha) and log(beta)
func methodOfMoments(data, responsibilities, weights []float64) []float64 {

	mean := weightedMean(data, responsibilities, weights)
	variance := weightedSampleVariance(data, responsibilities, weights, mean)

	//Need to put into an array --> Optimize package function takes an array
	//Need to use log --> Optimize package works in log-space
	guess := []float64{
		math.Log(mean * (mean*(1-mean)/variance - 1)),
		math.Log((1 - mean) * (mean*(1-mean)/variance - 1)),
	}

	return guess
}

// betaLogPDF --> Computes the log probability density (y/height) at a given value (x) for a given beta-distribution
// Input: Three float64s, alpha and beta for the parameters defining a beta-distribution, and value for the x-coordinate of a beta-distribtuion
// Output: A float64 for the log probability density of beta-distribution at x=value
func betaLogPDF(alpha, beta, value float64) float64 {
	//Formula: log of betaPDF = (alpha - 1)log(x) + (beta - 1)log(1 - x) - log(alpha, beta)
	//log(alpha, beta) = log(Gamma(alpha)) + log(Gamma(beta)) - log(Gamma(alpha + beta))
	a, _ := math.Lgamma(alpha)
	b, _ := math.Lgamma(beta)
	c, _ := math.Lgamma(alpha + beta)
	lnBeta := a + b - c

	return (alpha-1)*math.Log(value) + (beta-1)*math.Log(1-value) - lnBeta
}

// betaCDF() -->  Calculate the cumulative probability up to a given value for a given beta-distirbution
// Input: Three float64s, alpha and beta for the parameters defining a beta-distribution, and value for the x-coordinate of a beta-distribtuion
// Output: A float64 for the cumulative probability of the beta-distirbution up till x=value
func betaCDF(alpha, beta, value float64) float64 {
	return distuv.Beta{Alpha: alpha, Beta: beta}.CDF(value)
}

// fasterBetaQuantile() --> Computes the x-value for a specific percentile in a defined beta-distribution
// Input: Three float64s, alpha and beta for the parameters defining a beta-distribution, and percentile referring to the percentile we want a corresponding x-value for in the beta-distribution
// Output: A float64 referring to the x-value for the given percentile of the defined beta-distirbution
func betaQuantile(alpha, beta, percentile float64) float64 {
	return distuv.Beta{Alpha: alpha, Beta: beta}.Quantile(percentile)
}

// logLikelihood() --> Computes the negative weighted log likelihood of observing a distribution of beta values for a single sample under a beta distribution with specific alpha and beta paramters
// Input: Two float64s, alpha and beta, as the parameters defining a beta-distribution. Three slices of float64s, sampleData is a single sample's dataset, weights is the possible weighting each data value should be given (default to 1), and responsibilities is the array of probabilities that each data value belongs to a state (U, M, H).
// Ouput: The negative weighted log likelihood for a single sample with the defined beta-distribution
func logLikelihood(alpha, beta float64, sampleData, weights, responsibilities []float64) float64 {

	//Edge case: Make sure alpha and beta are positive as beta distributions only take positive parameters --> If not, return positive infinity
	//Should never need to happen as the function that calls this already takes the exp of a log --> Just in case
	if alpha <= 0 || beta <= 0 {
		return math.Inf(1)
	}

	logLike := 0.0 //Initialize log likelihood

	//Loop through all samples for a single CpG Island
	for i := range sampleData {
		//if CpGData[i] <= 0 || CpGData[i] >= 1 { //Edge case, if Beta-value is not calibrated --> Should never be given that we calibrate earlier
		//continue
		//}
		logLike += weights[i] * responsibilities[i] * betaLogPDF(alpha, beta, sampleData[i]) //Add the weighted loglikelihood for each data point --> Log likelihood takes sum not product
	}

	return -logLike //Return negative log likelihood --> Because the optimizer is built to find minimums not maximums.
}

// betaDistEstimation() --> Estimates a beta-distribution that fits a given sample. Uses maximum likelihood estimation via nelder-mead.
// Input: Three slices of float64s. data is the sample data, responsibilities is the array of probabilities that each data value belongs to a state (U, M, H), and weights is the possible weighting each data value should be given (default to 1)
// Output: Two float64s, alpha and beta, as the parameters defining the estimated beta-distribution for the sample
func betaDistEstimation(sampleData []float64, responsibilities []float64, weights []float64) (float64, float64) {
	// If a CpG island has beta-value = NaN --> Remove the data, responsibility, and weight for the CpG island from all three arrays
	dataObserv := make([]float64, 0, len(sampleData)) //Preassign array length --> Conserve memory
	respObserv := make([]float64, 0, len(responsibilities))
	weightObserv := make([]float64, 0, len(weights))

	for i := range sampleData {
		if !math.IsNaN(sampleData[i]) {
			dataObserv = append(dataObserv, sampleData[i])
			respObserv = append(respObserv, responsibilities[i])
			weightObserv = append(weightObserv, weights[i])
		}
	}

	// Edge case: If the updated arrays are length 1 --> Beta distribution is a straight line
	if len(dataObserv) <= 1 {
		return 1.0, 1.0
	}

	//Set up an initial guess for the beta and alpha using method of moments
	guess := methodOfMoments(dataObserv, respObserv, weightObserv) //Guess starts in log-space and as a result the optimize package will optimize in log-space

	// Optimze package --> Requires a function that returns some numerical value to minimize --> We are trying to maximize the log-likelihood --> Minimize the negative log-likelihood
	f := func(logParams []float64) float64 {
		//Take the guesses passed in by the optimizer out of log space
		alpha := math.Exp(logParams[0])
		beta := math.Exp(logParams[1])
		return logLikelihood(alpha, beta, dataObserv, weightObserv, respObserv)
	}

	//Optimize package --> Does the expectation maximization
	problem := optimize.Problem{Func: f}                            //Function whose return value is being minimized
	settings := &optimize.Settings{GradientThreshold: 0}            //Nelder-mead doesn't use a gradient/derivative --> Optimize function still requires it to be defined by default --> Set as 0
	result, err := optimize.Minimize(problem, guess, settings, nil) //Uses Nelder-Mead heuristic by default --> Essentially plugs in values over and over in some systematic way to find a minimum

	//Edge case: Error --> Just use the base beta distribution --> Uniform distribtuion
	if err != nil {
		return 1.0, 1.0
	}

	//Return alpha and beta parameters for beta distribution
	return math.Exp(result.X[0]), math.Exp(result.X[1]) //Optimize package works in log space --> Convert back to normal
}

// initalizeResponsibilities() --> Initalizes the 2D slice of responsibilities
// Input: Two ints, length referring to the number of a beta-values measured for a sample, and numStates referring to the number of states a CpG island can be. A 2D slice of float64s for the initialize responsibiltiies that might already exist for a sample (can be nil if none).
// Output: A 2D slice of float64s corresponding to the responsibilities for a sample, which are the probabilities a beta-value belongs to a given state methylated, unmethylated, or hemi-methylated
func initalizeResponsibiltiies(length, numStates int, initResponsibilities [][]float64) [][]float64 {
	// Create the 2d repsonsibilities array
	responsibilities := make([][]float64, length)

	//If initialResponsibilties are already passed --> Use as basis for building responsibilities
	if initResponsibilities != nil {
		//Up until the number of beta-values or length of initResponsibilities --> Copy over all the values to the responsibilities 2d-array
		for i := 0; i < length && i < len(initResponsibilities); i++ {
			responsibilities[i] = make([]float64, numStates)
			copy(responsibilities[i], initResponsibilities[i])
		}
		// If initResponsibilities is shorter than number of beta-values --> Initialize the rest
		for i := len(initResponsibilities); i < length; i++ {
			responsibilities[i] = make([]float64, numStates)
			for k := 0; k < numStates; k++ {
				responsibilities[i][k] = 1.0 / float64(numStates) //Initialize all the responsibilities as being equal
			}
		}
		//If no initialResponsibilities --> Initialize
	} else {
		//Each beta-value has 3 responsibilities --> Methylated, unmethylated, or hemimethylated
		for i := 0; i < length; i++ {
			responsibilities[i] = make([]float64, numStates)
			for k := 0; k < numStates; k++ {
				responsibilities[i][k] = 1.0 / float64(numStates) //Initialize all the responsibilities as being equal
			}
		}
	}

	return responsibilities
}

// updateResponsibilities() --> Computes the log-probabilities of each beta-value for each state and updates the responsibilities slice passed in accordingly
// Input: Two ints, length referring to the length of the sampleData, and numStates referring to the number of possible states a single beta-value can represent. Six slices of float64s. sampleData is the data for a sample.  weights is the probe-level weights for a datapoint in a sample. logProbs is currently empty. stateProportion tells you the proportion of each state over the total. alpha and beta are the parameters for a beta-distribution for each state for each sample. A 2D slice of float64s, responsibilities, containing the probabilities for every CpG island of sample being one of the possible states.
// Output: Nothing.
func updateResponsibilities(length, numStates int, sampleData, weights, logProbs, stateProportion, alpha, beta []float64, responsibilities [][]float64) {
	//For each CpG island in a sample
	for i := 0; i < length; i++ {

		//Initialize maxLog to negative infinity --> Need negative inifinity as the log of all probabilities are negative
		//maxLog is needed for what is essentially a LogSumExp function here
		maxLog := math.Inf(-1)

		//For each state of a single CpG island
		for k := 0; k < numStates; k++ {
			// Log-probability = log(stateProportion) + log(beta-dist(beta-value for a Cpg island))
			logProbs[k] = math.Log(stateProportion[k]+1e-16) + betaLogPDF(alpha[k], beta[k], sampleData[i]) //Add 1e-16 to log(stateProportion) to prevent taking log of 0 in edge cases

			//If the current log-probability is greater than the current maxLog --> Update
			if logProbs[k] > maxLog {
				maxLog = logProbs[k]
			}
		}

		// Shift log-probabilities by maxLog --> Exponentiate --> Accumulate sum for normalization
		totalResp := 0.0

		//For each state
		for k := 0; k < numStates; k++ {
			//Set responsibilities as the exponent of (logProb - maxLog) --> Subtracting maxLog prevents taking the exponent of a super small number and underflowing to 0
			responsibilities[i][k] = math.Exp(logProbs[k] - maxLog) //maxLog will always be negative as probabilities are between 0 and 1 --> Essentially, adding

			//Add up the total responsibilities --> Mass of the beta-distribution
			totalResp += responsibilities[i][k]
		}

		// Edge case: All probabilities extremely small --> Assign uniform responsibilities
		if totalResp == 0 {
			for k := 0; k < numStates; k++ {
				responsibilities[i][k] = 1.0 / float64(numStates)
			}

			// Normalize responsibilities for current sample
		} else {
			for k := 0; k < numStates; k++ {
				responsibilities[i][k] /= totalResp
			}
		}
	}
}

// fitEMBetaMixture: BMIQ-style EM that uses betaEst2 for alpha/beta estimation.
// Inputs:
// sampleData --> beta-values for a sample
// numStates --> The number of possible states a CpG island can be (methylated, unmethylated, or hemimetyhtlated)
// initResponsibilities --> Optional initial responsibilities. Can be nil
// maxIter --> Max number of iterations for updating responsibilities
// tolerance --> Min change between iterations. If lower, stop
// Outputs: An EMResult struct containing the alpha and beta for each state's beta-distirbutions for a sample, the fractions of each state for a sample, and the responsibilities for each state for every CpG island of sample
func fitEMBetaMixture(sampleData []float64, numStates int, initResponsibilities [][]float64, maxIter int, tolerance float64) (*EMResult, error) {
	//Length of SampleData = Number of beta-values from assay
	n := len(sampleData)

	// Use preprocessingClamp on sample data to remove extreme 0s and 1s --> May round to 0 or 1 due to floating point precision --> Problems later on
	for i := range sampleData {
		sampleData[i] = preprocessingClamp(sampleData[i], 1e-8) //I tried different 1e- values --> Settled on 1e-8
	}

	// Initialize weights
	//Refers to probe-level weights per CpG site --> Not relevant to 450k, but BMIQ naturally includes this in the algorithm --> Just array of all 1s
	weights := make([]float64, n)
	for i := range weights {
		weights[i] = 1.0
	}
	sumWeights := float64(n) //Sum of weights is just the number of beta-values measured as they are all 1

	// Initialize responsibilities
	responsibilities := initalizeResponsibiltiies(n, numStates, initResponsibilities)

	//Initialize parameters for the beta-distribution of each state
	alpha := make([]float64, numStates)           //alpha parameter for the beta-distirbution of a state
	beta := make([]float64, numStates)            //beta parameter for the beta-distribution of a state
	stateProportion := make([]float64, numStates) //prior probability that a random beta-value belongs to a component --> Essentially, the fraction of U, M, or H out of the total --> Updated as EM iterates
	mean := make([]float64, numStates)            // mean of the beta-distribution of a state --> = alpha/(alpha+beta) --> Need for convergence checking
	prevMean := make([]float64, numStates)        //mean from the past iteration of the EM loop --> Need to store to find the difference between iterations for convergence checking

	// small initialization for mixing (uniform)
	for k := 0; k < numStates; k++ {
		stateProportion[k] = 1.0 / float64(numStates)
	}

	//Intialize arrays --> These will be filled in later --> Declared outside of for-loop to let them be rewritten everytime --> More memory efficient
	respState := make([]float64, n)
	logProbs := make([]float64, numStates)

	// EM loop (BMIQ-like ordering)
	for iter := 0; iter < maxIter; iter++ {
		// Calculate stateProportion using current responsibilities and probe weights
		for state := 0; state < numStates; state++ {
			num := 0.0
			for i := 0; i < n; i++ {
				//Find the total mass/probability of the state --> Sum of the weighted responsibility for a state over all CpG islands in a sample
				num += weights[i] * responsibilities[i][state]
			}
			//Edge case: If the sampleData is empty or all weight are zero --> Just assign equal proportions to all states
			if sumWeights == 0 {
				stateProportion[state] = 1.0 / float64(numStates)

			} else {
				//stateProportion = Total Mass for a State / Total Weight for all CpG Islands
				stateProportion[state] = num / sumWeights
			}
			// Adjust very small stateProportions to a clamp value --> We will take a log of this later and if it is rounded to 0 due to floating point errors it will cause a problem
			if stateProportion[state] < 1e-12 { //Not going to use preprocessingClamp as that will have to be called on all values --> Slow
				stateProportion[state] = 1e-12
			}
		}

		// Store previous means to check convergence later
		copy(prevMean, mean)

		// For each state, estimate alpha, beta parameters for its coresponding beta-distribution using betaDistEstimation
		for state := 0; state < numStates; state++ {
			// Build a 1D slice to store the responsibilities for all CpG islands and a given state
			//Was declared outside --> Being rewritten every loop --> More memory efficient
			for i := 0; i < n; i++ {
				respState[i] = responsibilities[i][state]
			}

			//Estimate the beta-distirbution for a given sample and state --> Optimized here
			alphaState, betaState := betaDistEstimation(sampleData, respState, weights)

			//Update beta-distribution parameters after optimization
			alpha[state] = alphaState
			beta[state] = betaState
			mean[state] = alpha[state] / (alpha[state] + beta[state])
		}

		// Compute the log-probabilities of each beta-value for each state --> Update responsibilities
		updateResponsibilities(n, numStates, sampleData, weights, logProbs, stateProportion, alpha, beta, responsibilities)

		// Check mean convergence
		maxChange := 0.0 // maxChange is the largest change among all states
		for k := 0; k < numStates; k++ {
			change := math.Abs(mean[k] - prevMean[k])
			if change > maxChange {
				maxChange = change
			}
		}
		if maxChange < tolerance { //If largest change is below tolerance --> Stop iterating
			break
		}
	}

	// Build result
	result := &EMResult{
		Alpha:            alpha,
		Beta:             beta,
		StateProportion:  stateProportion,
		Responsibilities: responsibilities,
	}
	return result, nil
}

// dividebyProbe() --> Divides the sample data into two slices based on probe type and stores the indices for each probe type
// Input: A slice of float64s, sampleData, containing the beta-values for a given sample. A slice of ints, designType, indicating if the probe used to measure the CpG island (based on index) is type 1 or 2. A two ints, numType1 and numType2, indicating how many type 1 and type 2 probes were used
// Outputs: Two slices of ints for the indices measured by type 1 probes and type 2 probes. Two slices of float64s containing the divided sample data according to probe type used to measure a CpG
func divideByProbe(sampleData []float64, designType []int, numType1, numType2 int) ([]int, []int, []float64, []float64) {
	// Create two slices that store the indices/CpGs that use a probe 1 or probe 2
	type1indices := make([]int, 0, numType1) //Pre-allocate array size to limit memory
	type2indices := make([]int, 0, numType2)

	// Fill the slices
	for i, dt := range designType {
		if dt == 1 {
			type1indices = append(type1indices, i)
		} else if dt == 2 {
			type2indices = append(type2indices, i)
		}
	}

	// Split up the sampleData by probe type and preprocess clamp them
	type1betas := make([]float64, numType1)
	for i, idx := range type1indices {
		type1betas[i] = preprocessingClamp(sampleData[idx], 1e-8)
	}
	type2betas := make([]float64, numType2)
	for i, idx := range type2indices {
		type2betas[i] = preprocessingClamp(sampleData[idx], 1e-8)
	}

	return type1indices, type2indices, type1betas, type2betas
}

// BMIQSingle() --> Normalizes the beta-values for a single sample
// Input: A slice of float64s, sampleData, containing the beta-values for a given sample. A slice of ints, designType, indicating if the probe used to measure the CpG island (based on index) is type 1 or 2
// Output: A slice of float64s containing the normalized beta-values for a single sample, and an error indicating if the function worked
func BMIQSingle(sampleData []float64, designType []int) ([]float64, error) {

	//Check if length of sample data matches the length of the design
	if len(sampleData) != len(designType) {
		return nil, fmt.Errorf("Length of sample data does not match the probe design array.")
	}

	//Count the number of type 1 and type 2 probes --> Needed to pre-allocate an array size in memory
	numType1 := 0
	numType2 := 0
	for _, dt := range designType {
		if dt == 1 {
			numType1++
		} else if dt == 2 {
			numType2++
		}
	}

	//Edge case: If either probe is used for fewer than 10 sites --> Normalization is useless as beta-distribution estimation becomes very unwieldy
	if numType1 < 10 || numType2 < 10 {
		return nil, fmt.Errorf("Need at least 10 of each probe type.")
	}

	type1indices, type2indices, type1betas, type2betas := divideByProbe(sampleData, designType, numType1, numType2)

	numStates := 3

	// Fit the sample data to beta-distributions for each probe type and state --> Calculates the responsibilities for each CpG island to a state
	emType1, err := fitEMBetaMixture(type1betas, numStates, nil, 100, 1e-6)
	if err != nil {
		return nil, err
	}
	emType2, err := fitEMBetaMixture(type2betas, numStates, nil, 100, 1e-6)
	if err != nil {
		return nil, err
	}

	//Calculate the means of each beta distribution and the order of the means as indices --> Store in a slice
	stateMeans := make([]float64, numStates)
	stateMeanIndices := make([]int, numStates)

	for k := 0; k < numStates; k++ {
		stateMeans[k] = emType1.Alpha[k] / (emType1.Alpha[k] + emType1.Beta[k])
		stateMeanIndices[k] = k
	}

	// Sort through the slice --> Uses bubblesort
	for i := 0; i < numStates; i++ {
		for j := i + 1; j < numStates; j++ {
			if stateMeans[j] < stateMeans[i] {
				stateMeans[i], stateMeans[j] = stateMeans[j], stateMeans[i]
				stateMeanIndices[i], stateMeanIndices[j] = stateMeanIndices[j], stateMeanIndices[i]
			}
		}
	}

	// Build the orderMap --> Maps old index --> new ordered position
	orderMap := make([]int, numStates)
	for newPos := 0; newPos < numStates; newPos++ {
		oldIdx := stateMeanIndices[newPos]
		orderMap[oldIdx] = newPos
	}

	// Means of Type-I components
	//Assumes methylated region means are the fewest, then hemi-methylated, and finally methylated regions are the most frequent
	muU := stateMeans[0]
	muH := stateMeans[1]
	muM := stateMeans[2]

	//Hemimethylation fix
	// For each Type-II beta assigned mostly to H (component 1 after ordering):
	for i := range type2betas {

		// Find max-responsibility component for Type-II probe
		bestK := 0
		bestResp := emType2.Responsibilities[i][0]
		for k := 1; k < numStates; k++ {
			if emType2.Responsibilities[i][k] > bestResp {
				bestK = k
				bestResp = emType2.Responsibilities[i][k]
			}
		}

		// If the best component maps to the Hemimethylated state --> apply conformal transform
		if orderMap[bestK] == 1 { // 0=U, 1=H, 2=M after ordering
			x := type2betas[i]

			// Linear transform: stretch H so it lies between U and M
			transformed := (x-muU)*(muM-muU)/(muH-muU) + muU

			// Clamp to valid range
			if transformed < 0 {
				transformed = 0
			}
			if transformed > 1 {
				transformed = 1
			}
			type2betas[i] = transformed
		}
	}

	// Map type II beta values to type I distribution
	normalizedTypeIIBetas := make([]float64, len(type2betas))

	for i, betaVal := range type2betas {
		// Find state assignment --> Determined by what the max responsibility is for each CpG island
		bestComponent := 0
		bestResp := emType2.Responsibilities[i][0]
		for k := 1; k < numStates; k++ {
			if emType2.Responsibilities[i][k] > bestResp {
				bestComponent = k
				bestResp = emType2.Responsibilities[i][k]
			}
		}
		// Map quantile to type I component
		orderPosition := orderMap[bestComponent]
		emTypeIComp := -1
		for k := 0; k < numStates; k++ {
			if orderMap[k] == orderPosition {
				emTypeIComp = k
				break
			}
		}
		if emTypeIComp < 0 {
			return nil, fmt.Errorf("component mapping error")
		}
		p := betaCDF(emType2.Alpha[bestComponent], emType2.Beta[bestComponent], betaVal)
		q := betaQuantile(emType1.Alpha[emTypeIComp], emType1.Beta[emTypeIComp], p)
		normalizedTypeIIBetas[i] = q
	}

	// Combine results
	finalBetaValues := make([]float64, len(sampleData))
	for i, idx := range type1indices {
		finalBetaValues[idx] = type1betas[i]
	}
	for i, idx := range type2indices {
		finalBetaValues[idx] = normalizedTypeIIBetas[i]
	}
	for i := range finalBetaValues {
		if finalBetaValues[i] == 0 {
			finalBetaValues[i] = preprocessingClamp(sampleData[i], 1e-8)
		}
	}
	return finalBetaValues, nil
}

// BMIQAllSamples() --> Normalizes all samples from a dataset
// Input: A 2D slice of float64s, dataMatrix, containing the Illumina 450k dataset (rows = samples, columns = CpGs).A slice of ints, designType, indicating if the probe used to measure the CpG island (based on index) is type 1 or 2
// Ouput: A 2D slice of float64s of the dataset after normalization. An error indicating if there was a problem.
func BMIQAllSamples(dataMatrix [][]float64, designType []int) ([][]float64, error) {
	normalized := make([][]float64, len(dataMatrix))
	for i, sample := range dataMatrix {
		norm, err := BMIQSingle(sample, designType)
		if err != nil {
			return nil, fmt.Errorf("sample %d: %v", i, err)
		}
		normalized[i] = norm
	}
	return normalized, nil
}

// Made using AI
// readCSV() --> Parser. Reads in a csv file containinig our Illumina 450k datasets
// Input: A string, filename, containing the file name and/or path to the file contiaing the dataset
// Output: A 2D slice of float64s containing the read in dataset. An error indicating if there was a problem.
func readCSV(filename string) ([][]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make([][]float64, len(rows))
	for i, row := range rows {
		data[i] = make([]float64, len(row))
		for j, val := range row {
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, err
			}
			data[i][j] = f
		}
	}
	return data, nil
}

// Made using AI
// writeCSV() --> Saves the normalized data into a csv file.
// Input: A string, filename, indicating where to save and what to name the normalzied dataset
// Output: An error indicating if there was a problem
func writeCSV(filename string, data [][]float64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		strRow := make([]string, len(row))
		for i, val := range row {
			strRow[i] = strconv.FormatFloat(val, 'f', 6, 64) // 6 decimal places
		}
		if err := writer.Write(strRow); err != nil {
			return err
		}
	}
	return nil
}

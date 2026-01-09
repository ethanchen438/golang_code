# AgePredictor

Link to video demonstration: https://drive.google.com/file/d/1C4c_Tn2Gp2BKkv7iZTXqzB7QK_Hbs7_i/view?usp=sharing

Input: Any GEO database file that uses a 450k or 27k Illumina methylation probe. You will want to download the series matrix txt file for our input. 

Compiling in GO: Compile each code file by calling the appropriate directory and using go build and then go run "compiled file name" 

Step 1: Processing a matrix file 

Using the folder from FinalProjectProcessMatrices, you will need to adjust some filepaths in functions.go before running. Specifically line 20, 21, and 24. Line 20 determines the output path of where you want to dump your processed matrices. Line 21 determines the input path of the txt files you want to read in and process. Line 24 is the list of 21k probes taken from Horvath's paper as we had to downsize the amount of CpGIslands we were analyzing or else the final matrix when normalizing would become too large. 

Note: If you want to run your own data that is not taken from a GEO database file, you will need to ensure you have a row header with the name "!series_matrix_table_begin" and "!series_matrix_table_end" as this contains all the relevant CpGIsland beta values. Line 162 and 166 will be the relevant ones to change if you intend to change the required headers that include the CpGIsland beta values matrix. 

Note: You will also need a row header with "Age" in it so we have a chronological age to compare against the predicted biological age.  

Note: Any age that has months in the cell will converted into years by dividing by 12. None of the ages will be rounded however. If an age was labeled as "newborn" or "birth", it will be replaced with a 0 value. 

processMatrices Output: A transposed matrix with the Sample IDs as the rows, one column with the chronological age, and the rest of the columns will be the 21k probes that are relevant to our model. 

Note: Any SampleIDs that do not include a CpGIsland beta value for a specific probe will be replaced with a value of 0 since the matrix needs all values filled out in order to ensure that it is still square. 

Step 2: Normalizing the data

Required Packages: "gonum.org/v1/gonum/optimize" and "gonum.org/v1/gonum/stat/distuv"

In order to install packages in GO, you need to use the command "go get" and the package name as provided.
                   
Using the folder from FinalProjectNormalization, you will need to update the file path to read in the matrix, specifically line 15 in main.go. The matrix you use should be the one that was an output from FinalProjectProcessMatrices.

Step 3: Running elastic net regression on normalized data 

Using the folder from FinalProjectElasticNet 2, you will need to update the file paths to read in the matrices, specifically line 17,18, and 19 in main.go. The code will begin assembling a singular matrix filled with all the input files. The matrix you use should be the one that was an output from FinalProjectNormalization.

Note: We are using a combination from a set of values for our parameters. We chose 0.1, 0.5 1.0 for our alpha values, and 0.01, 0.1, 1.0 for our lambda values. You can edit the code to test a range of values instead but will take longer to test every combination. Lines 36 and 37 to be edited for these parameters in main.go

The output will contain a model trained on the inputted data sets and the final optimized parameters in the form of a csv file of weights.

Step 4: Predicting biological age from the model

Using the folder from FinalProjectAgePredictor, you will need to update the file path to read in the matrices, specifically lines 17, 19, and 21 in main.go. You will need a matrix of CpGs that you want to predict the biological age of (line 17), normalization paramters (output from FinalProjectElasticNet 2, line 19), and the model (output from FinalProjectElasticNet 2, line 21).

The output will include the chronological age and the predicted biological age(s) for the inputted dataset based on the model.

Step 5: Validating the model (optional)

Using the folder from FinalProjectModelTesting, you will need to update the file path to read in the vectors, specifically lines 15 and 21 in main.go. At line 15, use a csv file of the predicted age. At line 21, use a csv file of the true age.

This function will give you the Pearson correlation, median absloute error, and the mean error. This information will tell you how well the model performed.

Step 6: Visualizations (optional)

We have provided two R files in our github that can visualize parts of our model. 

The first file, CpG_coefficient_graph.R, graphs each relevant CpGIsland in our model and the strength/weight of them. 

The second file, predicted_vs_actual_graph.R, graphs the predicted biological age vs the actual chronological age for each dataset in the form of a line graph. 


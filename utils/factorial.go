package utils

import "math/big"

// CalcFactorial calculates the factorial of a non-negative integer n
// using the big.Int type to handle large numbers.
func CalcFactorial(n int64) *big.Int {
	if n < 0 {
		return big.NewInt(0) // Returns 0 for negative inputs as factorial is undefined
	}

	result := big.NewInt(1) // Initializes the result as 1, the factorial of 0
	for i := int64(1); i <= n; i++ {
		// Multiplies the result by i for each iteration
		result.Mul(result, big.NewInt(i))
	}

	return result
}

package common

import (
	"math/rand"
	"time"
)

// GenerateBinarySequences calls cb for every binary sequence of length n (values 0 or 1).
// If cb returns false generation stops early.
func GenerateBinarySequences(n int, cb func([]int) bool) {
	if n <= 0 {
		return
	}
	// We'll treat sequences as numbers from 0..(1<<n)-1 and output bits LSB..MSB as index 0..n-1
	limit := 1 << uint(n)
	seq := make([]int, n)
	for x := 0; x < limit; x++ {
		// fill sequence
		for i := 0; i < n; i++ {
			if (x>>uint(i))&1 == 1 {
				seq[i] = 1
			} else {
				seq[i] = 0
			}
		}
		if !cb(seq) {
			return
		}
	}
}

// GenerateBinarySequencesWithMaxOnes calls cb for every binary sequence of length n
// whose number of ones does not exceed m. If cb returns false generation stops early.
func GenerateBinarySequencesWithMaxOnes(n, m int, cb func([]int) bool) {
	if n <= 0 || m < 0 {
		return
	}
	if m > n {
		m = n
	}

	limit := 1 << uint(n)
	seq := make([]int, n)
	for x := 0; x < limit; x++ {
		ones := 0
		for i := 0; i < n; i++ {
			if (x>>uint(i))&1 == 1 {
				seq[i] = 1
				ones++
			} else {
				seq[i] = 0
			}
		}
		if ones > m {
			continue
		}
		if !cb(seq) {
			return
		}
	}
}

// GetAllBinarySequences returns a slice containing all binary sequences of length n.
func GetAllBinarySequences(n int) [][]int {
	result := make([][]int, 0)
	GenerateBinarySequences(n, func(s []int) bool {
		copySeq := make([]int, len(s))
		copy(copySeq, s)
		result = append(result, copySeq)
		return true
	})
	return result
}

// GetAllBinarySequencesWithMaxOnes returns all binary sequences of length n
// whose number of ones does not exceed m.
func GetAllBinarySequencesWithMaxOnes(n, m int) [][]int {
	result := make([][]int, 0)
	GenerateBinarySequencesWithMaxOnes(n, m, func(s []int) bool {
		copySeq := make([]int, len(s))
		copy(copySeq, s)
		result = append(result, copySeq)
		return true
	})
	return result
}

func GetRandomUniqueNumbers() []int {
	rand.Seed(time.Now().UnixNano())
	numbers := rand.Perm(256)
	return numbers[:32]
}

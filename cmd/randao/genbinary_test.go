package main

import (
	"reflect"
	"testing"
)

func TestGetAllBinarySequencesWithMaxOnes(t *testing.T) {
	got := GetAllBinarySequencesWithMaxOnes(3, 1)
	want := [][]int{
		{0, 0, 0},
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAllBinarySequencesWithMaxOnes(3, 1) = %v, want %v", got, want)
	}
}

func TestGetAllBinarySequencesWithMaxOnesClampM(t *testing.T) {
	got := GetAllBinarySequencesWithMaxOnes(3, 5)
	want := GetAllBinarySequences(3)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAllBinarySequencesWithMaxOnes(3, 5) = %v, want %v", got, want)
	}
}

func TestGetAllBinarySequencesWithMaxOnesInvalidInput(t *testing.T) {
	if got := GetAllBinarySequencesWithMaxOnes(3, -1); len(got) != 0 {
		t.Fatalf("expected empty result for negative m, got %v", got)
	}
	if got := GetAllBinarySequencesWithMaxOnes(0, 0); len(got) != 0 {
		t.Fatalf("expected empty result for n <= 0, got %v", got)
	}
}

func TestGenerateBinarySequencesWithMaxOnesStopEarly(t *testing.T) {
	count := 0
	GenerateBinarySequencesWithMaxOnes(4, 2, func(s []int) bool {
		count++
		return count < 3
	})

	if count != 3 {
		t.Fatalf("expected early stop after 3 callbacks, got %d", count)
	}
}

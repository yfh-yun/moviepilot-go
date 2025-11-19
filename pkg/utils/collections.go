package utils

import (
	"math/rand"
	"sort"
	"strings"
)

// Contains checks if a slice contains a specific element
func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// ContainsAny checks if a slice contains any of the specified elements
func ContainsAny[T comparable](slice []T, elements []T) bool {
	for _, element := range elements {
		if Contains(slice, element) {
			return true
		}
	}
	return false
}

// ContainsAll checks if a slice contains all of the specified elements
func ContainsAll[T comparable](slice []T, elements []T) bool {
	for _, element := range elements {
		if !Contains(slice, element) {
			return false
		}
	}
	return true
}

// IndexOf returns the index of the first occurrence of an element in a slice
func IndexOf[T comparable](slice []T, element T) int {
	for i, item := range slice {
		if item == element {
			return i
		}
	}
	return -1
}

// LastIndexOf returns the index of the last occurrence of an element in a slice
func LastIndexOf[T comparable](slice []T, element T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == element {
			return i
		}
	}
	return -1
}

// Remove removes an element from a slice
func Remove[T comparable](slice []T, element T) []T {
	result := []T{}
	for _, item := range slice {
		if item != element {
			result = append(result, item)
		}
	}
	return result
}

// RemoveAt removes an element at a specific index from a slice
func RemoveAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// RemoveDuplicates removes duplicate elements from a slice
func RemoveDuplicates[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := []T{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// Filter filters a slice based on a predicate function
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := []T{}
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map applies a function to each element of a slice
func Map[T any, U any](slice []T, mapper func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// Reduce reduces a slice to a single value using a reducer function
func Reduce[T any, U any](slice []T, initial U, reducer func(U, T) U) U {
	result := initial
	for _, item := range slice {
		result = reducer(result, item)
	}
	return result
}

// Find finds the first element that satisfies a predicate
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// FindAll finds all elements that satisfy a predicate
func FindAll[T any](slice []T, predicate func(T) bool) []T {
	return Filter(slice, predicate)
}

// Sort sorts a slice using a comparator function
func Sort[T any](slice []T, comparator func(T, T) bool) {
	sort.Slice(slice, func(i, j int) bool {
		return comparator(slice[i], slice[j])
	})
}

// Reverse reverses a slice
func Reverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, j := 0, len(slice)-1; i < len(slice); i, j = i+1, j-1 {
		result[i] = slice[j]
	}
	return result
}

// Chunk splits a slice into chunks of specified size
func Chunk[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// Flatten flattens a 2D slice into a 1D slice
func Flatten[T any](slices [][]T) []T {
	var result []T
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

// GroupBy groups elements by a key function
func GroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	groups := make(map[K][]T)

	for _, item := range slice {
		key := keyFunc(item)
		groups[key] = append(groups[key], item)
	}

	return groups
}

// Distinct returns distinct elements from a slice
func Distinct[T comparable](slice []T) []T {
	return RemoveDuplicates(slice)
}

// Intersection returns the intersection of two slices
func Intersection[T comparable](slice1, slice2 []T) []T {
	result := []T{}
	seen := make(map[T]bool)

	for _, item := range slice2 {
		seen[item] = true
	}

	for _, item := range slice1 {
		if seen[item] {
			result = append(result, item)
			delete(seen, item)
		}
	}

	return result
}

// Union returns the union of two slices
func Union[T comparable](slice1, slice2 []T) []T {
	result := append([]T{}, slice1...)
	result = append(result, slice2...)
	return RemoveDuplicates(result)
}

// Difference returns the difference between two slices
func Difference[T comparable](slice1, slice2 []T) []T {
	result := []T{}
	seen := make(map[T]bool)

	for _, item := range slice2 {
		seen[item] = true
	}

	for _, item := range slice1 {
		if !seen[item] {
			result = append(result, item)
		}
	}

	return result
}

// Shuffle shuffles a slice randomly
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	for i := len(result) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Take returns the first n elements of a slice
func Take[T any](slice []T, n int) []T {
	if n > len(slice) {
		n = len(slice)
	}
	return slice[:n]
}

// Skip returns a slice with the first n elements skipped
func Skip[T any](slice []T, n int) []T {
	if n > len(slice) {
		return []T{}
	}
	return slice[n:]
}

// First returns the first element of a slice
func First[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[0], true
}

// Last returns the last element of a slice
func Last[T any](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[len(slice)-1], true
}

// Count counts the number of elements that satisfy a predicate
func Count[T any](slice []T, predicate func(T) bool) int {
	count := 0
	for _, item := range slice {
		if predicate(item) {
			count++
		}
	}
	return count
}

// Any returns true if any element satisfies a predicate
func Any[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All returns true if all elements satisfy a predicate
func All[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Empty returns true if a slice is empty
func Empty[T any](slice []T) bool {
	return len(slice) == 0
}

// NotEmpty returns true if a slice is not empty
func NotEmpty[T any](slice []T) bool {
	return len(slice) > 0
}

// IsSliceEqual checks if two slices are equal
func IsSliceEqual[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true
}

// ConvertToStringSlice converts a slice of any type to string slice
func ConvertToStringSlice[T any](slice []T, converter func(T) string) []string {
	return Map(slice, converter)
}

// StringSliceToInterface converts a string slice to interface slice
func StringSliceToInterface(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, item := range slice {
		result[i] = item
	}
	return result
}

// InterfaceSliceToString converts an interface slice to string slice
func InterfaceSliceToString(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = item.(string)
	}
	return result
}

// JoinStringSlice joins a string slice with a separator
func JoinStringSlice(slice []string, separator string) string {
	return strings.Join(slice, separator)
}

// SplitStringToSlice splits a string by separator to create a slice
func SplitStringToSlice(str, separator string) []string {
	if str == "" {
		return []string{}
	}
	return strings.Split(str, separator)
}

// Max returns the maximum element in a slice
func Max[T comparable](slice []T, comparator func(T, T) bool) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	max := slice[0]
	for _, item := range slice[1:] {
		if comparator(item, max) {
			max = item
		}
	}

	return max, true
}

// Min returns the minimum element in a slice
func Min[T comparable](slice []T, comparator func(T, T) bool) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	min := slice[0]
	for _, item := range slice[1:] {
		if comparator(min, item) {
			min = item
		}
	}

	return min, true
}

// SumIntSlice returns the sum of an integer slice
func SumIntSlice(slice []int) int {
	sum := 0
	for _, item := range slice {
		sum += item
	}
	return sum
}

// AverageIntSlice returns the average of an integer slice
func AverageIntSlice(slice []int) float64 {
	if len(slice) == 0 {
		return 0
	}
	return float64(SumIntSlice(slice)) / float64(len(slice))
}

// SumFloatSlice returns the sum of a float slice
func SumFloatSlice(slice []float64) float64 {
	sum := 0.0
	for _, item := range slice {
		sum += item
	}
	return sum
}

// AverageFloatSlice returns the average of a float slice
func AverageFloatSlice(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	return SumFloatSlice(slice) / float64(len(slice))
}

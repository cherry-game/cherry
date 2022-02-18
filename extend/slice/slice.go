// Package cherrySlice code from: https://github.com/beego/beego/blob/develop/core/utils/slice.go
package cherrySlice

import (
	"math/rand"
	"time"
)

func Int32In(v int32, sl []int32) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// StringIn checks given string in string slice or not.
func StringIn(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func StringIndex(v string, sl []string) (int, bool) {
	for i, vv := range sl {
		if vv == v {
			return i, true
		}
	}
	return 0, false
}

// InInterface checks given interface in interface slice.
func InInterface(v interface{}, sl []interface{}) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// RandList generate an int slice from min to max.
func RandList(min, max int) []int {
	if max < min {
		min, max = max, min
	}
	length := max - min + 1
	t0 := time.Now()
	rand.Seed(int64(t0.Nanosecond()))
	list := rand.Perm(length)
	for index := range list {
		list[index] += min
	}
	return list
}

// Merge merges interface slices to one slice.
func Merge(slice1, slice2 []interface{}) (c []interface{}) {
	c = append(slice1, slice2...)
	return
}

// Reduce generates a new slice after parsing every value by reduce function
func Reduce(slice []interface{}, a func(interface{}) interface{}) (destSlice []interface{}) {
	for _, v := range slice {
		destSlice = append(destSlice, a(v))
	}
	return
}

// SliceRand returns random one from slice.
func Rand(a []interface{}) (b interface{}) {
	randNum := rand.Intn(len(a))
	b = a[randNum]
	return
}

// Sum sums all values in int64 slice.
func Sum(intslice []int64) (sum int64) {
	for _, v := range intslice {
		sum += v
	}
	return
}

// Filter generates a new slice after filter function.
func Filter(slice []interface{}, a func(interface{}) bool) (filterSlice []interface{}) {
	for _, v := range slice {
		if a(v) {
			filterSlice = append(filterSlice, v)
		}
	}
	return
}

// Diff returns diff slice of slice1 - slice2.
func Diff(slice1, slice2 []interface{}) (diffSlice []interface{}) {
	for _, v := range slice1 {
		if !InInterface(v, slice2) {
			diffSlice = append(diffSlice, v)
		}
	}
	return
}

// Intersect returns slice that are present in all the slice1 and slice2.
func Intersect(slice1, slice2 []interface{}) (diffSlice []interface{}) {
	for _, v := range slice1 {
		if InInterface(v, slice2) {
			diffSlice = append(diffSlice, v)
		}
	}
	return
}

// Chunk separates one slice to some sized slice.
func Chunk(slice []interface{}, size int) (chunkSlice [][]interface{}) {
	if size >= len(slice) {
		chunkSlice = append(chunkSlice, slice)
		return
	}
	end := size
	for i := 0; i <= (len(slice) - size); i += size {
		chunkSlice = append(chunkSlice, slice[i:end])
		end += size
	}
	return
}

// Range generates a new slice from begin to end with step duration of int64 number.
func Range(start, end, step int64) (intSlice []int64) {
	for i := start; i <= end; i += step {
		intSlice = append(intSlice, i)
	}
	return
}

// Pad prepends size number of val into slice.
func Pad(slice []interface{}, size int, val interface{}) []interface{} {
	if size <= len(slice) {
		return slice
	}
	for i := 0; i < (size - len(slice)); i++ {
		slice = append(slice, val)
	}
	return slice
}

// Unique cleans repeated values in slice.
func Unique(slice []interface{}) (uniqueSlice []interface{}) {
	for _, v := range slice {
		if !InInterface(v, uniqueSlice) {
			uniqueSlice = append(uniqueSlice, v)
		}
	}
	return
}

// Shuffle shuffles a slice.
func Shuffle(slice []interface{}) []interface{} {
	for i := 0; i < len(slice); i++ {
		a := rand.Intn(len(slice))
		b := rand.Intn(len(slice))
		slice[a], slice[b] = slice[b], slice[a]
	}
	return slice
}

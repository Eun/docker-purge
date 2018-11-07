package jq

// #cgo CFLAGS: -I/usr/local/include
// #cgo LDFLAGS: -ljq -lonig -lm
// #include "jq.c"
import "C"
import (
	"fmt"
	"unsafe"
)

// IsValidFilter checks if an jq filter is valid
func IsValidFilter(filter string) bool {
	f := C.CString(filter)
	result := C.IsValidFilter(f)
	C.free(unsafe.Pointer(f))
	return result == 1
}

// MatchesFilter returns if some json data matches a jq filter
func MatchesFilter(jsonData, filter string) (bool, error) {
	in := C.CString(jsonData)
	f := C.CString(filter)

	result := C.MatchesFilter(in, f)
	C.free(unsafe.Pointer(in))
	C.free(unsafe.Pointer(f))

	if result <= -1 {
		return false, fmt.Errorf("error %d", result)
	}

	return result == 1, nil
}

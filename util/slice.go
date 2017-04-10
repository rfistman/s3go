package util // sorry

import (
	"errors"
	"io"
	"reflect"
	"unsafe"
)

func SliceCast(ptr unsafe.Pointer, n int) []byte {
	hdr := reflect.SliceHeader{Data: uintptr(ptr), Len: n, Cap: n}
	return *(*[]byte)(unsafe.Pointer(&hdr))
}

func ReadShorts(reader io.Reader, p []int16) (int, error) {
	nb, err := reader.Read(SliceCast(unsafe.Pointer(&p[0]), 2*len(p)))
	// TODO: short reads in case of EOF?
	if err != nil {
		return 0, err
	}

	if nb%2 != 0 {
		return 0, errors.New("length not divisible by 2")
	}

	return nb / 2, nil
}

func ReadFloats(reader io.Reader, p []float32) (int, error) {
	nb, err := reader.Read(SliceCast(unsafe.Pointer(&p[0]), 4*len(p)))
	// TODO: short reads in case of EOF?
	if err != nil {
		return 0, err
	}

	if nb%4 != 0 {
		return 0, errors.New("length not divisible by 4")
	}

	return nb / 4, nil
}

package porcupine

// #cgo LDFLAGS: -lpv_porcupine
// #include <stdlib.h>
// #include "pv_porcupine.h"
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

var (
	// ErrOutOfMemory is returned when the call to porcupine results in PV_STATUS_OUT_OF_MEMORY
	ErrOutOfMemory = errors.New("porcupine: out of memory")

	// ErrIOError is returned when the call to porcupine results in PV_STATUS_IO_ERROR
	ErrIOError = errors.New("porcupine: IO error")

	// ErrInvalidArgument is returned when the call to porcupine results in PV_STATUS_INVALID_ARGUMENT
	ErrInvalidArgument = errors.New("porcupine: invalid argument")

	// ErrUnknownStatus is returned if the porcupine status code is not one of the above
	ErrUnknownStatus = errors.New("unknown status code")
)

// SampleRate returns the sample rate supported by porcupine
func SampleRate() int {
	tmp := C.pv_sample_rate()
	return int(tmp)
}

// FrameLength returns the frame length supported by porcupine
func FrameLength() int {
	tmp := C.pv_porcupine_frame_length()
	return int(tmp)
}

// SingleKeywordHandle represents an initialized porcupine instance able to handle a single keyword
type SingleKeywordHandle struct {
	once sync.Once
	h    *C.struct_pv_porcupine_object
}

// NewSingleKeywordHandle creates a porcupine instance for working with single keywords
func NewSingleKeywordHandle(modelFilePath, keywordFilePath string, sensitivity float64) (*SingleKeywordHandle, error) {
	var h *C.struct_pv_porcupine_object
	mf := C.CString(modelFilePath)
	kf := C.CString(keywordFilePath)

	defer func() {
		C.free(unsafe.Pointer(mf))
		C.free(unsafe.Pointer(kf))
	}()

	status := C.pv_porcupine_init(mf, kf, C.float(sensitivity), &h)
	if err := checkStatus(status); err != nil {
		return nil, err
	}

	return &SingleKeywordHandle{
		h: h,
	}, nil
}

// Process checks the provided audio frame and returns true if the word was detected
func (s *SingleKeywordHandle) Process(data []int16) (bool, error) {
	var result C.bool
	cData := C.malloc(C.size_t(len(data)) * C.size_t(unsafe.Sizeof(int16(0))))
	defer C.free(cData)

	// TODO is this efficient?
	tmp := (*[1<<30 - 1]int16)(cData)
	for i, v := range data {
		tmp[i] = v
	}

	status := C.pv_porcupine_process(s.h, (*C.int16_t)(unsafe.Pointer(cData)), &result)
	return bool(result), checkStatus(status)
}

// Close deletes the handle to porcupine
func (s *SingleKeywordHandle) Close() {
	s.once.Do(func() {
		C.pv_porcupine_delete(s.h)
		s.h = nil
	})
}

func checkStatus(status C.pv_status_t) error {
	switch status {
	case C.PV_STATUS_SUCCESS:
		return nil
	case C.PV_STATUS_OUT_OF_MEMORY:
		return ErrOutOfMemory
	case C.PV_STATUS_INVALID_ARGUMENT:
		return ErrInvalidArgument
	case C.PV_STATUS_IO_ERROR:
		return ErrIOError
	default:
		return ErrUnknownStatus
	}
}

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

// Porcupine interface
type Porcupine interface {
	Process(frame []int16) (string, error)
	Close()
}

// New creates a single or multiple keyword processing Porcupine instance
func New(modelPath string, keywords ...*Keyword) (Porcupine, error) {
	if len(keywords) == 0 {
		return nil, errors.New("expected at least one keyword")
	}

	if len(keywords) == 1 {
		return NewSingleKeywordHandle(modelPath, keywords[0])
	}

	return NewMultipleKeywordHandle(modelPath, keywords...)
}

// handle holds an initialized Porcupine object
type handle struct {
	once sync.Once
	h    *C.struct_pv_porcupine_object
}

// Close releases the Porcupine object
func (h *handle) Close() {
	h.once.Do(func() {
		C.pv_porcupine_delete(h.h)
		h.h = nil
	})
}

// Keyword represents an encoded keyword and its sensitivity
type Keyword struct {
	Value       string
	FilePath    string
	Sensitivity float32
}

// SingleKeywordHandle represents an initialized Porcupine instance able to handle a single keyword
type SingleKeywordHandle struct {
	*handle
	kw *Keyword
}

// NewSingleKeywordHandle creates a Porcupine instance for working with a single keyword
func NewSingleKeywordHandle(modelFilePath string, keyword *Keyword) (*SingleKeywordHandle, error) {
	var h *C.struct_pv_porcupine_object
	mf := C.CString(modelFilePath)
	kf := C.CString(keyword.FilePath)

	defer func() {
		C.free(unsafe.Pointer(mf))
		C.free(unsafe.Pointer(kf))
	}()

	status := C.pv_porcupine_init(mf, kf, C.float(keyword.Sensitivity), &h)
	if err := checkStatus(status); err != nil {
		return nil, err
	}

	return &SingleKeywordHandle{
		handle: &handle{h: h},
		kw:     keyword,
	}, nil
}

// Process checks the provided audio frame and returns the word if it was detected
func (s *SingleKeywordHandle) Process(data []int16) (string, error) {
	var result C.bool
	status := C.pv_porcupine_process(s.handle.h, (*C.int16_t)(unsafe.Pointer(&data[0])), &result)
	if err := checkStatus(status); err != nil || bool(result) == false {
		return "", err
	}

	return s.kw.Value, nil
}

// MultipleKeywordHandle represents an initialized Porcupine instance able to handle multiple keywords
type MultipleKeywordHandle struct {
	*handle
	kw []*Keyword
}

// NewMultipleKeywordHandle creates a Porcupine instance for working with multiple keywords
func NewMultipleKeywordHandle(modelFilePath string, keywords ...*Keyword) (*MultipleKeywordHandle, error) {
	mf := C.CString(modelFilePath)
	numKeywords := C.int(len(keywords))

	// create C arrays for keywords files and sensitivities
	cKeywords := C.malloc(C.size_t(len(keywords)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	tmpGoKeywords := (*[1<<30 - 1]*C.char)(cKeywords)

	tmpGoSensitivities := make([]float32, len(keywords))

	for i, k := range keywords {
		tmpGoKeywords[i] = C.CString(k.FilePath)
		tmpGoSensitivities[i] = k.Sensitivity
	}

	defer func() {
		for i := range keywords {
			C.free(unsafe.Pointer(tmpGoKeywords[i]))
		}
		C.free(cKeywords)
		C.free(unsafe.Pointer(mf))
	}()

	var h *C.struct_pv_porcupine_object
	status := C.pv_porcupine_multiple_keywords_init(mf, numKeywords, (**C.char)(unsafe.Pointer(cKeywords)), (*C.float)(unsafe.Pointer(&tmpGoSensitivities[0])), &h)
	if err := checkStatus(status); err != nil {
		return nil, err
	}

	return &MultipleKeywordHandle{
		handle: &handle{h: h},
		kw:     keywords,
	}, nil
}

// Process checks the provided audio frame and returns the index of the detected keyword
// If no keyword is detected, returns -1
func (s *MultipleKeywordHandle) Process(data []int16) (string, error) {
	var kwIndex C.int
	status := C.pv_porcupine_multiple_keywords_process(s.handle.h, (*C.int16_t)(unsafe.Pointer(&data[0])), &kwIndex)
	idx := int(kwIndex)
	if err := checkStatus(status); err != nil || idx < 0 || idx >= len(s.kw) {
		return "", err
	}

	return s.kw[idx].Value, nil
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

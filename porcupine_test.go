package porcupine

import (
	"testing"
)

func TestFrameLength(t *testing.T) {
	fl := FrameLength()

	if fl <= 0 {
		t.Errorf("expected frame length to be greater than zero. actual=%d", fl)
		t.Fail()
	}
}

func TestSampleRate(t *testing.T) {
	sr := SampleRate()

	if sr <= 0 {
		t.Errorf("expected sample rate to be greater than zero. actual=%d", sr)
		t.Fail()
	}
}

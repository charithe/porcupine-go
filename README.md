Go Wrapper For Picovoice Porcupine
===================================

[![GoDoc](https://godoc.org/github.com/charithe/porcupine-go?status.svg)](https://godoc.org/github.com/charithe/porcupine-go)

A Go wrapper for the [Picovoice Porcupine](https://github.com/Picovoice/Porcupine) wake-word detection engine.

```
go get github.com/charithe/porcupine-go
```

Requires the `pv_porcupine` library to be available in the `LD_LIBRARY_PATH` and `pv_porcupine.h` to be in the include path.
For example, on 64-bit Linux:

```shell
cp Porcupine/lib/linux/x86_64/* /usr/local/lib64/
cp Porcupine/include/* /usr/local/include
```

Demo
----

The demo application reads a 16-bit PCM S16LE stream from stdin and prints out the detected words to the console.
Accepted command-line argument are as follows:

```
  -input string
        Path to read input audio from (PCM 16-bit LE) (default "-")
  -keyword value
        Colon separated keyword, data file and sensitivity values (Eg. pineapple:pineapple_linux.ppn:0.5)
  -model_path string
        Path to the Porcupine model
```

### Single Keyword

The following invocation detects the word "pineapple" from the default audio input source.

```shell
gst-launch-1.0 -v alsasrc ! audioconvert ! audioresample ! audio/x-raw,channels=1,rate=16000,format=S16LE ! filesink location=/dev/stdout | \
    go run cmd/demo/main.go \
    -model_path=resources/porcupine_params.pv \
    -keyword=pineapple:resources/pineapple_linux.ppn:0.5
```

### Multiple Keywords

The following invocation detects "pineapple", "blueberry" and "grapefruit" from the default audio input source.

```shell
gst-launch-1.0 -v alsasrc ! audioconvert ! audioresample ! audio/x-raw,channels=1,rate=16000,format=S16LE ! filesink location=/dev/stdout | \
    go run cmd/demo/main.go \
    -model_path=resources/porcupine_params.pv \
    -keyword=pineapple:resources/pineapple_linux.ppn:0.5 \
    -keyword=blueberry:resources/blueberry_linux.ppn:0.5 \
    -keyword=grapefruit:resources/grapefruit_linux.ppn:0.5
```

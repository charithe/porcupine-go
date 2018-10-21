Go Wrapper For PicoVoice Porcupine
===================================

A work-in-progress Go wrapper for the [PicoVoice Porcupine](https://github.com/Picovoice/Porcupine) wake-word detection engine.

Requires the `pv_porcupine` library to be available in the `LD_LIBRARY_PATH` and `pv_porcupine.h` to be in the include path.
For example, on 64-bit Linux:

```shell
cp Porcupine/lib/linux/x86_64/* /usr/local/lib64/
cp Porcupine/include/* /usr/local/include

```

Demo
----

(Not working yet)

```shell
gst-launch-1.0 -v pulsesrc ! audioconvert ! audioresample ! audio/x-raw,channels=1,rate=16000,format=S16LE ! filesink location=/dev/stdout | go run cmd/demo/main.go -model=porcupine_params.pv -keyword=alexa_linux.ppn
```

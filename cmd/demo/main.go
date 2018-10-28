package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"

	porcupine "github.com/charithe/porcupine-go"
)

var (
	audioSrc    = flag.String("input", "-", "Path to input data")
	modelFile   = flag.String("model", "", "Path to the model file")
	keywordFile = flag.String("keyword", "", "Path to the keyword file")
	sensitivity = flag.Float64("sensitivity", 0.5, "Sensitivity")
)

func main() {
	flag.Parse()
	p, err := porcupine.NewSingleKeywordHandle(*modelFile, *keywordFile, *sensitivity)
	if err != nil {
		log.Fatalf("failed to initialize porcupine: %+v", err)
	}
	defer p.Close()

	var input io.Reader
	if *audioSrc == "-" {
		input = bufio.NewReader(os.Stdin)
	} else {
		f, err := os.Open(*audioSrc)
		if err != nil {
			log.Fatalf("failed to open input [%s]: %+v", *audioSrc, err)
		}
		defer f.Close()

		input = bufio.NewReader(f)
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt)

	listen(p, input, shutdownChan)
}

func listen(p *porcupine.SingleKeywordHandle, input io.Reader, shutdownChan <-chan os.Signal) {
	frameSize := porcupine.FrameLength()
	audioFrame := make([]int16, frameSize)
	buffer := make([]byte, frameSize*2)

	log.Printf("listening...")

	for {
		select {
		case <-shutdownChan:
			log.Printf("shutting down")
			return
		default:
			if err := readAudioFrame(input, buffer, audioFrame); err != nil {
				log.Printf("error: %+v", err)
				return
			}

			found, err := p.Process(audioFrame)
			if err != nil {
				log.Printf("error: %+v", err)
				continue
			}

			if found {
				log.Printf("keyword detected")
				return
			}
		}
	}
}

func readAudioFrame(src io.Reader, buffer []byte, audioFrame []int16) error {
	_, err := io.ReadFull(src, buffer)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(buffer)
	for i := 0; i < len(audioFrame); i++ {
		if err := binary.Read(buf, binary.LittleEndian, &audioFrame[i]); err != nil {
			return err
		}
	}

	return nil
}

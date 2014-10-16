package main

import (
	"flag"

	"image/png"
	"log"
	"os"

	"github.com/jeffallen/v4l"
)

var dev = flag.String("device", "/dev/video1", "filename of the video device")

func main() {
	flag.Parse()

	d, err := v4l.Open(*dev)
	if err != nil {
		log.Fatalf("Could not open device %v: %v", *dev, err)
	}

	ff := v4l.FrameFormat{
		Format: v4l.V4L2_PIX_FMT_UYVY,
		Width:  1280,
		Height: 720,
	}

	ch, err := d.Stream(ff)
	if err != nil {
		log.Fatalf("Stream setup error: %v", err)
	}

	img := <-ch
	if img == nil {
		log.Fatal("No image.")
	}

	fn := "snapshot.png"
	out, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("Could not write %v: %v", fn, err)
	}
	err = png.Encode(out, img)
	if err != nil {
		log.Fatalf("PNG encode: %v")
	}

	out.Close()

	d.Close()
}

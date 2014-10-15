package v4l

import "testing"

func TestDevice(t *testing.T) {
	dev := "/dev/video0"
	d, err := Open(dev)
	if err != nil {
		t.Fatalf("Could not open device %v: %v", dev, err)
	}

	fmt := FrameFormat{
		Format: V4L2_PIX_FMT_UYVY,
		Width:  1280,
		Height: 720,
	}

	ch, err := d.Stream(fmt)
	if err != nil {
		t.Fatalf("Stream expected no error, got %v")
	}

	img := <-ch
	if img.Bounds().Dx() != fmt.Width {
		t.Errorf("Expected width %v, got width %v", fmt.Width, img.Bounds().Dx())
	}
	if img.Bounds().Dy() != fmt.Height {
		t.Errorf("Expected height %v, got height %v", fmt.Height, img.Bounds().Dy())
	}
}

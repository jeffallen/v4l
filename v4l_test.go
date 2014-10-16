package v4l

import (
	"testing"
)

func TestFormatReq(t *testing.T) {
	req := v4l2_pix_format{
		Type:        uint32(_V4L2_BUF_TYPE_VIDEO_CAPTURE),
		Width:       uint32(1280),
		Height:      uint32(720),
		Pixelformat: uint32(V4L2_PIX_FMT_UYVY),
	}.asBytes()
	t.Logf("req: %#v", req)
}

func Disabled_TestDevice(t *testing.T) {
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

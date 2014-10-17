package v4l

import (
	"bytes"
	"image"
	"testing"
)

func TestFormatReq(t *testing.T) {
	// this was found by snooping on a v4l cli tool that works
	want := []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x5, 0x0, 0x0, 0xd0, 0x2, 0x0, 0x0, 0x55, 0x59, 0x56, 0x59}
	got := FrameFormat{
		Format: V4L2_PIX_FMT_UYVY,
		Width:  1280,
		Height: 720,
	}.req()

	got = trim(got)
	if !bytes.Equal(want, got) {
		t.Errorf("got: %#v != want", got)
	}
}

func TestRequestbufs(t *testing.T) {
	want := []byte{0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x2}
	got := trim(v4l2_requestbuffers{
		// For _V4L2_MEMORY_USERPTR, we do not neet to tell the driver how many
		// buffers we'll be using. http://lwn.net/Articles/240667/
		Count:  0,
		Type:   _V4L2_BUF_TYPE_VIDEO_CAPTURE,
		Memory: _V4L2_MEMORY_USERPTR,
	}.asBytes())

	got = trim(got)
	if !bytes.Equal(want, got) {
		t.Errorf("got: %#v != want", got)
	}
}

func TestAllocPageAligned(t *testing.T) {
	// Try lots of times because the allocator will
	// give us different outer blocks, and it is nice
	// to see that they all are resliced correctly.
	for i := 0; i < 100; i++ {
		sz := 100
		b := allocPageAligned(sz)
		if len(b) != sz {
			t.Fatal("wrong size")
		}

		if where(b)%uintptr(pageSize) != 0 {
			t.Fatal("not aligned")
		}
	}
}

func TestFrame(t *testing.T) {
	w, h := 4, 3
	frame := make([]byte, w*h*2)
	for i := range frame {
		frame[i] = byte(i)
	}

	im := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio422)
	t.Log("im: ", im)
	frameToImage(frame, im)
	t.Log("im: ", im)
	if im.Y[len(im.Y)-1] != 23 || im.Cb[len(im.Cb)-1] != 20 {
		t.Fatal("image wrong")
	}
}

func TestTrim(t *testing.T) {
	want := []byte{1, 2, 3}
	out := trim([]byte{1, 2, 3, 0, 0, 0})
	if !bytes.Equal(want, out) {
		t.Errorf("want %v != out %v", out, want)
	}
}

// trim removes trailing zeros from a byte slice; makes comparision strings smaller
// and more readable.
func trim(in []byte) []byte {
	i := len(in) - 1
	for ; i >= 0; i-- {
		if in[i] != 0 {
			break
		}
	}
	return in[0 : i+1]
}

// Package v4l gives access to V4L (Video For Linux).
// It accessess v4l directly via open(2) and ioctl(2).
// It does not use cgo wrappings of the C v4l library.
package v4l

import (
	"fmt"
	"image"
	"log"
	"os"
	"reflect"
	"sync"
	"syscall"
	"unsafe"
)

// A Format is one of the pixel formats specified by V4L.
// Only pixel formats supported by this package are here.
type Format uint32

const (
	V4L2_PIX_FMT_UYVY Format = 0x59565955 // 'UYVY', little endian
)

type imgInfo struct {
	bitsPerPixel int
	subsample    image.YCbCrSubsampleRatio
}

var infoMap = map[Format]imgInfo{
	V4L2_PIX_FMT_UYVY: {16, image.YCbCrSubsampleRatio422},
}

// A Device holds the state of a connection to a video device.
// Each Device can have at most one stream running.
type Device struct {
	path string
	f    *os.File
	wg   sync.WaitGroup
	ch   chan image.Image
}

type FrameFormat struct {
	Format        Format
	Width, Height int
	rect          image.Rectangle
}

func Open(path string) (dev *Device, err error) {
	dev = &Device{path: path}
	dev.f, err = os.Open(dev.path)
	return
}

// Close closes the underlying file handle to the V4L
// device. It stops any streams in progress, and waits
// for any goroutines to exit.
func (dev *Device) Close() {
	dev.f.Close()
	dev.f = nil

	// This is just to be sure that if the goroutine
	// does not exit, the user becomes aware (because
	// Close() hangs).
	dev.wg.Wait()
}

// Stream configures the device according to the provided FrameFormat.
// Stream returns a channel of Images. The channel is buffered
// so that if the consumer does not consume new images, new ones are
// lost. FrameFormat is validated, and may result in Stream
// returning an error if the frame format is not supported.
//
// It is an error to call Stream on a Device more than once.
//
// Stream starts a goroutine to collect frames from the Device.
// The goroutine exits when Close is called on the Device.
func (dev *Device) Stream(ff FrameFormat) (chan image.Image, error) {
	if dev.ch != nil {
		return nil, fmt.Errorf("A stream is already running on this device.")
	}
	if dev.f == nil {
		return nil, fmt.Errorf("Device is not open.")
	}

	ff.rect = image.Rect(0, 0, ff.Width, ff.Height)
	imgInfo, ok := infoMap[ff.Format]
	if !ok {
		return nil, fmt.Errorf("Frame format not supported.")
	}
	imageSize := imgInfo.bitsPerPixel / 8 * ff.Width * ff.Height

	// Setup V4L driver: format and kern<->user transfer method
	err := dev.setFormat(ff)
	if err != nil {
		return nil, err
	}
	err = dev.setUserptr()
	if err != nil {
		return nil, err
	}

	dev.ch = make(chan image.Image, 1)
	dev.wg.Add(1)
	go func() {
		frame := make([]byte, imageSize)
		for {
			req := v4l2_buffer{
				Type:    _V4L2_BUF_TYPE_VIDEO_CAPTURE,
				Memory:  _V4L2_MEMORY_USERPTR,
				Userptr: uint32(where(frame)),
				Length:  uint32(len(frame)),
			}.asBytes()

			// Enqueue the buffer.
			err := ioctl(dev.f.Fd(), _VIDIOC_QBUF, req)
			if err != nil {
				log.Print("qbuf error:", err)
				break
			}

			// Dequeue the same buffer, now filled: the ioctl blocks until
			// the next frame is available.
			err = ioctl(dev.f.Fd(), _VIDIOC_DQBUF, req)
			if err != nil {
				log.Print("dqbuf error:", err)
				break
			}

			im := image.NewYCbCr(ff.rect, imgInfo.subsample)
			frameToImage(frame, im)
			dev.ch <- im
		}
		close(dev.ch)
		dev.wg.Done()
	}()

	return dev.ch, nil
}

// frameToImage copies frame into image
func frameToImage(frame []byte, im *image.YCbCr) {
	// Format UVUY into Y, Cb and Cr planes
	// http://linuxtv.org/downloads/v4l-dvb-apis/V4L2-PIX-FMT-UYVY.html
	// U = Cb, V = Cr
	y, br := 0, 0
	for i := 0; i < len(frame); i += 4 {
		im.Cb[br] = frame[i+0]
		im.Y[y] = frame[i+1]
		im.Cr[br] = frame[i+2]
		im.Y[y+1] = frame[i+3]
		br += 1
		y += 2
	}
}

// setUserptr tells the kernel driver to expect us to allocate buffers
func (dev *Device) setUserptr() error {
	rb := v4l2_requestbuffers{
		Type:   _V4L2_BUF_TYPE_VIDEO_CAPTURE,
		Memory: _V4L2_MEMORY_USERPTR,
	}.asBytes()
	return ioctl(dev.f.Fd(), _VIDIOC_REQBUFS, rb)
}

// setFormat applies the FrameFormat to the Device.
func (dev *Device) setFormat(ff FrameFormat) error {
	return ioctl(dev.f.Fd(), _VIDIOC_S_FMT, ff.req())
}

// req formats a FrameFormat into a v4l2_pix_format, and then into a []byte,
// ready to be used by ioctl.
func (ff FrameFormat) req() []byte {
	return v4l2_pix_format{
		Type:        uint32(_V4L2_BUF_TYPE_VIDEO_CAPTURE),
		Width:       uint32(ff.Width),
		Height:      uint32(ff.Height),
		Pixelformat: uint32(ff.Format),
	}.asBytes()
}

// where returns the pointer to the data of the slice
func where(in []byte) uintptr {
	return (*reflect.SliceHeader)(unsafe.Pointer(&in)).Data
}

func ioctl(fd uintptr, req uintptr, arg []byte) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, req, where(arg))
	if e != 0 {
		return os.NewSyscallError("ioctl", e)
	}
	return nil
}

// Utilities for allocating page-aligned buffers
var pageSize = os.Getpagesize()

// allocPageAligned returns a []byte where underlying buffer is
// page aligned.
func allocPageAligned(size int) []byte {
	// Make a slice with underlying storage 1 page bigger than is requested.
	outer := make([]byte, size+pageSize)

	// find out how far we need to move forward in the underlying buffer
	// in order to be page aligned
	toNextPage := pageSize - int(where(outer)%uintptr(pageSize))

	// reslice the outer slice, resulting in an inner one
	// which is page aligned
	inner := outer[toNextPage : toNextPage+size]

	return inner
}

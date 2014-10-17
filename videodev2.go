package v4l

import (
	"bytes"
	"encoding/binary"
)

// Everything in this file is picked up from videodev2.h
// and hand translated to Go. Nothing in here should be public.

const _V4L2_BUF_TYPE_VIDEO_CAPTURE uint32 = 1
const _V4L2_MEMORY_USERPTR = 2

const _VIDIOC_S_FMT uintptr = 0xc0cc5605
const _VIDIOC_REQBUFS uintptr = 0xc0cc5608
const _VIDIOC_QBUF uintptr = 0xc0cc560f
const _VIDIOC_DQBUF uintptr = 0xc0cc5611

type v4l2_pix_format struct {
	Type, Width, Height, Pixelformat, Field          uint32
	Bytesperline, Sizeimage, Colorspace, Priv, Flags uint32
}

func (it v4l2_pix_format) asBytes() []byte { return asBytes(it) }

func asBytes(it interface{}) []byte {
	var buf = &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, it)
	return buf.Bytes()
}

type v4l2_requestbuffers struct {
	Count  uint32
	Type   uint32
	Memory uint32
}

func (it v4l2_requestbuffers) asBytes() []byte { return asBytes(it) }

type v4l2_buffer struct {
	Index, Type, Bytesused, Flags, Field uint32
	// struct timeval
	TvSec, TvUsec uint32
	// struct v4l2_timecode
	TcType, TcFlags                                                             uint32
	TcFrames, TcSeconds, TcMinutes, TcHours, TcUser0, TcUser1, TcUser2, TcUser3 uint8
	Sequence, Memory, Userptr, Length, Input, Reserved                          uint32
}

func (it v4l2_buffer) asBytes() []byte { return asBytes(it) }

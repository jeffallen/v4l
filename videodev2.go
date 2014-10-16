package v4l

// Everything in this file is picked up from videodev2.h
// and hand translated to Go. Nothing in here should be public.

const _V4L2_BUF_TYPE_VIDEO_CAPTURE uint32 = 1

const _VIDIOC_S_FMT uintptr = 0x5605 // ('V'<<8|5)

type v4l2_pix_format struct {
	Type, Width, Height, Pixelformat, Field          uint32
	Bytesperline, Sizeimage, Colorspace, Priv, Flags uint32
}

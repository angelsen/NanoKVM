package hid

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// Multi-touch protocol B — single-finger tap handler.
//
// The /dev/hidg3 gadget is created at boot by S03usbhid when /boot/usb.touch
// is present. Its descriptor contains a Touch Screen application collection
// with a Contact Identifier usage, which binds the host kernel to
// hid-multitouch.c (the real touchscreen driver) rather than hid-input.c
// (generic boot-class HID). That's the canonical path used by every
// physical touchscreen and is what Android's InputReader needs to classify
// the device as a touch screen with INPUT_PROP_DIRECT.
//
// Report format (6 bytes, no Report ID):
//
//	byte 0: tip_switch (bit 0) | padding (bits 1..7)
//	byte 1: contact identifier (always 0 — single-finger)
//	byte 2: X low byte
//	byte 3: X high byte
//	byte 4: Y low byte
//	byte 5: Y high byte
//
// Coordinates are little-endian 16-bit unsigned in the range 0..32767,
// matching the descriptor's Logical Max of 0x7FFF.
//
// Release semantics: with tip_switch=0 and no HOVERING quirk, the kernel
// treats the contact as released and emits ABS_MT_TRACKING_ID=-1 for the
// slot. Coordinates on the release report are ignored, so we zero them.

type TouchTapReq struct {
	X int `form:"x" validate:"gte=0,lte=32767"`
	Y int `form:"y" validate:"gte=0,lte=32767"`
}

// touchTapDurationMs is the hold time between the down and up reports for
// a tap. 50ms matches what physical taps produce on a human timescale and
// is enough for Android's GestureDetector to register a tap (not a
// long-press).
const touchTapDurationMs = 50

func (s *Service) TouchTap(c *gin.Context) {
	var req TouchTapReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if !s.hid.HasTouchDevice() {
		rsp.ErrRsp(c, -2, "touch device not available (set /boot/usb.touch and reboot)")
		return
	}

	down := []byte{
		0x01,             // tip_switch = 1
		0x00,             // contact id
		byte(req.X),      // X lo
		byte(req.X >> 8), // X hi
		byte(req.Y),      // Y lo
		byte(req.Y >> 8), // Y hi
	}
	up := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	s.hid.WriteHid3(down)
	time.Sleep(touchTapDurationMs * time.Millisecond)
	s.hid.WriteHid3(up)

	rsp.OkRsp(c)
	log.Debugf("hid touch_tap at (%d, %d)", req.X, req.Y)
}

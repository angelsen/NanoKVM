package hid

import (
	"errors"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Hid struct {
	g0         *os.File
	g1         *os.File
	g2         *os.File
	g3         *os.File
	kbMutex    sync.Mutex
	mouseMutex sync.Mutex
	touchMutex sync.Mutex
}

const (
	HID0 = "/dev/hidg0" // Keyboard
	HID1 = "/dev/hidg1" // Mouse (Relative Mode)
	HID2 = "/dev/hidg2" // Touchpad (Absolute Mode)
	HID3 = "/dev/hidg3" // Touchscreen (Multi-touch Protocol B, optional)
)

var (
	hid     *Hid
	hidOnce sync.Once
)

func GetHid() *Hid {
	hidOnce.Do(func() {
		hid = &Hid{}
	})
	return hid
}

func (h *Hid) Lock() {
	h.kbMutex.Lock()
	h.mouseMutex.Lock()
	h.touchMutex.Lock()
}

func (h *Hid) Unlock() {
	h.kbMutex.Unlock()
	h.mouseMutex.Unlock()
	h.touchMutex.Unlock()
}

func (h *Hid) OpenNoLock() {
	var err error
	h.CloseNoLock()

	h.g0, err = os.OpenFile(HID0, os.O_WRONLY, 0o666)
	if err != nil {
		log.Errorf("open %s failed: %s", HID0, err)
	}

	h.g1, err = os.OpenFile(HID1, os.O_WRONLY, 0o666)
	if err != nil {
		log.Errorf("open %s failed: %s", HID1, err)
	}

	h.g2, err = os.OpenFile(HID2, os.O_WRONLY, 0o666)
	if err != nil {
		log.Errorf("open %s failed: %s", HID2, err)
	}

	// hidg3 is optional: the touchscreen gadget is only created by the
	// init script when /boot/usb.touch is present. A missing device is
	// the normal case when touch is disabled, not an error.
	h.g3, err = os.OpenFile(HID3, os.O_WRONLY, 0o666)
	if err != nil {
		log.Debugf("open %s failed: %s (touch not enabled?)", HID3, err)
		h.g3 = nil
	}
}

func (h *Hid) CloseNoLock() {
	for _, file := range []*os.File{h.g0, h.g1, h.g2, h.g3} {
		if file != nil {
			_ = file.Sync()
			_ = file.Close()
		}
	}
}

func (h *Hid) Open() {
	h.kbMutex.Lock()
	defer h.kbMutex.Unlock()
	h.mouseMutex.Lock()
	defer h.mouseMutex.Unlock()
	h.touchMutex.Lock()
	defer h.touchMutex.Unlock()

	h.CloseNoLock()

	h.OpenNoLock()
}

func (h *Hid) Close() {
	h.kbMutex.Lock()
	defer h.kbMutex.Unlock()
	h.mouseMutex.Lock()
	defer h.mouseMutex.Unlock()
	h.touchMutex.Lock()
	defer h.touchMutex.Unlock()

	h.CloseNoLock()
}

func (h *Hid) WriteHid0(data []byte) {
	deadline := time.Now().Add(8 * time.Millisecond)

	h.kbMutex.Lock()
	_ = h.g0.SetWriteDeadline(deadline)
	_, err := h.g0.Write(data)
	h.kbMutex.Unlock()

	if err != nil {
		switch {
		case errors.Is(err, os.ErrClosed):
			log.Errorf("hid already closed, reopen it...")
			h.OpenNoLock()
		case errors.Is(err, os.ErrDeadlineExceeded):
			log.Debugf("write to %s timeout", HID0)
		default:
			log.Errorf("write to %s failed: %s", HID0, err)
		}
		return
	}

	log.Debugf("write to %s: %v", HID0, data)
}

func (h *Hid) WriteHid1(data []byte) {
	deadline := time.Now().Add(8 * time.Millisecond)

	h.mouseMutex.Lock()
	_ = h.g1.SetWriteDeadline(deadline)
	_, err := h.g1.Write(data)
	h.mouseMutex.Unlock()

	if err != nil {
		switch {
		case errors.Is(err, os.ErrClosed):
			log.Errorf("hid already closed, reopen it...")
			h.OpenNoLock()
		case errors.Is(err, os.ErrDeadlineExceeded):
			log.Debugf("write to %s timeout", HID1)
		default:
			log.Errorf("write to %s failed: %s", HID1, err)
		}
		return
	}

	log.Debugf("write to %s: %v", HID1, data)
}

func (h *Hid) WriteHid2(data []byte) {
	deadline := time.Now().Add(8 * time.Millisecond)

	h.mouseMutex.Lock()
	_ = h.g2.SetWriteDeadline(deadline)
	_, err := h.g2.Write(data)
	h.mouseMutex.Unlock()

	if err != nil {
		switch {
		case errors.Is(err, os.ErrClosed):
			log.Errorf("hid already closed, reopen it...")
			h.OpenNoLock()
		case errors.Is(err, os.ErrDeadlineExceeded):
			log.Debugf("write to %s timeout", HID2)
		default:
			log.Errorf("write to %s failed: %s", HID2, err)
		}
		return
	}

	log.Debugf("write to %s: %v", HID2, data)
}

// HasTouchDevice reports whether the optional touchscreen HID gadget
// (/dev/hidg3) was successfully opened. Handlers that drive touch should
// check this before calling WriteHid3.
func (h *Hid) HasTouchDevice() bool {
	return h.g3 != nil
}

func (h *Hid) WriteHid3(data []byte) {
	deadline := time.Now().Add(8 * time.Millisecond)

	h.touchMutex.Lock()
	if h.g3 == nil {
		h.touchMutex.Unlock()
		log.Debugf("%s not open, skipping write", HID3)
		return
	}
	_ = h.g3.SetWriteDeadline(deadline)
	_, err := h.g3.Write(data)
	h.touchMutex.Unlock()

	if err != nil {
		switch {
		case errors.Is(err, os.ErrClosed):
			log.Errorf("hid already closed, reopen it...")
			h.OpenNoLock()
		case errors.Is(err, os.ErrDeadlineExceeded):
			log.Debugf("write to %s timeout", HID3)
		default:
			log.Errorf("write to %s failed: %s", HID3, err)
		}
		return
	}

	log.Debugf("write to %s: %v", HID3, data)
}

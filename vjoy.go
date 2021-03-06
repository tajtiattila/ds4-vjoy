package main

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tajtiattila/hid/ds4"
	"github.com/tajtiattila/hid/ds4/ds4util"
	"github.com/tajtiattila/vjoy"
)

type VJD struct {
	mtx  sync.Mutex
	dev  *vjoy.Device
	dev2 *vjoy.Device
}

type vjoyHandler struct {
	vjd *VJD
	d   *ds4.Device
	tl  TouchLogic
	sl  SetStater
	gf  ds4util.Filter

	bw bool
}

const (
	NormalLogic    = 0
	BumpShiftLogic = 1
)

type connHandler struct {
	vjd VJD

	logic int

	swipe  bool
	slider TouchSlider

	// ngf creates a new gyro filter
	ngf func() ds4util.Filter
}

func (ch *connHandler) Connect(d *ds4.Device, e ds4util.Entry) (ds4util.StateHandler, error) {
	h := &vjoyHandler{
		vjd: &ch.vjd,
		d:   d,
		bw:  batteryWarn(e.Battery),
		gf:  ch.ngf(),
	}

	switch ch.logic {
	case NormalLogic:
		h.sl = SetStaterFunc(setState)
	case BumpShiftLogic:
		h.sl = new(bumperLogic)
	}

	h.tl = newTouchLogic(ch.swipe, ch.slider)

	h.setColor()
	return h, nil
}

func newTouchLogic(swipe bool, slider TouchSlider) TouchLogic {
	var slidef sliderFunc
	hasslider := true
	switch slider {
	case TouchLinearSlider:
		slidef = linearSlider
	case TouchThrottleSlider:
		slidef = throttleSlider
	default:
		slidef = noSlider
		hasslider = false
	}

	if swipe {
		sh := new(swipeHat)
		sh.logic = ds4util.NewSwipeLogic(sh)
		sh.slidef = slidef
		return sh
	}

	if !hasslider {
		return new(emptyLogic)
	}

	return &touchSlider{f: slidef}
}

func (h *vjoyHandler) State(s *ds4.State) error {
	bw := s.Battery&0x10 == 0 && s.Battery&0xF < 2
	if h.bw != bw {
		h.bw = bw
		h.setColor()
	}

	h.tl.HandleState(s)

	h.vjd.mtx.Lock()
	defer h.vjd.mtx.Unlock()

	vj := h.vjd.dev
	vj.Hat(1).SetDiscrete(h.tl.HatState())
	vj.Axis(vjoy.Slider0).Setf(h.tl.Slider())
	h.sl.SetState(vj, s)
	vj.Update()

	if vj2 := h.vjd.dev2; vj2 != nil {
		const m = 10000
		v := h.gf.Filter([]int{
			int(s.XGyro) * m,
			int(s.YGyro) * m,
			int(s.ZGyro) * m,
		})
		x, y, z := float64(v[0]), float64(v[1]), float64(v[2])
		r, p := ds4.GyroRollPitch(x, y, z)
		ri := dzscale(10, 50, r)
		pi := dzscale(10, 50, p)
		vj2.Axis(vjoy.AxisX).Seti(int(ri))
		vj2.Axis(vjoy.AxisY).Seti(int(pi))
		vj2.Update()
	}

	return nil
}

func (h *vjoyHandler) Close() error {
	h.vjd.mtx.Lock()
	defer h.vjd.mtx.Unlock()
	h.vjd.dev.Reset()
	h.vjd.dev.Update()
	return nil
}

func (h *vjoyHandler) setColor() {
	if h.bw {
		h.d.SetFlashColor(ds4.Color{0xff, 0x00, 0x00},
			time.Second/2, time.Second/2)
	} else {
		h.d.SetColor(ds4.Color{0xff, 0x88, 0x00})
	}
}

type SetStater interface {
	SetState(vj *vjoy.Device, s *ds4.State)
}

type SetStaterFunc func(vj *vjoy.Device, s *ds4.State)

func (f SetStaterFunc) SetState(vj *vjoy.Device, s *ds4.State) {
	f(vj, s)
}

func setState(vj *vjoy.Device, s *ds4.State) {
	button := s.Button & ^uint32(ds4.L2|ds4.R2)

	// reduce trigger sensitivity
	const triggerMinPull = 20
	if s.L2 >= triggerMinPull {
		button |= ds4.L2
	}
	if s.R2 >= triggerMinPull {
		button |= ds4.R2
	}

	for i, m := range []uint32{
		ds4.Cross,
		ds4.Circle,
		ds4.Square,
		ds4.Triangle,

		ds4.L1,
		ds4.R1,
		ds4.L2,
		ds4.R2,
		ds4.L3,
		ds4.R3,

		ds4.Options,
		ds4.Share,
		ds4.PS,
		ds4.Click,
	} {
		vj.Button(uint(i)).Set(button&m != 0)
	}

	vj.Hat(0).SetDiscrete(hat(button))

	vj.Axis(vjoy.AxisX).Setf(axis(s.LX))
	vj.Axis(vjoy.AxisY).Setf(axis(s.LY))
	vj.Axis(vjoy.AxisRX).Setf(axis(s.RX))
	vj.Axis(vjoy.AxisRY).Setf(axis(s.RY))
}

func batteryWarn(b byte) bool {
	return b&0x10 == 0 && b&0xF < 2
}

func axis(v byte) float32 {
	const deadzone = 0.07
	r0 := (float32(v)/255*2 - 1) * math.Sqrt2
	r := r0
	if r < 0 {
		r += deadzone
	} else {
		r -= deadzone
	}
	if r*r0 < 0 {
		return 0
	}
	if r < -1 {
		return -1
	}
	if 1 < r {
		return 1
	}
	return r
}

func hat(button uint32) vjoy.HatState {
	switch button & ds4.Dpad {
	case 0:
		return vjoy.HatN
	case 2:
		return vjoy.HatE
	case 4:
		return vjoy.HatS
	case 6:
		return vjoy.HatW
	}
	return vjoy.HatOff
}

type TouchLogic interface {
	HandleState(s *ds4.State)

	Slider() float32
	HatState() vjoy.HatState
}

type emptyLogic struct{}

func (*emptyLogic) HandleState(s *ds4.State) {}
func (*emptyLogic) Slider() float32          { return 0 }
func (*emptyLogic) HatState() vjoy.HatState  { return vjoy.HatOff }

type swipeHat struct {
	logic *ds4util.SwipeLogic

	swipe [4]int32

	slider float32
	slidef sliderFunc
}

func (h *swipeHat) HandleState(s *ds4.State) {
	h.logic.HandleState(s)
}

func (h *swipeHat) Slider() float32 {
	return h.slider
}

func (h *swipeHat) HatState() vjoy.HatState {
	switch {
	case atomic.LoadInt32(&h.swipe[0]) > 0:
		return vjoy.HatN
	case atomic.LoadInt32(&h.swipe[1]) > 0:
		return vjoy.HatE
	case atomic.LoadInt32(&h.swipe[2]) > 0:
		return vjoy.HatS
	case atomic.LoadInt32(&h.swipe[3]) > 0:
		return vjoy.HatW
	}
	return vjoy.HatOff
}

func (h *swipeHat) Touch(x, y int) {
	h.slider = h.slidef(x)
}

func (h *swipeHat) Click(x, y int) {
}

func (h *swipeHat) Swipe(dir int, ntouch int) {
	if ntouch == 1 {
		var n int
		switch dir {
		case ds4util.SwipeUp:
			n = int(vjoy.HatN)
		case ds4util.SwipeRight:
			n = int(vjoy.HatE)
		case ds4util.SwipeLeft:
			n = int(vjoy.HatW)
		case ds4util.SwipeDown:
			n = int(vjoy.HatS)
		default:
			return
		}

		time.AfterFunc(100*time.Millisecond, func() {
			atomic.AddInt32(&h.swipe[n], -1)
		})

		atomic.AddInt32(&h.swipe[n], 1)
	}
}

type touchSlider struct {
	v float32
	f sliderFunc
}

func (t *touchSlider) HandleState(s *ds4.State) {
	if s.Touch[0].Active() {
		t.v = t.f(int(s.Touch[0].X))
	}
}

func (t *touchSlider) Slider() float32 { return t.v }

func (*touchSlider) HatState() vjoy.HatState { return vjoy.HatOff }

type sliderFunc func(int) float32

func noSlider(int) float32 { return 0 }

func linearSlider(x int) float32 {
	const (
		right  = 1920
		border = 200
	)
	// useful range is 0..1
	// clamping is not needed, vjoy.Axis.Setf() will do just that
	return float32(x-border)/((right-2*border)/2) - 1
}

// touchThrottle is like touchSlider, but
// 1/3 of the pad is used for negative values and
// 2/3 for positive ones, with a zero gap in between.
func throttleSlider(x int) float32 {
	const (
		right  = 1920
		border = 200

		zeroleft  = 600
		zeroright = 800
	)
	switch {
	case x < zeroleft:
		return float32(x-zeroleft) / (zeroleft - border)
	case zeroright < x:
		return float32(x-zeroright) / (right - zeroright - border)
	}
	return 0
}

const (
	axisunit = 0x3fff
)

func dzscale(dz, max, v float64) int32 {
	switch {
	case v < -dz:
		v += dz
		if v < -max {
			return -axisunit
		}
	case dz < v:
		v -= dz
		if max < v {
			return axisunit
		}
	default:
		return 0
	}
	w := v * axisunit / max
	return int32(w + math.Copysign(0.5, w))
}

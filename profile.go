package main

import (
	"errors"
	"time"

	"github.com/tajtiattila/hid/ds4"
)

// NOTE: this module is not yet used

type Profile struct {
	Axis   AxisConfig
	Button ButtonConfig
	Touch  TouchConfig
	Gyro   GyroConfig
}

// it is the framework for a profile
type ButtonMask uint32

var buttonname = []struct {
	name string
	mask ButtonMask
}{
	{"Cross", ds4.Cross},
	{"Circle", ds4.Circle},
	{"Square", ds4.Square},
	{"Triangle", ds4.Triangle},
	{"L1", ds4.L1},
	{"R1", ds4.R1},
	{"L2", ds4.L2},
	{"R2", ds4.R2},
	{"L3", ds4.L3},
	{"R3", ds4.R3},
	{"Share", ds4.Share},
	{"Options", ds4.Options},
	{"PS", ds4.PS},
	{"Click", ds4.Click},
}

func ParseButtonMask(s string) (m ButtonMask, err error) {
	for s != "" {
		var btn string
		btn, s = commaSep(s)

		found := false
		for _, n := range buttonname {
			if btn == n.name {
				m |= n.mask
				found = true
				break
			}
		}
		if !found && err == nil {
			// mark error but continue parsing
			err = errors.New("Invalid button name: " + s)
		}
	}
	return m, err
}

func (b ButtonMask) String() string {
	buf := make([]byte, 0, 32*8)
	return string(b.appendFormat(buf))
}

// appendFormat appends the string representation of
// b to p. It panics if p does not have sufficient capacity.
func (b ButtonMask) appendFormat(p []byte) []byte {
	sep := false
	for _, n := range buttonname {
		if b&n.mask != 0 {
			if sep {
				p = append(p, ',')
			}
			l := len(p)
			p = p[:l+len(n.name)]
			copy(p[l:], n.name)
			sep = true
		}
	}
	return p
}

func (b ButtonMask) MarshalJSON() ([]byte, error) {
	p := append(make([]byte, 0, 32*8), '"')
	p = b.appendFormat(p)
	return append(p, '"'), nil
}

func (b *ButtonMask) UnmarshalJSON(p []byte) (err error) {
	if len(p) == 0 || p[0] != '"' || p[len(p)-1] != '"' {
		return errors.New("Button mask must be a string")
	}
	*b, err = ParseButtonMask(string(p[1 : len(p)-1]))
	return err
}

func commaSep(p string) (next, remaining string) {
	for i, r := range p {
		if r == ',' {
			return p[:i], p[i+1:]
		}
	}
	return p, ""
}

type AxisConfig struct {
	// normal mapping:
	//  DS4(LX,LY) → vjoy0(X,Y)
	//  DS4(RX,RY) → vjoy0(RX,RY)
	// shift changes this to:
	//  DS4(LX,LY) → vjoy0(Z,Y)
	//  DS4(RX,RY) → vjoy0(RX,RZ)
	Shift ButtonMask

	StickDZ  float64
	StickMax float64
}

type ButtonConfig struct {
	Shift      ButtonMask
	ShiftCombo []ButtonMask
	Shiftable  ButtonMask
}

const (
	TouchNoSlider TouchSlider = iota

	// Position mapped directly to slider -1..1
	TouchLinearSlider

	// TouchThrottleSlider makes it easy to set
	// the slider to zero. 1/3 of the touch area
	// is used for negative, and 2/3 for positive values.
	TouchThrottleSlider
)

type TouchSlider int

type TouchConfig struct {
	SwipeHat bool
	Slider   TouchSlider
}

type GyroConfig struct {
	SmoothAlpha  float64
	SmoothMovAvg time.Duration

	RollDZ  float64
	RollMax float64

	PitchDZ   float64
	PitchMax  float64
	PitchZero float64
}

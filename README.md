Sony® Playstation® Dual Shock 4 mapper for Windows

This program detects any DS4 controllers connected via bluetooth, and
feeds its control input to a [vjoystick](http://vjoystick.sourceforge.net/site/) device.

Currently only a single input device is supported.

Installation
============

The executable can be placed and run from anywhere. It might require the installation of [DS4Windows](http://ds4windows.com). DS4Windows should not be used together with this app.

[vjoystick](http://vjoystick.sourceforge.net/site/) must be installed and configured for this program to work.

To support all features, vjoystick should be configured with at least the following:

1st device
----------

* X, Y, RX, RY axes for sticks, Z and RZ for shifted sticks, and first slider for the touch slider.
* 32 buttons
* 4 discrete hats (unshifted D-pad, swipe D-pad, two for shifted D-pads)

2nd device
----------

Axes X, Y (used for gyro)


Configuration
=============

There is no GUI configuration yet.

Without any command line options, DS4 axes and buttons are mapped to vjoystick axes and buttons.

DS4 LX, LY → vjoystick #0 X, Y (eg. roll and pitch)
DS4 RX, RY → vjoystick #0 RX, RY (eg. lateral and vertical movement)

DS4 Dpad → vjoystick #0 hat #1

Command line options
--------------------

`-bumper` enables special bumper shift logic. With L1, R1 or Circle depressed different virtual buttons and hat inputs appear on the vjoystick device. This way clients (games) cannot falsely interpret multiple button presses incorrectly.

With `-bumper` the axis assignment is the same as above, but with L1 it is changed to:

DS4 LX, LY → vjoystick #0 Z, Y (eg. yaw and pitch)
DS4 RX, RY → vjoystick #0 RX, RZ (eg. lateral and longitudinal movement)

`-swipehat` enables a virtual hat which is pressed in the swipe direction for a short duration after swiping on the touch pad.

DS4 touchpad swipe → vjoystick #0 hat #2

`-slider` or `-throttle` enable the virtual slider.

`-slider` is linear with no neutral position. Touch positions are mapped directly to analog slider values between -1 and 1.

`-throttle` has a neutral position and prefers positive values. 1/3 of the usable area is converted to negative values between -1 and 0, and 2/3 is converted to positives between 0 and 1, with a gap in between where a slider is set to zero.

DS4 touch → vjoystick #0 slider #0

`-gyro` enables gyroscope input:

DS4 roll → vjoystick #1 X
DS4 pitch → vjoystick #1 Y

`-alpha` is the alpha value for 1st order smoothing filter. It should be greater than zero and less or equal to 1. Defaults to 1 (no smoothing). Small values reduce noise better at the expense of responsiveness.

`-movavg` is the moving average filter duration for gyro. Defaults to zero (no moving average filter). Larger values reduce noise at the expense of input lag.

The command line `ds4-vjoy -gyro -alpha 0.1 -movavg 200ms` seems to provide reasonable noise reduction and responsiveness.

Basic input
===========

	Left stick:  Joystick X/Y
	Right stick: Joystick RX/RY

	Cross:    Button 1
	Circle:   Button 2
	Square:   Button 3
	Triangle: Button 4

	D-Pad: Discrete Hat 1

	L1: Button 5
	R1: Button 6
	L2: Button 7
	R2: Button 8

	Share:   Button 9
	Options: Button 10

	LThumb: Button 11
	RThumb: Button 12

	PS: Button 13

Touch input
===========

Swiping on the touch pad can be converted to hat input.

	Swipe: Discrete Hat 2

Touch input can be converted to a single slider input based on the location of the touch.

	Touch or move finger: Slider 1

Gyroscope
=========

Sixaxis gyro movement can be used as joystick axes on the 2nd vjoystick device, useful for head look for example.

	Vertizal: X Axis
	Horizontal: Y Axis

Acknowledgements
================

Most of the library used by this program is based on [DS4Windows](http://ds4windows.com).


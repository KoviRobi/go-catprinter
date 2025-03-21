package catprinter

import (
	"github.com/makeworld-the-better-one/dither/v2"
	"log"
)

type PrinterOptions struct {
	feed        int
	bestQuality bool
	autoRotate  bool
	dither      bool
	fill        bool
	blackPoint  float32
	ditherAlgo  dither.ErrorDiffusionMatrix
}

// NewOptions creates a new PrinterOptions object with sane defaults.
func NewOptions() *PrinterOptions {
	return &PrinterOptions{
		feed:        40,
		bestQuality: true,
		autoRotate:  false,
		dither:      true,
		fill:        false,
		ditherAlgo:  dither.FloydSteinberg,
		blackPoint:  0.5,
	}
}

// Set paper feed after printing. Defaults to 5 lines.
func (o *PrinterOptions) SetFeed(feed int) *PrinterOptions {
	o.feed = feed
	return o
}

// SetBestQuality sets the quality option. Default is true.
// If true, prints slower with higher thermal strength, resulting in a darker image.
// Recommended for self-adhesive paper.
func (o *PrinterOptions) SetBestQuality(best bool) *PrinterOptions {
	o.bestQuality = best
	return o
}

// BestQuality returns the quality option.
func (o *PrinterOptions) BestQuality() bool {
	return o.bestQuality
}

// SetAutoRotate sets the auto rotate option. Default is false.
// If true and the image is landscape, it gets rotated to be printed in higher resolution.
func (o *PrinterOptions) SetAutoRotate(rotate bool) *PrinterOptions {
	o.autoRotate = rotate
	return o
}

// AutoRotate returns the auto rotate option.
func (o *PrinterOptions) AutoRotate() bool {
	return o.autoRotate
}

// SetDither sets the dither option. Default is true.
// If false, dithering is disabled, and the image is converted to black/white and each pixel is printed if less white than BlackPoint.
func (o *PrinterOptions) SetDither(dither bool) *PrinterOptions {
	o.dither = dither
	return o
}

// Dither returns the dither option.
func (o *PrinterOptions) Dither() bool {
	return o.dither
}

// SetDitherAlgo sets the dither algorithm. Default is FloydSteinberg.
func (o *PrinterOptions) SetDitherAlgo(algo dither.ErrorDiffusionMatrix) *PrinterOptions {
	o.ditherAlgo = algo
	return o
}

// DitherAlgo returns the dither algorithm.
func (o *PrinterOptions) DitherAlgo() dither.ErrorDiffusionMatrix {
	return o.ditherAlgo
}

func (o *PrinterOptions) SetFill(fill bool) *PrinterOptions {
	o.fill = fill
	return o
}

func (o *PrinterOptions) Fill() bool {
	return o.fill
}

// SetBlackPoint sets the black point. Default is 0.5.
// If 0.5, a gray pixel will be printed as black if it's less than 50% white. Only effective if Dither is disabled.
func (o *PrinterOptions) SetBlackPoint(bp float32) *PrinterOptions {
	if bp < 0 || bp > 1 {
		log.Panic("Invalid black point value. Must be between 0 and 1.")
	}
	o.blackPoint = bp
	return o
}

// BlackPoint returns the black point.
func (o *PrinterOptions) BlackPoint() float32 {
	return o.blackPoint
}

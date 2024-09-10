package catprinter

import (
	"github.com/disintegration/imaging"
	"github.com/makeworld-the-better-one/dither/v2"
	"image"
	"image/color"
	"image/draw"
	"log"
)

const printWidth = 384

func convertImageToBytes(img image.Image) ([]byte, error) {

	if img.Bounds().Dx() != printWidth {
		return nil, ErrInvalidImageSize
	}

	var byteArray []byte
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r != g || g != b || r != b || (r != 0 && r != 65535) {
				return nil, ErrNotBlackWhite
			}
			if r == 0 {
				byteArray = append(byteArray, 1) // black
			} else {
				byteArray = append(byteArray, 0) // white
			}
		}
	}

	return byteArray, nil

}

func grayscaleToBlackWhite(img image.Image, blackPoint float32) *image.NRGBA {

	bounds := img.Bounds()
	nrgbaImg := image.NewNRGBA(bounds)
	draw.Draw(nrgbaImg, bounds, img, bounds.Min, draw.Src)

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r != g || g != b || r != b {
				log.Panicln("logic error, image should have been grayscale")
			}
			if float32(r)/65535 < blackPoint {
				nrgbaImg.Set(x, y, color.Black)
			} else {
				nrgbaImg.Set(x, y, color.White)
			}
		}
	}

	return nrgbaImg

}

func ditherImage(img image.Image, algo dither.ErrorDiffusionMatrix) image.Image {

	palette := []color.Color{
		color.Black,
		color.White,
	}

	d := dither.NewDitherer(palette)
	d.Matrix = algo

	return d.Dither(img)

}

// FormatImage formats the image for printing by resizing it and dithering or grayscaling it
func (c *Client) FormatImage(img image.Image, opts *PrinterOptions) image.Image {

	if img.Bounds().Dx() > img.Bounds().Dy() && opts.autoRotate {
		img = imaging.Rotate90(img)
	}

	var newImg image.Image = imaging.New(img.Bounds().Dx(), img.Bounds().Dy(), color.White)
	newImg = imaging.OverlayCenter(newImg, img, 1)
	newImg = imaging.Resize(newImg, printWidth, 0, imaging.NearestNeighbor)

	if opts.dither {
		newImg = ditherImage(newImg, opts.ditherAlgo)
	} else {
		newImg = imaging.Grayscale(newImg)
		newImg = grayscaleToBlackWhite(newImg, opts.blackPoint)
	}

	if c.Debug.DumpImage {
		err := imaging.Save(newImg, "./image.png")
		if err != nil {
			log.Println("failed to save debugging image dump", err.Error())
		}
	}

	return newImg

}

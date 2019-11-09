package main

import (
	"image"
	"image/png"
	"log"
	"os"

	"github.com/UnwrittenFun/hotkey"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/gobuffalo/packr"
	"github.com/kbinani/screenshot"
	"golang.org/x/image/colornames"
)

func runScreenGrab() {
	monitor := pixelgl.PrimaryMonitor()
	x, y := monitor.Position()
	width, height := monitor.Size()

	screenSprite, screenshot, err := grabScreen(monitor)
	if err != nil {
		panic(err)
	}

	cfg := pixelgl.WindowConfig{
		Bounds:      pixel.R(x, y, width, height),
		Undecorated: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetCursorVisible(false)

	cursor, err := loadPicture("cursor.png")
	if err != nil {
		panic(err)
	}

	imd := imdraw.New(screenSprite.Picture())

	mask := pixel.RGB(0.8, 0.8, 0.8)
	var startPos *pixel.Vec

	for !win.Closed() {
		if win.JustPressed(pixelgl.KeyEscape) {
			break
		}

		win.Clear(colornames.Black)
		imd.Clear()
		imd.Reset()

		mousePos := win.MousePosition()

		if startPos == nil {
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				startPos = &mousePos
			}
		} else {
			imd.Picture = *startPos
			imd.Intensity = 1
			imd.Push(*startPos)
			imd.Picture = mousePos
			imd.Intensity = 1
			imd.Push(mousePos)
			imd.Rectangle(0)

			imd.Reset()
			imd.Color = colornames.Red
			imd.Push(*startPos, mousePos)
			imd.Rectangle(1)

			if win.JustReleased(pixelgl.MouseButtonLeft) {
				grabRect := image.Rect(int(startPos.X), int(height-startPos.Y), int(mousePos.X), int(height-mousePos.Y))
				grabImg := screenshot.SubImage(grabRect)
				file, err := os.Create("screen.png")
				if err != nil {
					panic(err)
				}
				png.Encode(file, grabImg)
				file.Close()
				break
			}
		}

		screenSprite.DrawColorMask(win, pixel.IM.Moved(win.Bounds().Center()), mask)
		imd.Draw(win)
		cursor.Draw(win, pixel.IM.Moved(mousePos))

		win.Update()
	}

	win.Destroy()
}

func run() {
	hk := hotkey.NewListener()

	_, err := hk.CreateAndRegisterHotkey(hotkey.ModCtrl+hotkey.ModAlt+hotkey.ModShift, 'P', func() {
		runScreenGrab()
	})
	if err != nil {
		log.Fatal(err)
	}

	hk.Listen()
}

var box packr.Box

func main() {
	box = packr.NewBox("./resources")

	pixelgl.Run(run)
}

func stop() {
	os.Exit(0)
}

func grabScreen(monitor *pixelgl.Monitor) (*pixel.Sprite, *image.RGBA, error) {
	x, y := monitor.Position()
	width, height := monitor.Size()
	img, err := screenshot.Capture(int(x), int(y), int(width), int(height))
	if err != nil {
		return nil, nil, err
	}

	pic := pixel.PictureDataFromImage(img)
	return pixel.NewSprite(pic, pic.Bounds()), img, nil
}

func loadPicture(path string) (*pixel.Sprite, error) {
	file, err := box.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	pic := pixel.PictureDataFromImage(img)
	return pixel.NewSprite(pic, pic.Bounds()), nil
}

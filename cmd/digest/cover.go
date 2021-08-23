package main

import (
	"aletheiaware.com/netgo"
	"fmt"
	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"path"
	"unicode"
)

type Cover struct {
	Width, Height  int
	Background     image.Image
	TextBackground uint8
	Emojis         string
	Logo           image.Image
	Title          string
	TitleFont      *truetype.Font
	TitleSize      float64
	TitleColor     color.Color
	Edition        [2]string
	EditionFont    *truetype.Font
	EditionSize    float64
	EditionColor   color.Color
	Topics         [DIGEST_LIMIT]string
	TopicFont      *truetype.Font
	TopicSize      float64
	TopicColor     color.Color
}

func (c *Cover) Image() image.Image {
	bounds := image.Rect(0, 0, c.Width, c.Height)
	rgba := image.NewRGBA(bounds)

	logoBounds := c.Logo.Bounds()
	logoLeft := (c.Width - logoBounds.Dx()) / 2.
	logoDest := image.Rect(logoLeft, 0, logoLeft+logoBounds.Dx(), logoBounds.Dy())

	titleFace := truetype.NewFace(c.TitleFont, &truetype.Options{
		Size: c.TitleSize,
		DPI:  96,
	})
	titleMetrics := titleFace.Metrics()
	titleAscent := titleMetrics.Ascent.Ceil()
	titleHeight := titleMetrics.Height.Ceil()
	titleWidth := font.MeasureString(titleFace, c.Title).Ceil()
	titleX := (c.Width - titleWidth) / 2.

	editionFace := truetype.NewFace(c.EditionFont, &truetype.Options{
		Size: c.EditionSize,
		DPI:  96,
	})
	editionMetrics := editionFace.Metrics()
	editionAscent := editionMetrics.Ascent.Ceil()
	editionHeight := editionMetrics.Height.Ceil()
	editionY := editionAscent + (logoBounds.Dy()-editionHeight)/2.
	halfWidth := c.Width / 2.
	editionYearDot := freetype.Pt((halfWidth-font.MeasureString(editionFace, c.Edition[0]).Ceil())/2., editionY)
	editionMonthDot := freetype.Pt(halfWidth+(halfWidth-font.MeasureString(editionFace, c.Edition[1]).Ceil())/2., editionY)

	topicFace := truetype.NewFace(c.TopicFont, &truetype.Options{
		Size: c.TopicSize,
		DPI:  96,
	})
	topicEmoji := func(dot fixed.Point26_6, r rune) (image.Rectangle, image.Image, bool) {
		file, err := os.Open(path.Join(c.Emojis, fmt.Sprintf("emoji_u%x.png", r)))
		if err != nil {
			log.Println(err)
			return image.Rectangle{}, nil, false
		}
		img, err := png.Decode(file)
		if err != nil {
			log.Println(err)
			return image.Rectangle{}, nil, false
		}
		s := int(c.TopicSize)
		bounds := image.Rect(0, 0, s, s)
		out := image.NewRGBA(bounds)
		xdraw.BiLinear.Scale(out, bounds, img, img.Bounds(), draw.Over, nil)
		x := dot.X.Ceil()
		y := dot.Y.Ceil()
		rect := image.Rect(x, y-s, x+s, y)
		return rect, out, true
	}
	topicMetrics := topicFace.Metrics()
	topicAscent := topicMetrics.Ascent.Ceil()
	topicHeight := topicMetrics.Height.Ceil()
	topicBoxWidth := int(float64(c.Width) / 2.5)
	leftTopicX := titleX
	rightTopicX := c.Width - titleX - topicBoxWidth

	// Background
	bgBounds := c.Background.Bounds()
	bgWidth := float64(bgBounds.Dx())
	bgHeight := float64(bgBounds.Dy())
	outWidth := float64(c.Width)
	outHeight := float64(c.Height)
	scale := math.Min(bgWidth/outWidth, bgHeight/outHeight)
	aWidth := outWidth * scale
	aHeight := outHeight * scale
	aX := (bgWidth - aWidth) / 2.
	aY := (bgHeight - aHeight) / 2.
	xdraw.BiLinear.Scale(rgba, bounds, c.Background, image.Rect(int(aX), int(aY), int(aX+aWidth), int(aY+aHeight)), draw.Over, nil)

	textBackground := image.NewUniform(color.NRGBA{A: c.TextBackground})

	// Header
	draw.Draw(rgba, image.Rect(0, 0, c.Width, titleHeight), textBackground, image.Point{}, draw.Over)

	// Logo
	draw.Draw(rgba, logoDest, c.Logo, image.Point{}, draw.Over)

	// Title
	DrawString(rgba, image.NewUniform(c.TitleColor), c.TitleFont, titleFace, nil, []rune(c.Title), freetype.Pt(titleX, titleAscent))

	// Edition
	DrawString(rgba, image.NewUniform(c.EditionColor), c.EditionFont, editionFace, nil, []rune(c.Edition[0]), editionYearDot)
	DrawString(rgba, image.NewUniform(c.EditionColor), c.EditionFont, editionFace, nil, []rune(c.Edition[1]), editionMonthDot)

	if !netgo.IsLive() {
		betaText := "BETA"
		betaDot := freetype.Pt((c.Width-font.MeasureString(editionFace, betaText).Ceil())/2., titleAscent+editionHeight)
		DrawString(rgba, image.NewUniform(&color.RGBA{0xff, 0x40, 0, 0xff}), c.EditionFont, editionFace, nil, []rune(betaText), betaDot)
	}

	// Topics
	topicMeasurer := func(s []rune) int {
		return MeasureString(c.TopicFont, topicFace, func(r rune) (fixed.Int26_6, bool) {
			_, err := os.Stat(path.Join(c.Emojis, fmt.Sprintf("emoji_u%x.png", r)))
			if err != nil {
				if os.IsNotExist(err) {
					return 0, false
				}
			}
			return fixed.I(int(c.TopicSize)), true
		}, s)
	}
	var (
		topicLines   [DIGEST_LIMIT][]*Line
		topicHeights [DIGEST_LIMIT]int
	)
	for i := 0; i < DIGEST_LIMIT; i++ {
		lines := SplitLines(c.Topics[i], topicBoxWidth, topicMeasurer)
		topicLines[i] = lines
		topicHeights[i] = len(lines) * topicHeight
	}
	var leftHeight, rightHeight int
	for i := 0; i < DIGEST_LIMIT; i += 2 {
		leftHeight += topicHeights[i]
		rightHeight += topicHeights[i+1]
	}
	leftGap := (c.Height - titleHeight - leftHeight) / 4
	rightGap := (c.Height - titleHeight - rightHeight) / 4
	leftY := titleHeight
	rightY := titleHeight
	bgX := 10
	for i := 0; i < DIGEST_LIMIT; i += 2 {
		leftY += leftGap
		rightY += rightGap

		// Left
		lines := topicLines[i]
		for j, l := range lines {
			if len(l.Text) == 0 {
				continue
			}
			x := leftTopicX
			y := leftY + (j * topicHeight)
			draw.Draw(rgba, image.Rect(x-bgX, y-topicAscent, x+l.Width+bgX, y-topicAscent+topicHeight), textBackground, image.Point{}, draw.Over)
			DrawString(rgba, image.NewUniform(c.TopicColor), c.TopicFont, topicFace, topicEmoji, []rune(l.Text), freetype.Pt(x, y))
		}
		leftY += topicHeights[i]

		// Right
		lines = topicLines[i+1]
		for j, l := range lines {
			if len(l.Text) == 0 {
				continue
			}
			x := rightTopicX + (topicBoxWidth - l.Width)
			y := rightY + (j * topicHeight)
			draw.Draw(rgba, image.Rect(x-bgX, y-topicAscent, x+l.Width+bgX, y-topicAscent+topicHeight), textBackground, image.Point{}, draw.Over)
			DrawString(rgba, image.NewUniform(c.TopicColor), c.TopicFont, topicFace, topicEmoji, []rune(l.Text), freetype.Pt(x, y))
		}
		rightY += topicHeights[i+1]
	}

	return rgba
}

func MeasureString(font *truetype.Font, face font.Face, emoji func(rune) (fixed.Int26_6, bool), runes []rune) int {
	var advance fixed.Int26_6
	previous := rune(-1)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if previous >= 0 {
			advance += face.Kern(previous, r)
		}
		index := font.Index(r)
		if index > 0 {
			a, ok := face.GlyphAdvance(r)
			if !ok {
				log.Println("Couldn't find Glyph Advance for", r)
				continue
			}
			advance += a
		} else {
			a, ok := emoji(r)
			if !ok {
				log.Println("Couldn't find Emoji for", r)
			}
			advance += a
		}
		previous = r
	}
	return advance.Ceil()
}

func DrawString(dest draw.Image, src image.Image, font *truetype.Font, face font.Face, emoji func(fixed.Point26_6, rune) (image.Rectangle, image.Image, bool), runes []rune, dot fixed.Point26_6) {
	var (
		ok        bool
		rect      image.Rectangle
		mask      image.Image
		maskPoint image.Point
		advance   fixed.Int26_6
	)
	previous := rune(-1)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if previous >= 0 {
			dot.X += face.Kern(previous, r)
		}
		index := font.Index(r)
		if index > 0 {
			rect, mask, maskPoint, advance, ok = face.Glyph(dot, r)
			if !ok {
				log.Println("Couldn't find Glyph for", r)
				continue
			}
			draw.DrawMask(dest, rect, src, image.Point{}, mask, maskPoint, draw.Over)
		} else {
			rect, src, ok := emoji(dot, r)
			if !ok {
				log.Println("Couldn't find Emoji for", r)
				continue
			}
			draw.Draw(dest, rect, src, image.Point{}, draw.Over)
			advance = fixed.I(rect.Dx())
		}

		dot.X += advance
		previous = r
	}
}

type Line struct {
	Text  string
	Width int
}

func SplitLines(text string, maxWidth int, measurer func([]rune) int) (lines []*Line) {
	textWidth := measurer([]rune(text))
	delta := maxWidth - textWidth
	if delta < 0 {
		// Split line
		wrappoint := -1
		start := 0
		end := 0
		runes := []rune(text)
		for end < len(runes) {
			c := runes[end]
			if unicode.IsSpace(c) {
				wrappoint = end
			}
			textWidth := measurer(runes[start : end+1])
			delta := maxWidth - textWidth
			if delta < 0 {
				var substring []rune
				if wrappoint == -1 {
					substring = runes[start:end]
					end++
				} else {
					substring = runes[start:wrappoint]
					end = wrappoint + 1
				}
				lines = append(lines, &Line{
					Text:  string(substring),
					Width: measurer(substring),
				})
				start = end
				wrappoint = -1
			} else {
				end++
			}
		}
		if end-start > 0 {
			substring := runes[start:end]
			lines = append(lines, &Line{
				Text:  string(substring),
				Width: measurer(substring),
			})
		}
	} else {
		lines = append(lines, &Line{
			Text:  text,
			Width: textWidth,
		})
	}
	return
}

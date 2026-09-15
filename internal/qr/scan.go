package qr

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

var ErrNoQR = errors.New("no QR code found in the image")

func ScanQR(imageBytes []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return "", err
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}

	reader := qrcode.NewQRCodeReader()
	hints := map[gozxing.DecodeHintType]any{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}

	result, err := reader.Decode(bmp, hints)
	if err != nil {
		return "", ErrNoQR
	}

	return result.GetText(), nil
}

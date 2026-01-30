package qr

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"


	"github.com/makiuchi-d/gozxing"
        "github.com/makiuchi-d/gozxing/qrcode"
)

func DetectQR(imageBytes []byte) bool {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return false
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return false
	}

	reader := qrcode.NewQRCodeReader()

	_, err := reader.Decode(bmp, nil)
	
	return err == nil
}

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

	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return "", nil
	}

	return result.GetText(), nil
}

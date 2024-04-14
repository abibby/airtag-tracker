package main

import (
	"bytes"
	"image"
	"image/png"
	"os/exec"
)

func TesseractImg(img image.Image) (string, error) {
	buff := &bytes.Buffer{}
	err := png.Encode(buff, img)
	if err != nil {
		return "", err
	}
	return tesseract(buff.Bytes())
}
func tesseract(imageBytes []byte) (string, error) {
	cmd := exec.Command("tesseract", "-", "-", "-l", "eng")
	cmd.Stdin = bytes.NewBuffer(imageBytes)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

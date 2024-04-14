package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/stoewer/go-strcase"
)

var (
	Password = os.Getenv("PASSWORD")
)

const keyPath = "/.ssh/id_rsa"

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func main() {
	err := generateKeys()
	if err != nil {
		panic(err)
	}

	s := ssh.Server{
		Addr:            ":2222",
		Handler:         handle,
		PasswordHandler: PasswordHandler,
	}
	err = s.SetOption(ssh.HostKeyFile(keyPath))
	if err != nil {
		panic(err)
	}
	err = s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func handle(s ssh.Session) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("connection failed: %v", err)
		}
	}()

	b, err := io.ReadAll(s)
	if err != nil {
		log.Printf("failed to read image: %v", err)
		return
	}

	tags, err := ProcessImage(b)
	if err != nil {
		log.Printf("failed to read image: %v", err)
		return
	}
	for _, tag := range tags {
		err = SendState(&HAState{
			EntityID: "device_tracker.airtag_" + strcase.SnakeCase(tag.Name),
			State:    tag.Location.DisplayName,
			Attributes: map[string]any{
				"friendly_name": tag.Name,

				"source_type": "gps",
				"latitude":    tag.Location.Lat,
				"longitude":   tag.Location.Lon,
			},
		})
		if err != nil {
			log.Printf("failed to save state: %v", err)
			return
		}
	}

}

func PasswordHandler(ctx ssh.Context, password string) bool {
	return password == Password
}

type Tag struct {
	Name     string
	Location *OSMLocation
}

func ProcessImage(b []byte) ([]*Tag, error) {

	err := os.WriteFile("/images/"+time.Now().Format(time.RFC3339)+".png", b, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	itemsImg := cropToItems(img)
	crb := &bytes.Buffer{}
	err = png.Encode(crb, itemsImg)
	if err != nil {
		return nil, fmt.Errorf("cropped: %v", err)
	}
	err = os.WriteFile("/images/"+time.Now().Format(time.RFC3339)+"-cropped.png", crb.Bytes(), 0o644)
	if err != nil {
		return nil, fmt.Errorf("cropped: %v", err)
	}
	txt, err := TesseractImg(itemsImg)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	log.Print("raw text\n" + txt)

	parts := strings.SplitN(txt, "\n", 2)
	name := parts[0]
	rest := strings.ReplaceAll(strings.Split(parts[1], "\n\n")[0], "\n", " ")
	matches := regexp.MustCompile(`(.*) .+ ((?i)[\d,]+ \w+ ago|now)`).FindStringSubmatch(rest)
	streetAddress := matches[1]
	location, err := LookupAddress(streetAddress)
	if err != nil {
		return nil, err
	}

	log.Printf("name: %s", name)
	log.Printf("street adderss: %s", streetAddress)
	// return nil, fmt.Errorf("%#v %#v", name, location)
	return []*Tag{
		{
			Name:     name,
			Location: location,
		},
	}, nil
}

func crop(img image.Image, x, y, width, height float64) image.Image {

	bounds := img.Bounds()
	w := float64(bounds.Dx()) / 100
	h := float64(bounds.Dy()) / 100

	cropSize := image.Rect(int(w*x), int(h*y), int(w*x+w*width), int(h*y+h*height))

	return img.(SubImager).SubImage(cropSize)

}

func cropToItems(img image.Image) image.Image {
	bounds := img.Bounds()
	if bounds.Dx() > bounds.Dy() {
		return crop(
			img,
			6, 60,
			24.5, 25,
		)
	}
	return crop(img, 9, 61.5, 34.25, 20)
}

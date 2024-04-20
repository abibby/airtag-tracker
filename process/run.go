package process

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/stoewer/go-strcase"
)

var (
	ImagePath = os.Getenv("IMAGE_PATH")
)

func init() {
	if ImagePath == "" {
		ImagePath = "/images"
	}
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func Handle(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	tags, err := ProcessImage(b)
	if err != nil {
		err = fmt.Errorf("failed to process image: %w", err)
		wErr := os.WriteFile(path.Join(ImagePath, time.Now().Format(time.RFC3339)+".png"), b, 0o644)
		if wErr != nil {
			return errors.Join(
				err,
				fmt.Errorf("failed to save failed image: %w", wErr),
			)
		}
		return err
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
			return fmt.Errorf("failed to save state: %w", err)
		}
	}
	return nil
}

type Tag struct {
	Name     string
	Location *OSMLocation
}

func ProcessImage(b []byte) ([]*Tag, error) {
	img, _, err := image.Decode(bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	itemsImg := cropToItems(img)

	dumpImage(itemsImg, "cropped")

	tagImgs := vSplitImage(itemsImg, color.RGBA{208, 213, 204, 255})

	tags := make([]*Tag, 0, len(tagImgs))

	for _, tagImg := range tagImgs {
		if tagImg.Bounds().Dy() < 50 {
			continue
		}
		// dumpImage(tagImg, fmt.Sprintf("tag-%d", i))

		partImgs := hSplitImage(tagImg, color.RGBA{250, 255, 242, 255})
		if len(partImgs) != 2 {
			log.Printf("invalid number of parts %d", len(partImgs))
		}
		nameImg := partImgs[0]
		// distanceImg := partImgs[1]
		// dumpImage(nameImg, fmt.Sprintf("tag-%d-name", i))

		txt, err := TesseractImg(nameImg)
		if err != nil {
			return nil, fmt.Errorf("failed to process image: %w", err)
		}

		log.Print("raw text\n" + txt)

		if strings.Contains(txt, "Identify Found Item") {
			continue
		}

		parts := strings.SplitN(txt, "\n", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("not enough lines")
		}
		name := parts[0]
		rest := regexp.MustCompile(`[ \t\n]+`).ReplaceAllString(parts[1], " ")
		matches := regexp.MustCompile(`(.*) .+ ((?i)[\d,]+ \w+ ago|now)`).FindStringSubmatch(rest)
		if len(matches) == 0 {
			return nil, fmt.Errorf("could not find match")
		}
		streetAddress := matches[1]
		location, err := LookupAddress(streetAddress)
		if err != nil {
			return nil, err
		}

		log.Printf("name: %s", name)
		log.Printf("street adderss: %s", streetAddress)
		tags = append(tags, &Tag{
			Name:     name,
			Location: location,
		})
	}

	return tags, nil
}

func cropPercent(img image.Image, x, y, width, height float64) image.Image {

	bounds := img.Bounds()
	w := float64(bounds.Dx()) / 100
	h := float64(bounds.Dy()) / 100

	return crop(img, int(w*x), int(h*y), int(w*x+w*width), int(h*y+h*height))

}
func crop(img image.Image, x0, y0, x1, y1 int) image.Image {
	return img.(SubImager).SubImage(image.Rect(x0, y0, x1, y1))

}

func cropToItems(img image.Image) image.Image {
	bounds := img.Bounds()
	if bounds.Dx() > bounds.Dy() {
		return cropPercent(
			img,
			6, 60,
			24.5, 25,
		)
	}
	return cropPercent(img, 9, 61.5, 34.25, 20)
}

func dumpImage(img image.Image, name string) {
	crb := &bytes.Buffer{}
	err := png.Encode(crb, img)
	if err != nil {
		log.Printf("dump image: %v", err)
		return
	}

	err = os.WriteFile(path.Join(ImagePath, fmt.Sprintf("%s-%s.png", time.Now().Format(time.RFC3339), name)), crb.Bytes(), 0o644)
	if err != nil {
		log.Printf("dump image: %v", err)
		return
	}
}

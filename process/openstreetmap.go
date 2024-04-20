package process

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type OSMAddress struct {
	HouseNumber string `json:"house_number"`
	Road        string `json:"road"`
	City        string `json:"city"`
	State       string `json:"state"`
	ISO3166     string `json:"ISO3166-2-lvl4"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}
type OSMLocation struct {
	PlaceId     int64       `json:"place_id"`
	Licence     string      `json:"licence"`
	OsmType     string      `json:"osm_type"`
	OsmId       int64       `json:"osm_id"`
	Lat         string      `json:"lat"`
	Lon         string      `json:"lon"`
	Class       string      `json:"class"`
	Type        string      `json:"type"`
	PlaceRank   int         `json:"place_rank"`
	Importance  float64     `json:"importance"`
	Addresstype string      `json:"addresstype"`
	Name        string      `json:"name"`
	DisplayName string      `json:"display_name"`
	Address     *OSMAddress `json:"address"`
	BoundingBox []string    `json:"boundingbox"`
}

func (l *OSMLocation) Latf() float64 {
	lat, err := strconv.ParseFloat(l.Lat, 64)
	if err != nil {
		panic(err)
	}
	return lat
}
func (l *OSMLocation) Lonf() float64 {
	lon, err := strconv.ParseFloat(l.Lon, 64)
	if err != nil {
		panic(err)
	}
	return lon
}

func LookupAddress(address string) (*OSMLocation, error) {
	u := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&addressdetails=1", url.QueryEscape(address))
	// fmt.Printf("Lookup: %s\n", u)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	addrs := []*OSMLocation{}
	err = json.NewDecoder(resp.Body).Decode(&addrs)
	if err != nil {
		return nil, err
	}

	if len(addrs) < 1 {
		return nil, fmt.Errorf("no locations at the address")
	}
	addr := addrs[0]

	return addr, nil
}

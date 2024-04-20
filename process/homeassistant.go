package process

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	HABase = os.Getenv("HOME_ASSISTANT_BASE")
	HAKey  = os.Getenv("HOME_ASSISTANT_KEY")
)

type HAState struct {
	Attributes  map[string]any `json:"attributes"`
	EntityID    string         `json:"entity_id"`
	LastChanged string         `json:"last_changed,omitempty"`
	LastUpdated string         `json:"last_updated,omitempty"`
	State       string         `json:"state"`
}

func SendState(state *HAState) error {
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = haDoRequest(http.MethodPost, fmt.Sprintf("/api/states/%s", url.PathEscape(state.EntityID)), bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	return nil
}

func DeviceTrackerSee(deviceID string, loc *OSMLocation) error {
	b, err := json.Marshal(&HAState{
		State: loc.DisplayName,
		Attributes: map[string]any{
			"dev_id": deviceID,
			"gps":    []float64{loc.Latf(), loc.Lonf()},
		},
	})
	if err != nil {
		return err
	}
	_, err = haDoRequest(http.MethodPost, "/api/services/device_tracker/see", bytes.NewBuffer(b))
	return err
}

func haDoRequest(method, path string, body io.Reader) (*http.Response, error) {
	href := strings.TrimSuffix(HABase, "/") + "/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequest(method, href, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+HAKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if 200 > resp.StatusCode || resp.StatusCode >= 299 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("home-assistant: %s: bad status %s: %s", href, resp.Status, b)
	}
	return resp, nil
}

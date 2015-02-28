package crawlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Atom uses github releases API
func Atom() (string, string, error) {
	res, err := http.Get("https://api.github.com/repos/atom/atom/releases")
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	var releases []struct {
		Name   string `json:"name"`
		Assets []struct {
			URL string `json:"browser_download_url"`
		}
	}

	err = json.NewDecoder(res.Body).Decode(&releases)
	if err != nil {
		return "", "", err
	}

	for _, release := range releases[0].Assets {
		if strings.HasSuffix(release.URL, ".x86_64.rpm") {
			return strings.TrimSpace(releases[0].Name), strings.TrimSpace(release.URL), nil
		}
	}
	return "", "", errors.New("unable to find release")
}

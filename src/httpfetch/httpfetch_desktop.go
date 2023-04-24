//go:build linux || darwin || windows

package httpfetch

import (
	"io/ioutil"
	"net/http"
)

func GetBytes(targetURL string) ([]byte, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

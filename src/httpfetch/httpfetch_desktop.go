//go:build linux || darwin || windows

package httpfetch

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Data []byte
	Code int
}

func PostJSON(targetURL string, jsonBytes []byte) (Response, error) {
	var result Response
	resp, err := http.Post(targetURL, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	result.Data = data
	result.Code = resp.StatusCode
	return result, nil
}

func GetBytes(targetURL string) ([]byte, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

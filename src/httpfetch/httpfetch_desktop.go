//go:build linux || darwin || windows

package httpfetch

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Response struct {
	Data []byte
	Code int
}

func PostJSON(targetURL string, jsonBytes []byte) (Response, error) {
	var err error
	for i := 0; i < 2; i++ {
		var result Response
		result, err = tryPostJSON(targetURL, jsonBytes)
		if err == nil {
			return result, nil
		}
		time.Sleep(time.Second / 2)
	}
	return Response{}, err
}

func tryPostJSON(targetURL string, jsonBytes []byte) (Response, error) {
	resp, err := http.Post(targetURL, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	var data []byte
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{Data: data, Code: resp.StatusCode}, nil
}

func GetBytes(targetURL string) ([]byte, error) {
	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

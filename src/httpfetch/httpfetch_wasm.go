//go:build wasm

package httpfetch

import (
	"errors"
	"syscall/js"
)

type Response struct {
	Data []byte
	Code int
}

func PostJSON(targetURL string, jsonBytes []byte) (Response, error) {
	res := doFetch(targetURL, map[string]any{
		"method": "POST",
		"headers": map[string]any{
			"Accept":       "application/json",
			"Content-Type": "application/json",
		},
		"body": string(jsonBytes),
	})
	return Response{Data: res.data, Code: res.status}, res.err
}

func GetBytes(targetURL string) ([]byte, error) {
	res := doFetch(targetURL, nil)
	return res.data, res.err
}

type fetchResult struct {
	data   []byte
	status int
	err    error
}

func doFetch(targetURL string, params map[string]any) fetchResult {
	ch := make(chan fetchResult)

	var fetch js.Value
	if len(params) == 0 {
		fetch = js.Global().Call("fetch", targetURL)
	} else {
		fetch = js.Global().Call("fetch", targetURL, params)
	}

	fetch.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		status := args[0].Get("status").Int()
		args[0].Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			size := args[0].Get("byteLength").Int()
			data := make([]byte, size)
			u8array := js.Global().Get("Uint8Array").New(args[0])
			numBytes := js.CopyBytesToGo(data, u8array)
			if numBytes != size {
				ch <- fetchResult{status: status, err: errors.New("incomplete bytes copy")}
			}
			ch <- fetchResult{status: status, data: data}
			return nil
		}))
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) any {
		ch <- fetchResult{err: errors.New(args[0].Get("message").String())}
		return nil
	}))

	res := <-ch
	return res
}

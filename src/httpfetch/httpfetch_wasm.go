//go:build wasm

package httpfetch

import (
	"errors"
	"syscall/js"
)

func GetBytes(targetURL string) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}

	ch := make(chan result)

	fetch := js.Global().Call("fetch", targetURL)
	fetch.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			size := args[0].Get("byteLength").Int()
			data := make([]byte, size)
			u8array := js.Global().Get("Uint8Array").New(args[0])
			numBytes := js.CopyBytesToGo(data, u8array)
			if numBytes != size {
				ch <- result{err: errors.New("incomplete bytes copy")}
			}
			ch <- result{data: data}
			return nil
		}))
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) any {
		ch <- result{err: errors.New(args[0].Get("message").String())}
		return nil
	}))

	res := <-ch
	return res.data, res.err
}

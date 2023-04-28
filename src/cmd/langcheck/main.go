package main

import (
	"fmt"
	"sort"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/assets"
)

func main() {
	var translations = []struct {
		key string
	}{
		{key: "en"},
		{key: "ru"},
	}

	var missingKeys []string
	missing := 0
	invalid := 0

	ctx := ge.NewContext(ge.ContextConfig{
		Mute:       true,
		FixedDelta: true,
	})
	ctx.Loader.OpenAssetFunc = assets.MakeOpenAssetFunc(ctx, "")
	assets.RegisterRawResources(ctx)

	eng := translations[0]
	engData := langData{
		dict: loadDict(ctx, eng.key),
	}
	for _, t := range translations[1:] {
		dict := loadDict(ctx, t.key)
		engData.dict.WalkKeys(func(k string) {
			if dict.Has(k) {
				return
			}
			missing++
			missingKeys = append(missingKeys, fmt.Sprintf("[-] %s misses %s translation", t.key, k))
		})
		dict.WalkKeys(func(k string) {
			if !engData.dict.Has(k) {
				invalid++
				missingKeys = append(missingKeys, fmt.Sprintf("[!] %s has excessive %s key", t.key, k))
			}
		})
	}

	sort.Strings(missingKeys)
	for _, k := range missingKeys {
		fmt.Println(k)
	}
	fmt.Printf("missing translations: %d\n", missing)
	fmt.Printf("invalid keys: %d\n", invalid)
}

type langData struct {
	dict *langs.Dictionary
}

func loadDict(ctx *ge.Context, key string) *langs.Dictionary {
	var id resource.RawID
	switch key {
	case "en":
		id = assets.RawDictEn
	case "ru":
		id = assets.RawDictRu
	default:
		panic(fmt.Sprintf("unsupported lang: %q", key))
	}
	dict, err := langs.ParseDictionary(key, 4, ctx.Loader.LoadRaw(id).Data)
	if err != nil {
		panic(err)
	}
	if err := dict.Load("", ctx.Loader.LoadRaw(id+1).Data); err != nil {
		panic(err)
	}
	if err := dict.Load("", ctx.Loader.LoadRaw(id+2).Data); err != nil {
		panic(err)
	}
	if err := dict.Load("", ctx.Loader.LoadRaw(id+3).Data); err != nil {
		panic(err)
	}
	return dict
}

package gamedata

type DisplayRatio struct {
	Name   string
	Width  float64
	Height float64
}

var SupportedDisplayRatios = []DisplayRatio{
	{Name: "16:9", Width: 960, Height: 540},
	{Name: "16:10", Width: 960, Height: 600},
	{Name: "18:9", Width: 1080, Height: 540},
	{Name: "19:9", Width: 1140, Height: 540},
	{Name: "20:9", Width: 1200, Height: 540},
	{Name: "21:9", Width: 1260, Height: 540},
}

func FindDisplayRatio(name string) int {
	for i, d := range SupportedDisplayRatios {
		if d.Name == name {
			return i
		}
	}
	panic("invalid screen ratio")
}

func MaxDisplayWidth() float64 {
	v := 0.0
	for _, d := range SupportedDisplayRatios {
		if d.Width > v {
			v = d.Width
		}
	}
	return v
}

package gamedata

type DisplayRatio struct {
	Name   string
	Width  float64
	Height float64
}

var SupportedDisplayRatio = []DisplayRatio{
	{Name: "16:9", Width: 960, Height: 540},
	{Name: "18:9", Width: 1080, Height: 540},
	{Name: "19:9", Width: 1140, Height: 540},
	{Name: "20:9", Width: 1200, Height: 540},
	{Name: "21:9", Width: 1260, Height: 540},
}

func MaxDisplayWidth() float64 {
	v := 0.0
	for _, d := range SupportedDisplayRatio {
		if d.Width > v {
			v = d.Width
		}
	}
	return v
}

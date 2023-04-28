package main

import (
	"math/rand"
)

func floatRange(rng *rand.Rand, min, max float64) float64 {
	return min + rng.Float64()*(max-min)
}

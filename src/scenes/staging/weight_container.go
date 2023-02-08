package staging

import (
	"github.com/quasilyte/gmath"
)

type weightContainer[T comparable] struct {
	Elems []weightContainerElem[T]
}

type weightContainerElem[T comparable] struct {
	Key    T
	Weight float64
}

func newWeightContainer[T comparable](keys ...T) *weightContainer[T] {
	elems := make([]weightContainerElem[T], len(keys))
	for i, k := range keys {
		elems[i].Key = k
	}
	return &weightContainer[T]{Elems: elems}
}

func (c *weightContainer[T]) MaxKey() T {
	maxKey := c.Elems[0].Key
	maxValue := c.Elems[0].Weight
	for _, kv := range c.Elems[1:] {
		if kv.Weight > maxValue {
			maxValue = kv.Weight
			maxKey = kv.Key
		}
	}
	return maxKey
}

func (c *weightContainer[T]) GetWeight(k T) float64 {
	for _, kv := range c.Elems {
		if kv.Key == k {
			return kv.Weight
		}
	}
	return 0
}

func (c *weightContainer[T]) SetWeight(k T, weight float64) {
	for i, kv := range c.Elems {
		if kv.Key == k {
			c.Elems[i].Weight = weight
			continue
		}
	}
}

func (c *weightContainer[T]) AddWeight(k T, weight float64) {
	total := 0.0
	for i, kv := range c.Elems {
		if kv.Key == k {
			newValue := gmath.Clamp(kv.Weight+weight, 0, 1)
			total += newValue
			c.Elems[i].Weight = newValue
			continue
		}
		total += kv.Weight
	}
	for i, kv := range c.Elems {
		if kv.Weight == 0 {
			continue
		}
		c.Elems[i].Weight = kv.Weight / total
	}
}

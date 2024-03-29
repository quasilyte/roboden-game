package gameui

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/roboden-game/gameui/eui"
)

type NavDir int

const (
	NavRight NavDir = iota
	NavDown
	NavLeft
	NavUp
)

type NavTree struct {
	blocks []*NavBlock
}

func NewNavTree() *NavTree {
	return &NavTree{
		blocks: make([]*NavBlock, 0, 4),
	}
}

func (t *NavTree) NewBlock() *NavBlock {
	b := &NavBlock{}
	t.blocks = append(t.blocks, b)
	return b
}

func (t *NavTree) GetFirstElem() *NavElem {
	for _, b := range t.blocks {
		if b.Disabled {
			continue
		}
		if e := b.GetFirstElem(); e != nil {
			return e
		}
	}
	return nil
}

func (t *NavTree) FindElem(w eui.Widget) *NavElem {
	for _, b := range t.blocks {
		for _, e := range b.elems {
			if e.Widget == w {
				return e
			}
		}
	}
	return nil
}

type NavBlock struct {
	elems []*NavElem
	paths []*NavBlock

	Edges    [4]*NavBlock
	Disabled bool
}

func NewMultiNavBlock(paths ...*NavBlock) *NavBlock {
	return &NavBlock{
		paths: paths,
	}
}

func (b *NavBlock) GetFirstElem() *NavElem {
	for _, e := range b.elems {
		if e.Widget.GetWidget().Disabled {
			continue
		}
		return e
	}
	return nil
}

func (b *NavBlock) NewElem(w widget.PreferredSizeLocateableWidget) *NavElem {
	e := &NavElem{
		Widget: w,
		block:  b,
	}
	b.elems = append(b.elems, e)
	return e
}

func (b *NavBlock) Find(dir NavDir) *NavElem {
	if b.Disabled {
		return nil
	}

	other := b.Edges[dir]
	if other == nil {
		return nil
	}

	if other.paths != nil {
		for _, p := range other.paths {
			if p.Disabled {
				continue
			}
			if e := p.GetFirstElem(); e != nil {
				return e
			}
		}
		return nil
	}

	return other.GetFirstElem()
}

type NavElem struct {
	Widget widget.PreferredSizeLocateableWidget

	Edges [4]*NavElem

	block *NavBlock
}

func (e *NavElem) Find(dir NavDir) *NavElem {
	if e.Edges[dir] != nil {
		return e.Edges[dir]
	}
	return e.block.Find(dir)
}

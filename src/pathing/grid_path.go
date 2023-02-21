package pathing

import (
	"strings"
)

const (
	gridPathBytes  = (16 - 2)
	gridPathMaxLen = gridPathBytes * 4
)

type GridPath struct {
	bytes [gridPathBytes]byte
	len   byte
	pos   byte
}

type BuildPathResult struct {
	Steps    GridPath
	Complete bool
}

func (p GridPath) String() string {
	parts := make([]string, 0, p.len)
	prevPos := p.pos // Restore the pos later
	p.Rewind()
	for p.HasNext() {
		parts = append(parts, p.Next().String())
	}
	p.pos = prevPos
	return "{" + strings.Join(parts, ",") + "}"
}

func (p *GridPath) Len() int {
	return int(p.len)
}

func (p *GridPath) HasNext() bool {
	return p.pos != 0
}

func (p *GridPath) Rewind() {
	p.pos = p.len
}

func (p *GridPath) Current() Direction {
	return p.get(p.pos - 1)
}

func (p *GridPath) Next() Direction {
	d := p.Current()
	p.pos--
	return d
}

func (p *GridPath) push(dir Direction) {
	i := p.pos
	p.pos++
	p.len++
	byteIndex := i / 4
	bitShift := (i % 4) * 2
	if byteIndex < uint8(len(p.bytes)) {
		p.bytes[byteIndex] |= byte(dir << bitShift)
	}
}

func (p *GridPath) get(i byte) Direction {
	byteIndex := i / 4
	bitShift := (i % 4) * 2
	if byteIndex < uint8(len(p.bytes)) {
		return Direction((p.bytes[byteIndex] >> bitShift) & 0b11)
	}
	return DirNone
}

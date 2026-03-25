package formatter

import "go/format"

type Gofmt struct{}

func NewGofmt() Gofmt {
	return Gofmt{}
}

func (Gofmt) Name() string {
	return "gofmt"
}

func (Gofmt) Format(src []byte) ([]byte, error) {
	return format.Source(src)
}

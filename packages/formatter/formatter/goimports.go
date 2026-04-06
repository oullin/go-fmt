package formatter

import "golang.org/x/tools/imports"

type Goimports struct{}

func NewGoimports() Goimports {
	return Goimports{}
}

func (Goimports) Name() string {
	return "goimports"
}

func (Goimports) Format(src []byte) ([]byte, error) {
	return imports.Process("", src, nil)
}

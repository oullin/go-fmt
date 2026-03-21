package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

type Goimports struct{}

func NewGoimports() Goimports {
	return Goimports{}
}

func (Goimports) Name() string {
	return "goimports"
}

func (Goimports) Format(src []byte) ([]byte, error) {
	cmd := exec.Command("goimports")
	cmd.Stdin = bytes.NewReader(src)

	var out bytes.Buffer

	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if isCommandMissing(err) {
			return src, nil
		}

		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%s", bytes.TrimSpace(stderr.Bytes()))
		}

		return nil, err
	}

	return out.Bytes(), nil
}

func isCommandMissing(err error) bool {
	var execErr *exec.Error

	if errors.As(err, &execErr) {
		return errors.Is(execErr.Err, exec.ErrNotFound)
	}

	return false
}

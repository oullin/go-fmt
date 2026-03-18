package engine

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"os/exec"

	"github.com/oullin/go-fmt/internal/config"
	"github.com/oullin/go-fmt/internal/rules"
	"github.com/oullin/go-fmt/internal/rules/spacing"
)

type Engine struct {
	cfg   config.Config
	rules []rules.Rule
}

func New(cfg config.Config) *Engine {
	engine := &Engine{cfg: cfg}

	if cfg.Rules.Spacing.Enabled {
		engine.rules = append(engine.rules, spacing.New())
	}

	return engine
}

func (e *Engine) Check(paths []string) (Report, error) {
	return e.run(paths, false)
}

func (e *Engine) Format(paths []string) (Report, error) {
	return e.run(paths, true)
}

func (e *Engine) run(paths []string, write bool) (Report, error) {
	files, err := CollectGoFiles(paths, e.cfg)

	if err != nil {
		return Report{}, err
	}

	report := Report{
		Result: "pass",
		Files:  len(files),
	}

	for _, file := range files {
		result := e.processFile(file, write)
		report.Results = append(report.Results, result)

		if result.Changed {
			report.Changed++
		}
	}

	switch {
	case report.ErrorCount() > 0:
		report.Result = "fail"
	case write && report.Changed > 0:
		report.Result = "fixed"
	case !write && report.Changed > 0:
		report.Result = "fail"
	}

	return report, nil
}

func (e *Engine) processFile(path string, write bool) FileResult {
	original, err := os.ReadFile(path)

	if err != nil {
		return FileResult{File: path, Error: fmt.Sprintf("read file: %v", err)}
	}

	current := append([]byte(nil), original...)
	result := FileResult{File: path}

	for _, rule := range e.rules {
		violations, formatted, err := rule.Apply(path, current)

		if err != nil {
			result.Error = fmt.Sprintf("%s: %v", rule.Name(), err)

			return result
		}

		result.Violations = append(result.Violations, violations...)

		if !bytes.Equal(formatted, current) {
			current = formatted
			result.Applied = append(result.Applied, rule.Name())
		}
	}

	if e.cfg.Formatters.Gofmt {
		formatted, err := format.Source(current)

		if err != nil {
			result.Error = fmt.Sprintf("gofmt: %v", err)

			return result
		}

		if !bytes.Equal(formatted, current) {
			current = formatted
			result.Applied = append(result.Applied, "gofmt")
		}
	}

	if e.cfg.Formatters.Goimports {
		formatted, changed, err := runGoimports(current)

		if err != nil {
			result.Error = fmt.Sprintf("goimports: %v", err)

			return result
		}

		if changed {
			current = formatted
			result.Applied = append(result.Applied, "goimports")
		}
	}

	if bytes.Equal(original, current) {
		return result
	}

	result.Changed = true
	result.Diff = generateDiff(string(original), string(current))

	if write {
		if err := os.WriteFile(path, current, 0o644); err != nil {
			result.Error = fmt.Sprintf("write file: %v", err)
			result.Changed = false

			return result
		}
	}

	return result
}

func runGoimports(content []byte) ([]byte, bool, error) {
	cmd := exec.Command("goimports")
	cmd.Stdin = bytes.NewReader(content)

	var out bytes.Buffer

	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if isCommandMissing(err) {
			return content, false, nil
		}

		if stderr.Len() > 0 {
			return nil, false, fmt.Errorf("%s", bytes.TrimSpace(stderr.Bytes()))
		}

		return nil, false, err
	}

	formatted := out.Bytes()

	return formatted, !bytes.Equal(formatted, content), nil
}

func isCommandMissing(err error) bool {
	var execErr *exec.Error

	if errors.As(err, &execErr) {
		return errors.Is(execErr.Err, exec.ErrNotFound)
	}

	return false
}

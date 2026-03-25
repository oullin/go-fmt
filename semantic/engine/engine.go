package engine

import (
	"bytes"
	"fmt"
	"os"

	"github.com/oullin/go-fmt/semantic/config"
	"github.com/oullin/go-fmt/semantic/formatter"
	"github.com/oullin/go-fmt/semantic/rules"
)

type Engine struct {
	cfg        config.Config
	rules      []rules.Rule
	formatters []formatter.Formatter
}

func New(cfg config.Config, rr []rules.Rule, ff []formatter.Formatter) *Engine {
	return &Engine{
		cfg:        cfg,
		rules:      rr,
		formatters: ff,
	}
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

	for _, f := range e.formatters {
		formatted, err := f.Format(current)

		if err != nil {
			result.Error = fmt.Sprintf("%s: %v", f.Name(), err)

			return result
		}

		if !bytes.Equal(formatted, current) {
			current = formatted
			result.Applied = append(result.Applied, f.Name())
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

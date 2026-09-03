package oxlintconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type generatedConfig struct {
	Extends []string                   `json:"extends"`
	Rules   map[string]json.RawMessage `json:"rules"`
}

func TestWithBatchesComposesBundledBaseWithRootOverlay(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{"rules":{"require-await":"error"}}`)
	overlay := writeConfig(t, root, ".oxlintrc.jsonc", "// project policy\n{\n\t\"rules\": {\"require-await\": \"off\"}\n}\n")
	file := writeConfig(t, root, "app.ts", "export const value = 1;\n")

	var generated string

	err := WithBatches(Request{RootDir: root, BundledConfig: base, Files: []string{file}}, func(batches []Batch) error {
		if len(batches) != 1 {
			t.Fatalf("batches = %d, want 1", len(batches))
		}

		generated = batches[0].ConfigPath

		if filepath.Dir(generated) != root {
			t.Fatalf("generated config directory = %q, want %q", filepath.Dir(generated), root)
		}

		if !reflect.DeepEqual(batches[0].Files, []string{file}) {
			t.Fatalf("files = %#v, want [%q]", batches[0].Files, file)
		}

		config := readGeneratedConfig(t, generated)

		if !reflect.DeepEqual(config.Extends, []string{base}) {
			t.Fatalf("extends = %#v, want [%q]", config.Extends, base)
		}

		if got := string(config.Rules["require-await"]); got != `"off"` {
			t.Fatalf("require-await = %s, want off", got)
		}

		if overlay == generated {
			t.Fatal("consumer config was reused instead of materialising an entry config")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithBatches: %v", err)
	}

	if _, err := os.Stat(generated); !os.IsNotExist(err) {
		t.Fatalf("generated config still exists after callback: %v", err)
	}
}

func TestWithBatchesGroupsRootAndNearestNestedOverlays(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{}`)
	rootConfig := writeConfig(t, root, ".oxlintrc.json", `{"rules":{"eqeqeq":"error"}}`)
	nestedDir := filepath.Join(root, "packages", "generated")
	nestedConfig := writeConfig(t, nestedDir, ".oxlintrc", `{"rules":{"eqeqeq":"off"}}`)
	rootFile := writeConfig(t, root, "app.ts", "export const root = 1;\n")
	nestedFile := writeConfig(t, nestedDir, "client.ts", "export const nested = 1;\n")

	err := WithBatches(Request{RootDir: root, BundledConfig: base, Files: []string{nestedFile, rootFile}}, func(batches []Batch) error {
		if len(batches) != 2 {
			t.Fatalf("batches = %d, want 2", len(batches))
		}

		got := map[string][]string{}

		for _, batch := range batches {
			config := readGeneratedConfig(t, batch.ConfigPath)
			got[batch.Files[0]] = config.Extends
		}

		want := map[string][]string{
			rootFile:   {base},
			nestedFile: {base, rootConfig},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("config chains = %#v, want %#v (nested config %q)", got, want, nestedConfig)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithBatches: %v", err)
	}
}

func TestWithBatchesRejectsMissingExplicitOverlayWithoutFiles(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{}`)
	missing := filepath.Join(root, "missing.json")
	called := false

	err := WithBatches(Request{
		RootDir:       root,
		BundledConfig: base,
		GlobalOverlay: missing,
	}, func([]Batch) error {
		called = true

		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "FMTKIT_OXLINTRC") || !strings.Contains(err.Error(), missing) {
		t.Fatalf("error = %v, want missing FMTKIT_OXLINTRC path", err)
	}

	if called {
		t.Fatal("callback ran for an invalid explicit overlay")
	}
}

func TestWithBatchesKeepsConsumerExtendsAfterFmtkitLayers(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{}`)
	projectBase := writeConfig(t, root, "project-base.json", `{"rules":{"eqeqeq":"error"}}`)
	file := writeConfig(t, root, "app.ts", "export const value = 1;\n")
	writeConfig(t, root, ".oxlintrc.json", `{"extends":["./project-base.json"],"rules":{"eqeqeq":"off"}}`)

	err := WithBatches(Request{RootDir: root, BundledConfig: base, Files: []string{file}}, func(batches []Batch) error {
		config := readGeneratedConfig(t, batches[0].ConfigPath)
		want := []string{base, "./project-base.json"}

		if !reflect.DeepEqual(config.Extends, want) {
			t.Fatalf("extends = %#v, want %#v (project base %q)", config.Extends, want, projectBase)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithBatches: %v", err)
	}
}

func TestWithBatchesDoesNotInjectAnExplicitlyExtendedRootTwice(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{}`)
	writeConfig(t, root, ".oxlintrc.json", `{"overrides":[{"files":["*.test.ts"],"rules":{"no-console":"off"}}]}`)
	nestedDir := filepath.Join(root, "package")
	writeConfig(t, nestedDir, ".oxlintrc.json", `{"extends":["../.oxlintrc.json"]}`)
	file := writeConfig(t, nestedDir, "app.ts", "export const value = 1;\n")

	err := WithBatches(Request{RootDir: root, BundledConfig: base, Files: []string{file}}, func(batches []Batch) error {
		config := readGeneratedConfig(t, batches[0].ConfigPath)
		want := []string{base, "../.oxlintrc.json"}

		if !reflect.DeepEqual(config.Extends, want) {
			t.Fatalf("extends = %#v, want %#v", config.Extends, want)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithBatches: %v", err)
	}
}

func TestWithBatchesAppliesExplicitOverlayLast(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{}`)
	rootConfig := writeConfig(t, root, ".oxlintrc.json", `{}`)
	nestedDir := filepath.Join(root, "package")
	nestedConfig := writeConfig(t, nestedDir, ".oxlintrc.json", `{}`)
	file := writeConfig(t, nestedDir, "app.ts", "export const value = 1;\n")
	globalDir := t.TempDir()
	globalBase := writeConfig(t, globalDir, "company-base.json", `{}`)
	global := writeConfig(t, globalDir, "strict.jsonc", `{"extends":["./company-base.json"]}`)

	err := WithBatches(Request{
		RootDir:       root,
		BundledConfig: base,
		GlobalOverlay: global,
		Files:         []string{file},
	}, func(batches []Batch) error {
		if filepath.Dir(batches[0].ConfigPath) != globalDir {
			t.Fatalf("generated config directory = %q, want %q", filepath.Dir(batches[0].ConfigPath), globalDir)
		}

		config := readGeneratedConfig(t, batches[0].ConfigPath)
		want := []string{base, rootConfig, nestedConfig, "./company-base.json"}

		if !reflect.DeepEqual(config.Extends, want) {
			t.Fatalf("extends = %#v, want %#v (company base %q)", config.Extends, want, globalBase)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("WithBatches: %v", err)
	}
}

func TestWithBatchesRejectsAmbiguousDirectoryConfig(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{}`)
	writeConfig(t, root, ".oxlintrc.json", `{}`)
	writeConfig(t, root, ".oxlintrc.jsonc", `{}`)
	file := writeConfig(t, root, "app.ts", "export const value = 1;\n")

	err := WithBatches(Request{RootDir: root, BundledConfig: base, Files: []string{file}}, func([]Batch) error {
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "multiple oxlint configs") {
		t.Fatalf("error = %v, want ambiguous-config failure", err)
	}
}

func TestWithBatchesRejectsTrailingCommasLikeOxlint(t *testing.T) {
	root := t.TempDir()
	base := writeConfig(t, t.TempDir(), ".oxlintrc.json", `{}`)
	writeConfig(t, root, ".oxlintrc.jsonc", `{"rules":{"eqeqeq":"error",},}`)
	file := writeConfig(t, root, "app.ts", "export const value = 1;\n")

	err := WithBatches(Request{RootDir: root, BundledConfig: base, Files: []string{file}}, func([]Batch) error {
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "trailing comma") {
		t.Fatalf("error = %v, want Oxlint-compatible trailing-comma failure", err)
	}
}

func readGeneratedConfig(t *testing.T, path string) generatedConfig {
	t.Helper()

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}

	var config generatedConfig

	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("decode generated config: %v", err)
	}

	return config
}

func writeConfig(t *testing.T, directory, name, contents string) string {
	t.Helper()

	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}

	path := filepath.Join(directory, name)

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}

	return path
}

package report

import (
	"bytes"
	"testing"

	"github.com/oullin/go-fmt/packages/correctness"
	semanticengine "github.com/oullin/go-fmt/packages/semantic/engine"
	"github.com/oullin/go-fmt/packages/semantic/rules"
)

func TestProjectReportNormalizesSemanticAndCorrectnessResults(t *testing.T) {
	projected := projectReport("/work", sampleCombinedReport())

	if projected.Result != "fail" {
		t.Fatalf("unexpected combined result: %#v", projected)
	}

	if projected.Semantic.Result != "fail" || projected.Semantic.Files != 2 || projected.Semantic.Changed != 1 || projected.Semantic.Violations != 1 {
		t.Fatalf("unexpected semantic summary: %#v", projected.Semantic)
	}

	if len(projected.Semantic.Results) != 2 {
		t.Fatalf("unexpected semantic results: %#v", projected.Semantic.Results)
	}

	changed := projected.Semantic.Results[0]

	if changed.File != "sample.go" || !changed.Changed || len(changed.Applied) != 2 || len(changed.Violations) != 1 || changed.Error != "" {
		t.Fatalf("unexpected changed result: %#v", changed)
	}

	errored := projected.Semantic.Results[1]

	if errored.File != "broken.go" || errored.Error != "parse error" {
		t.Fatalf("unexpected errored result: %#v", errored)
	}

	if len(projected.Semantic.Errors) != 2 {
		t.Fatalf("unexpected semantic errors: %#v", projected.Semantic.Errors)
	}

	if projected.Semantic.Errors[0].File != "walk.go" || projected.Semantic.Errors[1].File != "broken.go" {
		t.Fatalf("unexpected semantic error paths: %#v", projected.Semantic.Errors)
	}

	if projected.Correctness.Status != "fail" || len(projected.Correctness.Errors) != 1 || projected.Correctness.Errors[0].File != "module-a" {
		t.Fatalf("unexpected correctness projection: %#v", projected.Correctness)
	}
}

func TestRenderJSONUsesProjectedReport(t *testing.T) {
	var out bytes.Buffer

	if err := RenderJSON(&out, "/work", sampleCombinedReport()); err != nil {
		t.Fatalf("render json: %v", err)
	}

	const want = "{\"result\":\"fail\",\"semantic\":{\"result\":\"fail\",\"files\":2,\"changed\":1,\"results\":[{\"file\":\"sample.go\",\"applied\":[\"spacing\",\"gofmt\"],\"violations\":[{\"rule\":\"spacing\",\"line\":7,\"message\":\"after if statement\"}],\"changed\":true}],\"errors\":[{\"file\":\"walk.go\",\"message\":\"walk failed\"},{\"file\":\"broken.go\",\"message\":\"parse error\"}]},\"correctness\":{\"status\":\"fail\",\"errors\":[{\"file\":\"module-a\",\"message\":\"automatic go vet ./... failed:\\nvet output\"}]}}\n"

	if out.String() != want {
		t.Fatalf("unexpected json output:\n%s", out.String())
	}
}

func TestRenderAgentUsesProjectedReport(t *testing.T) {
	var out bytes.Buffer

	if err := RenderAgent(&out, "/work", sampleCombinedReport()); err != nil {
		t.Fatalf("render agent: %v", err)
	}

	const want = "{\n  \"result\": \"fail\",\n  \"semantic\": {\n    \"result\": \"fail\",\n    \"summary\": {\n      \"files\": 2,\n      \"changed\": 1,\n      \"violations\": 1\n    },\n    \"changed\": [\n      {\n        \"file\": \"sample.go\",\n        \"steps\": [\n          \"spacing\",\n          \"gofmt\"\n        ]\n      }\n    ],\n    \"violations\": [\n      {\n        \"file\": \"sample.go\",\n        \"rule\": \"spacing\",\n        \"line\": 7,\n        \"message\": \"after if statement\"\n      }\n    ],\n    \"errors\": [\n      {\n        \"file\": \"walk.go\",\n        \"message\": \"walk failed\"\n      },\n      {\n        \"file\": \"broken.go\",\n        \"message\": \"parse error\"\n      }\n    ]\n  },\n  \"correctness\": {\n    \"status\": \"fail\",\n    \"errors\": [\n      {\n        \"file\": \"module-a\",\n        \"message\": \"automatic go vet ./... failed:\\nvet output\"\n      }\n    ]\n  }\n}\n"

	if out.String() != want {
		t.Fatalf("unexpected agent output:\n%s", out.String())
	}
}

func sampleCombinedReport() Combined {
	return Combined{
		Semantic: semanticengine.Report{
			Result:  "fail",
			Files:   2,
			Changed: 1,
			Results: []semanticengine.FileResult{
				{
					File:    "/work/sample.go",
					Applied: []string{"spacing", "gofmt"},
					Violations: []rules.Violation{
						{
							Rule:    "spacing",
							Line:    7,
							Message: "after if statement",
						},
					},
					Changed: true,
				},
				{
					File:  "/work/broken.go",
					Error: "parse error",
				},
			},
			Errors: []semanticengine.ErrorResult{
				{
					File:    "/work/walk.go",
					Message: "walk failed",
				},
			},
		},
		Correctness: correctness.Report{
			Root: "/work",
			Errors: []correctness.ErrorResult{
				{
					File:    "/work/module-a",
					Message: "automatic go vet ./... failed:\nvet output",
				},
			},
		},
	}
}

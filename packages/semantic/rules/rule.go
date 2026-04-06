package rules

type Violation struct {
	Rule    string `json:"rule"`
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

type Rule interface {
	Name() string
	Apply(path string, src []byte) ([]Violation, []byte, error)
}

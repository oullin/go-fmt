package cli

const (
	CheckMode  Mode = "check"
	FormatMode Mode = "format"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

package formatter

type Formatter interface {
	Name() string
	Format(src []byte) ([]byte, error)
}

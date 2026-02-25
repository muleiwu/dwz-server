package interfaces

type ShortCode interface {
	DeCode(string) (string, error)
	EnCode(string) (string, error)
}

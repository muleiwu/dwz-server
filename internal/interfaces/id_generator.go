package interfaces

type IDGenerator interface {
	Generate() (string, error)
}

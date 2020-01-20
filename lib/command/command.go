package command

type DbgpCommand interface {
	Handle() (string, error)
}

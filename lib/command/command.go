package command

type DbgpCommand interface {
	Handle() (string, error)
	GetName() string
}

type DbgpInitCommand interface {
	Handle() DbgpInitResult
	GetName() string
	Close()
	GetKey() string
}

type DbgpInitResult interface {
	AsXML() (string, error)
	IsSuccess() bool
}

package command

type DbgpCommand interface {
	Handle() (string, error)
	GetName() string
}

type DbgpCloudInitCommand interface {
	Handle() (string, error)
	GetKey() string
	GetName() string
	AddConnection() error
	Close()
}

type DbgpInitResult interface {
	AsXML() (string, error)
	IsSuccess() bool
}

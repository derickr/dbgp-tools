package command

type DbgpCommand interface {
	Handle() (string, error)
	GetName() string
}

type DbgpCloudCommand interface {
	Handle() (string, error)
	GetKey() string
	GetName() string
	ActUponConnection() error
	Close()
}

type DbgpInitResult interface {
	AsXML() (string, error)
	IsSuccess() bool
}

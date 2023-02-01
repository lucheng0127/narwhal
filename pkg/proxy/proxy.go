package proxy

const (
	DefaultPort int = 8888
)

type Server interface {
	Launch() error
	Stop()
}

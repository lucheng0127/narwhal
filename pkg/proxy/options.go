package proxy

type Option func(s *ProxyServer)

func ListenPort(port int) Option {
	return func(s *ProxyServer) {
		s.port = port
	}
}

package proxy

type Option func(s *ProxyServer)

func ListenPort(port int) Option {
	return func(s *ProxyServer) {
		s.port = port
	}
}

func Users(users map[string]string) Option {
	return func(s *ProxyServer) {
		s.users = users
	}
}

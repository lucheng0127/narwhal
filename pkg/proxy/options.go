package proxy

type Option func(s *ProxyServer)
type COption func(c *ClientServer)

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

func Host(host string) COption {
	return func(c *ClientServer) {
		c.host = host
	}
}

func RemotePort(port uint16) COption {
	return func(c *ClientServer) {
		c.rPort = port
	}
}

func LocalPort(port uint16) COption {
	return func(c *ClientServer) {
		c.lPort = port
	}
}

func Uid(uid string) COption {
	return func(c *ClientServer) {
		c.uid = uid
	}
}

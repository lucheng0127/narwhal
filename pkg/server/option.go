package server

type ServerOption func(s *Server)

func ListenPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

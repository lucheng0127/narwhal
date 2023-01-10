package request

// Client connect to server will send a request
//
//	+----+----------+----------+
//	|VER | NMETHODS | METHODS  |
//	+----+----------+----------+
//	| 1  |    1     | 1 to 255 |
//	+----+----------+----------+
type MethodRequest struct {
	VER      byte
	NMETHODS byte
	METHODS  []byte
}

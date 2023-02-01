package proxy

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/lucheng0127/narwhal/pkg/connection"
)

func TestNewProxyServer(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want Server
	}{
		{
			name: "default port",
			args: args{},
			want: &ProxyServer{port: 8888},
		},
		{
			name: "port configured",
			args: args{
				[]Option{
					ListenPort(8001),
				},
			},
			want: &ProxyServer{port: 8001},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProxyServer(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProxyServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyServer_Launch(t *testing.T) {
	type fields struct {
		port       int
		ln         net.Listener
		users      map[string]string
		authedConn map[string]connection.Connection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Listen error",
			fields:  fields{port: 8888},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProxyServer{
				port:       tt.fields.port,
				ln:         tt.fields.ln,
				users:      tt.fields.users,
				authedConn: tt.fields.authedConn,
			}

			if tt.name == "Listen error" {
				monkey.Patch(net.Listen, func(network, address string) (net.Listener, error) {
					return nil, errors.New("Listen error")
				})
			}

			if err := s.Launch(); (err != nil) != tt.wantErr {
				t.Errorf("ProxyServer.Launch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

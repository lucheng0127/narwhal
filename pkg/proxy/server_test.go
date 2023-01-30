package proxy

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
	"github.com/lucheng0127/narwhal/mocks"
	"github.com/lucheng0127/narwhal/pkg/connection"
)

func TestNewProxyServer(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want *ProxyServer
	}{
		{
			name: "Port configured",
			args: args{
				opts: []Option{ListenPort(8888)},
			},
			want: &ProxyServer{port: 8888},
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
		users      map[string]string
		authedConn map[string]connection.Connection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "listen error",
			fields: fields{
				port:       8888,
				users:      make(map[string]string),
				authedConn: make(map[string]connection.Connection),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProxyServer{
				port:       tt.fields.port,
				users:      tt.fields.users,
				authedConn: tt.fields.authedConn,
			}
			if tt.name == "listen error" {
				monkey.Patch(net.Listen, func(network, address string) (net.Listener, error) {
					return nil, errors.New("listen error")
				})
			}
			if err := s.Launch(); (err != nil) != tt.wantErr {
				t.Errorf("ProxyServer.Launch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLaunchServer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConn := mocks.NewMockConnection(mockCtrl)
	monkey.Patch(connection.NewServerConnection, func(conn net.Conn) connection.Connection {
		return mockConn
	})
	ctx := context.Background()
	monkey.Patch(utils.NewTraceContext, func() context.Context {
		return ctx
	})
	mockConn.EXPECT().Serve(ctx).Times(1)

	server := NewProxyServer(ListenPort(8888))
	go server.Launch()
	time.Sleep(100 * time.Microsecond)
	net.Dial("tcp", "127.0.0.1:8888")
	time.Sleep(100 * time.Microsecond)
}

package proxy

import (
	"reflect"
	"testing"
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
			name: "Port not configed",
			args: args{},
			want: nil,
		},
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

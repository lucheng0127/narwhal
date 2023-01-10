package server

import (
	"reflect"
	"testing"
)

func TestNewServer(t *testing.T) {
	type args struct {
		opt []ServerOption
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "Without port configured",
			args: args{},
			want: nil,
		},
		{
			name: "Port configured",
			args: args{[]ServerOption{ListenPort(8001)}},
			want: &Server{port: 8001},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.opt...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

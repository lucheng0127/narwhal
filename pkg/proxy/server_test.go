package proxy

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lucheng0127/narwhal/mocks"
	"github.com/lucheng0127/narwhal/pkg/connection"
	"github.com/stretchr/testify/require"
)

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
			name: "8888",
			fields: fields{
				port:       8888,
				users:      make(map[string]string),
				authedConn: make(map[string]connection.Connection),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ProxyServer{
				port:       tt.fields.port,
				users:      tt.fields.users,
				authedConn: tt.fields.authedConn,
			}

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockConnection := mocks.NewMockConnection(mockCtrl)
			mockConnection.EXPECT().Serve()

			var wg sync.WaitGroup

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				go s.Launch()
				time.Sleep(1 * time.Second)
				wg.Done()
			}(&wg)

			time.Sleep(100 * time.Microsecond)
			_, err := net.Dial("tcp", "127.0.0.1:8888")
			require.NoError(t, err)
			wg.Wait()
		})
	}
}

package connection

import (
	"net"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	uuid "github.com/satori/go.uuid"
)

func TestNewServerConnection(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	go func() {
		net.Listen("tcp", "127.0.0.1:8888")
	}()
	mockConn, _ := net.Dial("tcp", "127.0.0.1:8888")
	mockAuthCtx := uuid.NewV4().String()
	mockUid := uuid.NewV4().String()

	type args struct {
		conn net.Conn
	}
	tests := []struct {
		name string
		args args
		want Connection
	}{
		{
			name: "normal",
			args: args{conn: mockConn},
			want: &SConn{
				arrs: Arrs{
					Conn:      mockConn,
					UID:       mockUid,
					AuthCtx:   mockAuthCtx,
					ProxyConn: true,
				},
			},
		},
	}
	for _, tt := range tests {

		got := NewServerConnection(tt.args.conn)
		if tt.name == "normal" {
			got.SetAuthCtx(mockAuthCtx)
			got.SetUID(mockUid)
			got.SetToProxyConn()
		}

		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServerConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}

//func TestSConn_BindAndProxy(t *testing.T) {
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//
//	mockConn := mock_net.NewMockConn(mockCtrl)
//	mockPkt := mock_protocol.NewMockPKG(mockCtrl)
//
//	monkey.Patch(
//		protocol.NewPkt,
//		func(byte, []byte) protocol.PKG {
//			return mockPkt
//		},
//	)
//	mockPkt.EXPECT().Encode().AnyTimes().Return([]byte("fake data"), nil)
//	mockConn.EXPECT().Write([]byte("fake data")).Return(1, nil)
//	mockConn.EXPECT().RemoteAddr().AnyTimes().Return("127.0.0.1:55555")
//
//	sc := NewServerConnection(mockConn)
//	go sc.BindAndProxy(8888)
//	time.Sleep(100 * time.Microsecond)
//	net.Dial("tcp", "127.0.0.1:8888")
//	time.Sleep(10 * time.Second)
//}

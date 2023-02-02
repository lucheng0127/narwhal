package proxy

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/lucheng0127/narwhal/mocks"
	"github.com/lucheng0127/narwhal/pkg/connection"
	"github.com/lucheng0127/narwhal/pkg/protocol"
	uuid "github.com/satori/go.uuid"
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
			name: "port user configured",
			args: args{
				[]Option{
					ListenPort(8001),
					Users(map[string]string{"user": "0"}),
				},
			},
			want: &ProxyServer{port: 8001, users: map[string]string{"user": "0"}},
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

func TestProxyServer_availabledPort(t *testing.T) {
	type fields struct {
		port       int
		ln         net.Listener
		users      map[string]string
		authedConn map[string]connection.Connection
	}
	type args struct {
		authCtx string
		port    int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "all",
			fields: fields{
				users: map[string]string{"user": "0"},
			},
			args: args{
				authCtx: "user",
				port:    22,
			},
			want: true,
		},
		{
			name: "single port ok",
			fields: fields{
				users: map[string]string{"user": "22"},
			},
			args: args{
				authCtx: "user",
				port:    22,
			},
			want: true,
		},
		{
			name: "single port not ok",
			fields: fields{
				users: map[string]string{"user": "80"},
			},
			args: args{
				authCtx: "user",
				port:    22,
			},
			want: false,
		},
		{
			name: "multi port ok",
			fields: fields{
				users: map[string]string{"user": "22,80"},
			},
			args: args{
				authCtx: "user",
				port:    22,
			},
			want: true,
		},
		{
			name: "multi port not ok",
			fields: fields{
				users: map[string]string{"user": "22,80"},
			},
			args: args{
				authCtx: "user",
				port:    8000,
			},
			want: false,
		},
		{
			name: "port range ok",
			fields: fields{
				users: map[string]string{"user": "8000-8100"},
			},
			args: args{
				authCtx: "user",
				port:    8100,
			},
			want: true,
		},
		{
			name: "port range not ok",
			fields: fields{
				users: map[string]string{"user": "8000-8100"},
			},
			args: args{
				authCtx: "user",
				port:    8101,
			},
			want: false,
		},
		{
			name: "user not exist",
			fields: fields{
				users: map[string]string{"user": "8000-8100"},
			},
			args: args{
				authCtx: "user1",
				port:    8101,
			},
			want: false,
		},
		{
			name: "error format port range 1",
			fields: fields{
				users: map[string]string{"user": "8000-8050-8100"},
			},
			args: args{
				authCtx: "user",
				port:    8101,
			},
			want: false,
		},
		{
			name: "error format port range 2",
			fields: fields{
				users: map[string]string{"user": "xx-8100"},
			},
			args: args{
				authCtx: "user",
				port:    8101,
			},
			want: false,
		},
		{
			name: "error format port range 3",
			fields: fields{
				users: map[string]string{"user": "8000-xx"},
			},
			args: args{
				authCtx: "user",
				port:    8101,
			},
			want: false,
		},
		{
			name: "error format multi port",
			fields: fields{
				users: map[string]string{"user": "xx,8100"},
			},
			args: args{
				authCtx: "user",
				port:    8101,
			},
			want: false,
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
			if got := s.availabledPort(tt.args.authCtx, tt.args.port); got != tt.want {
				t.Errorf("ProxyServer.availabledPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyServer_getUserByUid(t *testing.T) {
	type fields struct {
		port       int
		ln         net.Listener
		users      map[string]string
		authedConn map[string]connection.Connection
	}
	type args struct {
		uid string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "user exist",
			fields: fields{users: map[string]string{"user": "0"}},
			args:   args{uid: "user"},
			want:   "user",
		},
		{
			name:   "user not exist",
			fields: fields{users: map[string]string{"user": "0"}},
			args:   args{uid: "user1"},
			want:   "",
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
			if got := s.getUserByUid(tt.args.uid); got != tt.want {
				t.Errorf("ProxyServer.getUserByUid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyServer_getAuthedConn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConn := mocks.NewMockConnection(mockCtrl)

	type fields struct {
		port       int
		ln         net.Listener
		users      map[string]string
		authedConn map[string]connection.Connection
	}
	type args struct {
		authCtx string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   connection.Connection
	}{
		{
			name: "connection exist",
			fields: fields{
				authedConn: map[string]connection.Connection{
					"123": mockConn,
				},
			},
			args: args{authCtx: "123"},
			want: mockConn,
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
			if got := s.getAuthedConn(tt.args.authCtx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProxyServer.getAuthedConn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyServer_auth(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConn := mocks.NewMockConnection(mockCtrl)
	mockConn.EXPECT().GetArrs().AnyTimes()
	mockConn.EXPECT().SetAuthCtx(gomock.Any()).AnyTimes()
	mockConn.EXPECT().SetUID(gomock.Any()).AnyTimes()

	mockPkt := mocks.NewMockPKG(mockCtrl)
	mockPayload := mocks.NewMockPL(mockCtrl)
	mockUuid := uuid.NewV4()

	type fields struct {
		port       int
		ln         net.Listener
		users      map[string]string
		authedConn map[string]connection.Connection
	}
	type args struct {
		conn connection.Connection
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "read pkt error",
			fields:  fields{users: map[string]string{"user": "0"}},
			args:    args{conn: mockConn},
			want:    "",
			wantErr: true,
		},
		{
			name:    "auth ok",
			fields:  fields{users: map[string]string{"user": "0"}},
			args:    args{conn: mockConn},
			want:    mockUuid.String(),
			wantErr: false,
		},
		{
			name:    "auth not ok",
			fields:  fields{users: map[string]string{"user": "0"}},
			args:    args{conn: mockConn},
			want:    "",
			wantErr: true,
		},
		{
			name: "pconn ok",
			fields: fields{
				users: map[string]string{"user": "0"},
				authedConn: map[string]connection.Connection{
					mockUuid.String(): mockConn,
				},
			},
			args:    args{conn: mockConn},
			want:    "",
			wantErr: false,
		},
		{
			name: "pconn not ok",
			fields: fields{
				users: map[string]string{"user": "0"},
				authedConn: map[string]connection.Connection{
					mockUuid.String(): mockConn,
				},
			},
			args:    args{conn: mockConn},
			want:    "",
			wantErr: true,
		},
		{
			name: "invalidate request",
			fields: fields{
				users: map[string]string{"user": "0"},
				authedConn: map[string]connection.Connection{
					mockUuid.String(): mockConn,
				},
			},
			args:    args{conn: mockConn},
			want:    "",
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

			if tt.name == "read pkt error" {
				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return nil, errors.New("read pkt error")
					},
				)
			} else if tt.name == "auth ok" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepAuth)
				mockPkt.EXPECT().GetPayload().Return(mockPayload)
				mockPayload.EXPECT().String().Return("user")

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)

				monkey.Patch(uuid.NewV4, func() uuid.UUID {
					return mockUuid
				})
			} else if tt.name == "auth not ok" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepAuth)
				mockPkt.EXPECT().GetPayload().Return(mockPayload)
				mockPayload.EXPECT().String().Return("user1")

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)

				monkey.Patch(uuid.NewV4, func() uuid.UUID {
					return mockUuid
				})
			} else if tt.name == "pconn ok" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepPConn)
				mockPkt.EXPECT().GetPayload().Return(mockPayload)
				mockPayload.EXPECT().String().Return(mockUuid.String())
				mockConn.EXPECT().SetToProxyConn().Times(1)

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)

				monkey.Patch(uuid.NewV4, func() uuid.UUID {
					return mockUuid
				})
			} else if tt.name == "pconn not ok" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepPConn)
				mockPkt.EXPECT().GetPayload().Return(mockPayload)
				mockPayload.EXPECT().String().Return("123")

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)

				monkey.Patch(uuid.NewV4, func() uuid.UUID {
					return mockUuid
				})
			} else if tt.name == "invalidate request" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepNone)

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)

				monkey.Patch(uuid.NewV4, func() uuid.UUID {
					return mockUuid
				})
			}

			got, err := s.auth(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProxyServer.auth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProxyServer.auth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyServer_bind(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConn := mocks.NewMockConnection(mockCtrl)
	mockConn.EXPECT().GetArrs().Return(connection.Arrs{UID: "user"}).AnyTimes()
	mockPkt := mocks.NewMockPKG(mockCtrl)
	mockPayload := mocks.NewMockPL(mockCtrl)

	type fields struct {
		port       int
		ln         net.Listener
		users      map[string]string
		authedConn map[string]connection.Connection
	}
	type args struct {
		conn connection.Connection
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "bind ok",
			fields: fields{
				users: map[string]string{"user": "22"},
			},
			args:    args{conn: mockConn},
			want:    22,
			wantErr: false,
		},
		{
			name: "invalidate bport",
			fields: fields{
				users: map[string]string{"user": "22"},
			},
			args:    args{conn: mockConn},
			want:    -1,
			wantErr: true,
		},
		{
			name: "not permitted bport",
			fields: fields{
				users: map[string]string{"user": "22"},
			},
			args:    args{conn: mockConn},
			want:    -1,
			wantErr: true,
		},
		{
			name: "invalidate req",
			fields: fields{
				users: map[string]string{"user": "22"},
			},
			args:    args{conn: mockConn},
			want:    -1,
			wantErr: true,
		},
		{
			name: "parse error",
			fields: fields{
				users: map[string]string{"user": "22"},
			},
			args:    args{conn: mockConn},
			want:    -1,
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

			if tt.name == "bind ok" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepBind)
				mockPkt.EXPECT().GetPayload().Return(mockPayload)
				mockPayload.EXPECT().Int().Return(22)

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)
			} else if tt.name == "invalidate bport" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepBind)
				mockPkt.EXPECT().GetPayload().Return(mockPayload)
				mockPayload.EXPECT().Int().Return(-1)

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)
			} else if tt.name == "not permitted bport" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepBind)
				mockPkt.EXPECT().GetPayload().Return(mockPayload)
				mockPayload.EXPECT().Int().Return(8000)

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)
			} else if tt.name == "invalidate req" {
				mockPkt.EXPECT().GetPCode().Return(protocol.RepNone)

				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return mockPkt, nil
					},
				)
			} else if tt.name == "parse error" {
				monkey.Patch(
					protocol.ReadFromConn,
					func(conn net.Conn) (protocol.PKG, error) {
						return nil, errors.New("parse error")
					},
				)
			}

			got, err := s.bind(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProxyServer.bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProxyServer.bind() = %v, want %v", got, tt.want)
			}
		})
	}
}

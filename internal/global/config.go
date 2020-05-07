package global

import (
	"github.com/profzone/eden-framework/pkg/courier/transport_grpc"
	"github.com/profzone/eden-framework/pkg/courier/transport_http"
	"github.com/sirupsen/logrus"
	"longhorn/upload-packager/pkg/consensus"
)

var Config = struct {
	LogLevel logrus.Level

	// administrator
	GRPCServer transport_grpc.ServeGRPC
	HTTPServer transport_http.ServeHTTP

	// consensus
	Raft *consensus.Raft
}{
	LogLevel: logrus.DebugLevel,

	GRPCServer: transport_grpc.ServeGRPC{
		Port: 8900,
	},
	HTTPServer: transport_http.ServeHTTP{
		Port:     8001,
		WithCORS: true,
	},

	Raft: &consensus.Raft{
		ListenAddr:        "127.0.0.1:8101",
		DataDir:           "./build/node1",
		DataPrefix:        "",
		BootstrapAsLeader: true,
		JoinAddr:          "",
	},
}

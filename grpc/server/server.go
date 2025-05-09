package server

import (
	"fmt"
	"foundation-go/log"
	"foundation-go/utility/envutil"
	"net"
	"os"
	"strconv"
	"strings"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func init() {
	grpc_logrus.ReplaceGrpcLogger(logrus.NewEntry(logrus.StandardLogger()))
}

// RunGRPCServer TODO: 标准化，Discovery、IP、docker container name直连
func RunGRPCServer(serviceName string, registerServer func(server *grpc.Server)) {
	port := os.Getenv("PORT")
	addr := ":" + port
	// e2e test的时候读环境变量ip直连
	if ok, err := strconv.ParseBool(os.Getenv("E2E")); envutil.IsDev() || (err == nil && ok) {
		addr = getE2EGRPCAddr(serviceName)
	}
	RunGRPCServerOnAddr(addr, registerServer)
}

func getE2EGRPCAddr(serviceName string) string {
	key := fmt.Sprintf("%s_GRPC_ADDR", strings.ToUpper(serviceName))
	return os.Getenv(key)
}

func RunGRPCServerOnAddr(addr string, registerServer func(server *grpc.Server)) {
	logrusEntry := logrus.NewEntry(logrus.StandardLogger())
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpc_tags.UnaryServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
			log.GRPCUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_tags.StreamServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry),
		),
	)
	registerServer(grpcServer)

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Infof("Starting gRPC server, Listening: %s", addr)
	if err := grpcServer.Serve(listen); err != nil {
		logrus.Panic(err)
	}
}

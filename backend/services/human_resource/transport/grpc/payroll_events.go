package grpc

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/payrollevents"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewPayrollEventServer membuat server privat dengan CA klien workload yang terpisah.
func NewPayrollEventServer(certFile, keyFile, clientCAFile string, receiver payrollevents.Receiver) (*stdgrpc.Server, error) {
	tlsConfig, err := loadServerTLS(certFile, keyFile, clientCAFile)
	if err != nil {
		return nil, err
	}
	server := stdgrpc.NewServer(
		stdgrpc.Creds(credentials.NewTLS(tlsConfig)),
	)
	if err := payrollevents.Register(server, receiver); err != nil {
		return nil, err
	}
	return server, nil
}

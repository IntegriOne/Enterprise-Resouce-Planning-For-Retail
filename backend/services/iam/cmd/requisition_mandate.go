package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/terminationaccess"
	"github.com/integrione/erp-retail/backend/services/iam/services/provisioning"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func startHRMWorkloads(cfg *config.Config, users *user.Service, provisioningService *provisioning.Service, logger *slog.Logger, errorsOut chan<- error) (func(), error) {
	var verifier user.TerminationAccessVerifier
	var conn *grpc.ClientConn
	if access := cfg.HRMTerminationAccess; access != (config.HRMTerminationAccessConfig{}) {
		if cfg.RequisitionMandate == (config.RequisitionMandateConfig{}) {
			return nil, errors.New("employee termination requires the private IAM workload listener")
		}
		clientConn, client, err := terminationaccess.Dial(access.Target, access.ServerName, access.CertFile, access.KeyFile, access.CAFile)
		if err != nil {
			return nil, fmt.Errorf("configure HRM termination access verifier: %w", err)
		}
		conn, verifier = clientConn, client
		if provisioningService != nil {
			provisioningService.ConfigureWorkerAccountLinkVerifier(client)
		}
	}
	stop, err := startRequisitionMandate(cfg.RequisitionMandate, users, verifier, logger, errorsOut)
	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil, fmt.Errorf("start IAM HRM workload listener: %w", err)
	}
	return func() {
		stop()
		if conn != nil {
			_ = conn.Close()
		}
	}, nil
}

func startRequisitionMandate(cfg config.RequisitionMandateConfig, users *user.Service, verifier user.TerminationAccessVerifier, logger *slog.Logger, errorsOut chan<- error) (func(), error) {
	if cfg == (config.RequisitionMandateConfig{}) {
		return func() {}, nil
	}
	tlsConfig, err := requisitionMandateTLS(cfg)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.ChainUnaryInterceptor(interceptor.RecoveryUnaryInterceptor(logger), interceptor.LoggingUnaryInterceptor(logger)),
	)
	mapper := handler.NewErrorMapper(logger)
	handler.RegisterRequisitionMandate(server, users, mapper)
	if verifier != nil {
		handler.RegisterEmployeeTermination(server, users, verifier, mapper)
	}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errorsOut <- fmt.Errorf("serve requisition mandate: %w", err)
		}
	}()
	logger.Info("starting IAM requisition mandate listener", slog.String("address", listener.Addr().String()))
	return server.Stop, nil
}

func requisitionMandateTLS(cfg config.RequisitionMandateConfig) (*tls.Config, error) {
	if cfg.Address == "" || cfg.CertFile == "" || cfg.KeyFile == "" || cfg.ClientCAFile == "" {
		return nil, errors.New("requisition mandate address, certificate, key, and client CA are required")
	}
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load requisition mandate server certificate: %w", err)
	}
	ca, err := os.ReadFile(cfg.ClientCAFile)
	if err != nil {
		return nil, fmt.Errorf("read requisition mandate client CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, errors.New("invalid requisition mandate client CA")
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}, nil
}

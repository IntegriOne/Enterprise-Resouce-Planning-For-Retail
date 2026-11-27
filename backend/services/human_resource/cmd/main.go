package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/integrione/erp-retail/backend/services/human_resource/config"
	iamintegration "github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/integration/iam"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/outbox"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	transportgrpc "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	stdgrpc "google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("HRM runtime stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load HRM configuration: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	client, err := authsession.NewTLSClient(cfg.SecurityRedis)
	if err != nil {
		return errors.New("configure HRM session Redis TLS client failed")
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error("close HRM session Redis client failed")
		}
	}()
	if err := client.Ping(startup).Err(); err != nil {
		return errors.New("connect HRM session Redis failed")
	}
	store, err := authsession.NewStore(client, authsession.NewKeyspace(cfg.SecurityNamespace))
	if err != nil {
		return errors.New("configure HRM session store failed")
	}
	policy, err := security.NewPolicy(store, security.PermissionCatalog())
	if err != nil {
		return fmt.Errorf("configure HRM security policy: %w", err)
	}
	serviceOptions := []appservice.RegistryOption{
		appservice.WithPrincipalLinkVerifier(workerLinkVerifier{key: cfg.GatewayHMACSecret}),
	}
	var terminationPublisher outbox.Publisher
	if cfg.IAMMandateTarget != "" || cfg.IAMMandateTLS != (config.TLSConfig{}) {
		tls := cfg.IAMMandateTLS
		conn, mandate, err := iamintegration.DialRequisitionMandate(cfg.IAMMandateTarget, tls.ServerName, tls.CertFile, tls.KeyFile, tls.CAFile)
		if err != nil {
			return fmt.Errorf("initialize IAM requisition mandate client: %w", err)
		}
		defer conn.Close()
		serviceOptions = append(serviceOptions, appservice.WithRequisitionMandateVerifier(mandate))
		if cfg.TerminationAccessAddress != "" {
			terminationPublisher = mandate
		}
	}
	// TODO(finance_accounting): Pasang adapter bukti anggaran, clearance, jurnal, penerbitan,
	// pembayaran, dan settlement saat layanan Finance tersedia. Lihat agents/notes/finance_accounting.
	// TODO(document_records_management): Pasang adapter bukti dan penyimpanan PDF privat
	// saat layanan DRM tersedia. Lihat agents/notes/document_records_management.
	dependencies, err := buildDependencies(startup, cfg, policy, logger, terminationPublisher, serviceOptions...)
	if err != nil {
		return fmt.Errorf("initialize HRM dependencies: %w", err)
	}
	defer func() {
		if err := dependencies.Close(); err != nil {
			logger.Error("close HRM database failed")
		}
	}()
	server, err := transportgrpc.NewServer(transportgrpc.Config{
		Address: cfg.GRPCAddress(), ServiceName: cfg.ServiceName, TLSCertFile: cfg.TLS.CertFile,
		TLSKeyFile: cfg.TLS.KeyFile, TLSCAFile: cfg.TLS.CAFile,
		GatewayHMACSecret: cfg.GatewayHMACSecret, ShutdownTimeout: cfg.ShutdownTimeout,
	}, dependencies.Services, logger)
	if err != nil {
		return fmt.Errorf("configure HRM gRPC: %w", err)
	}
	if err := startup.Err(); err != nil {
		return err
	}
	workloadFailed := make(chan error, 2)
	terminationTLS := cfg.TerminationAccessTLS
	stopTermination, err := transportgrpc.StartTerminationAccess(cfg.TerminationAccessAddress,
		terminationTLS.CertFile, terminationTLS.KeyFile, terminationTLS.CAFile, dependencies.Services.Personnel, logger, workloadFailed)
	if err != nil {
		return fmt.Errorf("configure HRM termination access: %w", err)
	}
	defer stopTermination()

	var payrollServer *stdgrpc.Server
	if cfg.PayrollEventsAddress != "" {
		if !dependencies.Services.Payroll.PayrollEventsReady() {
			return errors.New("HRM payroll events requires a configured payroll event provider")
		}
		tls := cfg.PayrollEventsTLS
		payrollServer, err = transportgrpc.NewPayrollEventServer(tls.CertFile, tls.KeyFile, tls.CAFile, dependencies.Services.Payroll)
		if err != nil {
			return fmt.Errorf("configure HRM payroll events: %w", err)
		}
		listener, err := net.Listen("tcp", cfg.PayrollEventsAddress)
		if err != nil {
			return fmt.Errorf("listen HRM payroll events: %w", err)
		}
		defer payrollServer.Stop()
		logger.Info("starting HRM payroll events listener", "address", listener.Addr().String())
		go func() {
			if err := payrollServer.Serve(listener); err != nil && !errors.Is(err, stdgrpc.ErrServerStopped) {
				workloadFailed <- fmt.Errorf("serve HRM payroll events: %w", err)
			}
		}()
	}
	stopRPC := func(shutdown context.Context) error {
		if payrollServer == nil {
			return server.Stop(shutdown)
		}
		payrollStopped := make(chan error, 1)
		go func() { payrollStopped <- transportgrpc.DrainServer(shutdown, payrollServer) }()
		return errors.Join(server.Stop(shutdown), <-payrollStopped)
	}

	cancel()
	logger.Info("HRM mutation validators registered for all catalog resources")
	logger.Info("starting private HRM gRPC listener", "address", cfg.GRPCAddress())
	served := make(chan error, 1)
	relayed := make(chan error, 1)
	go func() { served <- server.ListenAndServe() }()
	go func() { relayed <- dependencies.OutboxRelay.Run(ctx) }()
	select {
	case err := <-served:
		stop()
		var payrollErr error
		if payrollServer != nil {
			shutdown, finish := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			payrollErr = transportgrpc.DrainServer(shutdown, payrollServer)
			finish()
		}
		if relayErr := <-relayed; relayErr != nil {
			return errors.Join(err, relayErr, payrollErr)
		}
		if err == nil {
			err = errors.New("HRM gRPC server stopped unexpectedly")
		}
		return errors.Join(err, payrollErr)
	case err := <-relayed:
		shutdown, finish := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer finish()
		stopErr := stopRPC(shutdown)
		serveErr := <-served
		if err == nil && ctx.Err() != nil {
			return errors.Join(stopErr, serveErr)
		}
		if err == nil {
			err = errors.New("HRM outbox relay stopped unexpectedly")
		}
		return errors.Join(err, stopErr, serveErr)
	case err := <-workloadFailed:
		stop()
		shutdown, finish := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer finish()
		return errors.Join(err, stopRPC(shutdown), <-served, <-relayed)

	case <-ctx.Done():
		shutdown, finish := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer finish()
		stopErr := stopRPC(shutdown)
		serveErr := <-served
		relayErr := <-relayed
		return errors.Join(stopErr, serveErr, relayErr)
	}
}

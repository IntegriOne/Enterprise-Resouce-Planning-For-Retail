// Package iam menghubungkan HRM ke kontrak privat Identity and Access Management.
package iam

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// RequisitionMandateClient meminta bukti mandat unit untuk principal peninjau.
type RequisitionMandateClient struct {
	client      iamv1.RequisitionMandateServiceClient
	termination iamv1.EmployeeTerminationServiceClient
}

// DialRequisitionMandate membuat koneksi privat IAM yang memverifikasi sertifikat server.
func DialRequisitionMandate(target, serverName, certFile, keyFile, caFile string) (*grpc.ClientConn, *RequisitionMandateClient, error) {
	if target == "" || serverName == "" || certFile == "" || keyFile == "" || caFile == "" {
		return nil, nil, fmt.Errorf("%w: IAM mandate target and TLS files are required", hrerrors.ErrInvalidInput)
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("load IAM mandate client certificate: %w", err)
	}
	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read IAM mandate server CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return nil, nil, fmt.Errorf("%w: invalid IAM mandate server CA", hrerrors.ErrInvalidInput)
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert}, RootCAs: roots, ServerName: serverName,
	})))
	if err != nil {
		return nil, nil, err
	}
	return conn, &RequisitionMandateClient{client: iamv1.NewRequisitionMandateServiceClient(conn), termination: iamv1.NewEmployeeTerminationServiceClient(conn)}, nil
}

// VerifyRequisitionMandate mencocokkan keputusan IAM dengan principal dan unit yang diminta.
func (c *RequisitionMandateClient) VerifyRequisitionMandate(ctx context.Context, principalID, unitID string) error {
	if c == nil || c.client == nil {
		return hrerrors.ErrUnavailable
	}
	if entity.ValidateUUID("principal_id", principalID) != nil || entity.ValidateUUID("organization_unit_id", unitID) != nil {
		return hrerrors.ErrReferenceConflict
	}
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	result, err := c.client.VerifyRequisitionMandate(callCtx, &iamv1.VerifyRequisitionMandateRequest{
		PrincipalId: principalID, OrganizationUnitId: unitID,
	})
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return hrerrors.ErrUnavailable
	}
	if result == nil || !strings.EqualFold(result.GetPrincipalId(), principalID) ||
		!strings.EqualFold(result.GetOrganizationUnitId(), unitID) {
		return hrerrors.ErrUnavailable
	}
	if !result.GetAllowed() {
		return hrerrors.ErrForbidden
	}
	return nil
}

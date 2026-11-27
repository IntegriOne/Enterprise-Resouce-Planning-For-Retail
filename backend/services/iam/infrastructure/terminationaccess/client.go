// Package terminationaccess membaca keputusan penutupan akun dari workload HRM.
package terminationaccess

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workload/termination/v1"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Client struct {
	rpc pb.TerminationAccessServiceClient
}

var _ user.TerminationAccessVerifier = (*Client)(nil)

// Dial membuka koneksi mTLS ke listener workload HRM.
func Dial(target, serverName, certFile, keyFile, caFile string) (*grpc.ClientConn, *Client, error) {
	if target == "" || serverName == "" || certFile == "" || keyFile == "" || caFile == "" {
		return nil, nil, fmt.Errorf("%w: termination access target and TLS settings are required", iamerrors.ErrInvalidInput)
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("load termination access client certificate: %w", err)
	}
	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, nil, fmt.Errorf("read termination access server CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return nil, nil, fmt.Errorf("%w: invalid termination access server CA", iamerrors.ErrInvalidInput)
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert}, RootCAs: roots, ServerName: serverName,
	})))
	if err != nil {
		return nil, nil, err
	}
	return conn, &Client{rpc: pb.NewTerminationAccessServiceClient(conn)}, nil
}

// Verify memeriksa korelasi respons sebelum keputusan dapat dipakai IAM.
func (c *Client) Verify(ctx context.Context, workerID, planID, effectiveDate string) (user.TerminationAccessDecision, error) {
	if c == nil || c.rpc == nil {
		return user.TerminationAccessDecision{}, iamerrors.ErrUnavailable
	}
	if !entity.IsValidUUID(workerID) || !entity.IsValidUUID(planID) {
		return user.TerminationAccessDecision{}, iamerrors.ErrInvalidInput
	}
	req := &pb.VerifyTerminationAccessRequest{WorkerId: workerID, OffboardingPlanId: planID}
	if effectiveDate != "" {
		if _, err := time.Parse(time.DateOnly, effectiveDate); err != nil {
			return user.TerminationAccessDecision{}, iamerrors.ErrInvalidInput
		}
		req.TerminationEffectiveDate = &effectiveDate
	}
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	result, err := c.rpc.VerifyTerminationAccess(callCtx, req)
	if err != nil {
		if ctx.Err() != nil {
			return user.TerminationAccessDecision{}, ctx.Err()
		}
		return user.TerminationAccessDecision{}, iamerrors.ErrUnavailable
	}
	date, dateErr := time.Parse(time.DateOnly, result.GetTerminationEffectiveDate())
	principal := result.GetPrincipalId()
	if !strings.EqualFold(result.GetWorkerId(), workerID) || !strings.EqualFold(result.GetOffboardingPlanId(), planID) ||
		dateErr != nil || date.IsZero() || (effectiveDate != "" && result.GetTerminationEffectiveDate() != effectiveDate) ||
		((principal != "" || result.GetMayDisable()) && !entity.IsValidUUID(principal)) {
		return user.TerminationAccessDecision{}, iamerrors.ErrUnavailable
	}
	return user.TerminationAccessDecision{PrincipalID: principal, EffectiveDate: result.GetTerminationEffectiveDate(), MayDisable: result.GetMayDisable()}, nil
}

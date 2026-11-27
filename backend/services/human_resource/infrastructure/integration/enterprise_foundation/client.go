// Package enterprisefoundation mengimplementasikan klien gRPC fail-closed ke layanan Enterprise Foundation.
package enterprisefoundation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

const defaultTimeout = 5 * time.Second

// Client memverifikasi master badan hukum, unit organisasi, lokasi, dan party ke Enterprise Foundation.
type Client struct {
	enabled       bool
	timeout       time.Duration
	hmacSecret    []byte
	legalEntities efv1.LegalEntityServiceClient
	orgUnits      efv1.OrganizationUnitServiceClient
	locations     efv1.LocationServiceClient
	parties       efv1.PartyServiceClient
}

var (
	_ security.PositionReferenceVerifier   = (*Client)(nil)
	_ security.WorkerReferenceVerifier     = (*Client)(nil)
	_ security.FoundationVerifier          = (*Client)(nil)
	_ security.CandidateFoundationVerifier = (*Client)(nil)
)

// NewDisabledClient membentuk adapter nonaktif yang selalu gagal tertutup saat dipanggil.
func NewDisabledClient() *Client {
	return &Client{enabled: false, timeout: defaultTimeout}
}

// NewClient membangun adapter Enterprise Foundation dari koneksi gRPC yang telah tersedia.
func NewClient(conn grpc.ClientConnInterface, hmacSecret []byte, timeout time.Duration) *Client {
	if conn == nil {
		return &Client{enabled: true, timeout: defaultTimeout}
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Client{
		enabled:       true,
		timeout:       timeout,
		hmacSecret:    append([]byte(nil), hmacSecret...),
		legalEntities: efv1.NewLegalEntityServiceClient(conn),
		orgUnits:      efv1.NewOrganizationUnitServiceClient(conn),
		locations:     efv1.NewLocationServiceClient(conn),
		parties:       efv1.NewPartyServiceClient(conn),
	}
}

// Dial membuka koneksi gRPC non-blocking untuk adapter yang menandatangani setiap panggilan.
func Dial(target string, hmacSecret []byte, timeout time.Duration) (*grpc.ClientConn, *Client, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" || len(hmacSecret) < 32 {
		return nil, nil, fmt.Errorf("%w: enterprise foundation client target and HMAC secret are required", hrerrors.ErrUnavailable)
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	conn, err := grpc.NewClient(
		trimmed,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(false)),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  250 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   2 * time.Second,
			},
			MinConnectTimeout: 2 * time.Second,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: initialize enterprise foundation client failed", hrerrors.ErrUnavailable)
	}
	return conn, NewClient(conn, hmacSecret, timeout), nil
}

// VerifyPositionReferences memeriksa badan hukum, unit organisasi, lokasi opsional, serta rentang efektivitas entitas.
func (c *Client) VerifyPositionReferences(ctx context.Context, check security.PositionReferenceCheck) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := entity.ValidateUUID("legal_entity_reference_id", check.LegalEntityID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	if err := entity.ValidateUUID("organization_unit_reference_id", check.OrganizationUnitID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	if check.WorkLocationID != nil {
		if err := entity.ValidateUUID("work_location_reference_id", *check.WorkLocationID); err != nil {
			return hrerrors.ErrReferenceConflict
		}
	}
	le, err := c.fetchLegalEntity(ctx, check.LegalEntityID)
	if err != nil {
		return err
	}
	if err := validateLegalEntityWindow(le, check.ValidFrom, check.ValidTo); err != nil {
		return err
	}
	if err := c.VerifyOrganizationUnit(ctx, check.OrganizationUnitID, check.LegalEntityID); err != nil {
		return err
	}
	if check.WorkLocationID != nil {
		if err := c.VerifyLocation(ctx, *check.WorkLocationID, check.LegalEntityID); err != nil {
			return err
		}
	}
	return nil
}

// VerifyWorkerReferences memastikan party adalah individu aktif dan badan hukum utama berstatus aktif.
func (c *Client) VerifyWorkerReferences(ctx context.Context, check security.WorkerReferenceCheck) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := c.VerifyParty(ctx, check.PersonPartyID, true); err != nil {
		return err
	}
	if err := entity.ValidateUUID("legal_entity_reference_id", check.LegalEntityID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	legal, err := c.fetchLegalEntity(ctx, check.LegalEntityID)
	if err != nil {
		return err
	}
	for _, date := range []*time.Time{check.HireDate, check.ServiceStartDate} {
		if date != nil {
			if err := validateLegalEntityWindow(legal, *date, date); err != nil {
				return err
			}
		}
	}
	return nil
}

// VerifyLegalEntity memastikan badan hukum terdaftar aktif dan belum dihapus.
func (c *Client) VerifyLegalEntity(ctx context.Context, legalEntityID string) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := entity.ValidateUUID("legal_entity_reference_id", legalEntityID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	_, err := c.fetchLegalEntity(ctx, legalEntityID)
	return err
}

// VerifyLegalEntityWindow memastikan badan hukum aktif mencakup rentang tanggal bisnis.
func (c *Client) VerifyLegalEntityWindow(ctx context.Context, legalEntityID string, start, end time.Time) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := entity.ValidateUUID("legal_entity_reference_id", legalEntityID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	legal, err := c.fetchLegalEntity(ctx, legalEntityID)
	if err != nil {
		return err
	}
	return validateLegalEntityWindow(legal, start, &end)
}

// VerifyOrganizationUnit memastikan unit organisasi aktif, belum dihapus, dan terikat ke badan hukum bila diberikan.
func (c *Client) VerifyOrganizationUnit(ctx context.Context, organizationUnitID, legalEntityID string) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := entity.ValidateUUID("organization_unit_reference_id", organizationUnitID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	if legalEntityID != "" {
		if err := entity.ValidateUUID("legal_entity_reference_id", legalEntityID); err != nil {
			return hrerrors.ErrReferenceConflict
		}
	}
	rpcCtx, cancel := c.callContext(ctx)
	defer cancel()
	resp, err := c.orgUnits.GetOrganizationUnitById(rpcCtx, &efv1.GetOrganizationUnitByIdRequest{Id: organizationUnitID})
	if err != nil {
		return mapRPCError(ctx, err)
	}
	if resp == nil {
		return hrerrors.ErrUnavailable
	}
	unit := resp.GetOrganizationUnit()
	if unit == nil || !strings.EqualFold(unit.GetId(), organizationUnitID) || unit.GetDeletedAt() != nil || !unit.GetIsActive() {
		return hrerrors.ErrReferenceConflict
	}
	if legalEntityID != "" && !strings.EqualFold(unit.GetLegalEntityId(), legalEntityID) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// VerifyLocation memastikan lokasi aktif, belum dihapus, dan terikat ke badan hukum bila diberikan.
func (c *Client) VerifyLocation(ctx context.Context, locationID, legalEntityID string) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := entity.ValidateUUID("location_reference_id", locationID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	if legalEntityID != "" {
		if err := entity.ValidateUUID("legal_entity_reference_id", legalEntityID); err != nil {
			return hrerrors.ErrReferenceConflict
		}
	}
	rpcCtx, cancel := c.callContext(ctx)
	defer cancel()
	resp, err := c.locations.GetLocationById(rpcCtx, &efv1.GetLocationByIdRequest{Id: locationID})
	if err != nil {
		return mapRPCError(ctx, err)
	}
	if resp == nil {
		return hrerrors.ErrUnavailable
	}
	loc := resp.GetLocation()
	if loc == nil || !strings.EqualFold(loc.GetId(), locationID) || loc.GetDeletedAt() != nil || !loc.GetIsActive() {
		return hrerrors.ErrReferenceConflict
	}
	if legalEntityID != "" && !strings.EqualFold(loc.GetLegalEntityId(), legalEntityID) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// VerifyParty memastikan mitra/orang terdaftar aktif, belum dihapus, dan berjenis individu bila disyaratkan.
func (c *Client) VerifyParty(ctx context.Context, partyID string, requireIndividual bool) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if err := entity.ValidateUUID("party_reference_id", partyID); err != nil {
		return hrerrors.ErrReferenceConflict
	}
	rpcCtx, cancel := c.callContext(ctx)
	defer cancel()
	resp, err := c.parties.GetPartyById(rpcCtx, &efv1.GetPartyByIdRequest{Id: partyID})
	if err != nil {
		return mapRPCError(ctx, err)
	}
	if resp == nil {
		return hrerrors.ErrUnavailable
	}
	party := resp.GetParty()
	if party == nil || !strings.EqualFold(party.GetId(), partyID) || party.GetDeletedAt() != nil || party.GetStatus() != efv1.PartyStatus_PARTY_STATUS_ACTIVE {
		return hrerrors.ErrReferenceConflict
	}
	if requireIndividual && party.GetKind() != efv1.PartyKind_PARTY_KIND_INDIVIDUAL {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// VerifyPartyContact memeriksa kontak aktif dengan ID dan party pemilik yang sesuai.
func (c *Client) VerifyPartyContact(ctx context.Context, contactID, partyID string) error {
	if err := c.ready(ctx); err != nil {
		return err
	}
	if entity.ValidateUUID("contact_reference_id", contactID) != nil ||
		entity.ValidateUUID("person_party_reference_id", partyID) != nil {
		return hrerrors.ErrReferenceConflict
	}
	rpcCtx, cancel := c.callContext(ctx)
	defer cancel()
	resp, err := c.parties.GetPartyContactById(rpcCtx, &efv1.GetPartyContactByIdRequest{ContactId: contactID})
	if err != nil {
		return mapRPCError(ctx, err)
	}
	if resp == nil {
		return hrerrors.ErrUnavailable
	}
	contact := resp.GetContact()
	if contact == nil || !strings.EqualFold(contact.GetId(), contactID) ||
		!strings.EqualFold(contact.GetPartyId(), partyID) || !contact.GetIsActive() || contact.GetDeletedAt() != nil {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (c *Client) ready(ctx context.Context) error {
	if ctx == nil {
		return hrerrors.ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if c == nil || !c.enabled || c.legalEntities == nil || c.orgUnits == nil || c.locations == nil || c.parties == nil {
		return hrerrors.ErrUnavailable
	}
	return nil
}

func (c *Client) callContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := c.timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining <= timeout {
			return c.signContext(ctx), func() {}
		}
	}
	bounded, cancel := context.WithTimeout(ctx, timeout)
	return c.signContext(bounded), cancel
}

func (c *Client) signContext(ctx context.Context) context.Context {
	if len(c.hmacSecret) == 0 {
		return ctx
	}
	return withSignedInternalContext(ctx, c.hmacSecret)
}

func (c *Client) fetchLegalEntity(ctx context.Context, legalEntityID string) (*efv1.LegalEntity, error) {
	rpcCtx, cancel := c.callContext(ctx)
	defer cancel()
	resp, err := c.legalEntities.GetLegalEntityById(rpcCtx, &efv1.GetLegalEntityByIdRequest{Id: legalEntityID})
	if err != nil {
		return nil, mapRPCError(ctx, err)
	}
	if resp == nil {
		return nil, hrerrors.ErrUnavailable
	}
	le := resp.GetLegalEntity()
	if le == nil || !strings.EqualFold(le.GetId(), legalEntityID) || le.GetDeletedAt() != nil || le.GetStatus() != efv1.LegalEntityStatus_LEGAL_ENTITY_STATUS_ACTIVE {
		return nil, hrerrors.ErrReferenceConflict
	}
	return le, nil
}

func validateLegalEntityWindow(le *efv1.LegalEntity, validFrom time.Time, validTo *time.Time) error {
	if le == nil || validFrom.IsZero() ||
		(validTo != nil && (validTo.IsZero() || entity.CompareDates(*validTo, validFrom) < 0)) {
		return hrerrors.ErrReferenceConflict
	}
	if incorporated := le.GetIncorporationDate(); incorporated != nil &&
		(!incorporated.IsValid() || entity.CompareDates(validFrom, incorporated.AsTime()) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	if dissolution := le.GetDissolutionDate(); dissolution != nil {
		if !dissolution.IsValid() {
			return hrerrors.ErrReferenceConflict
		}
		dissolved := dissolution.AsTime()
		if entity.CompareDates(validFrom, dissolved) > 0 || validTo == nil || entity.CompareDates(*validTo, dissolved) > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

func mapRPCError(callerCtx context.Context, err error) error {
	if callerCtx != nil && callerCtx.Err() != nil {
		return callerCtx.Err()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return hrerrors.ErrUnavailable
	}
	st, ok := status.FromError(err)
	if !ok {
		return hrerrors.ErrUnavailable
	}
	switch st.Code() {
	case codes.NotFound, codes.InvalidArgument, codes.FailedPrecondition:
		return hrerrors.ErrReferenceConflict
	case codes.Canceled:
		if callerCtx != nil && callerCtx.Err() != nil {
			return callerCtx.Err()
		}
		return hrerrors.ErrUnavailable
	default:
		return hrerrors.ErrUnavailable
	}
}

func withSignedInternalContext(ctx context.Context, hmacSecret []byte) context.Context {
	caller, ok := authcontext.FromContext(ctx)
	if !ok {
		caller = authcontext.CallerContext{}
	}
	if strings.TrimSpace(caller.RequestID) == "" {
		caller.RequestID = newRequestID()
	}
	return authcontext.AppendToOutgoingGRPCContext(ctx, caller, hmacSecret, time.Now())
}

func newRequestID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("hrm-%d", time.Now().UnixNano())
	}
	return "hrm-" + hex.EncodeToString(raw[:])
}

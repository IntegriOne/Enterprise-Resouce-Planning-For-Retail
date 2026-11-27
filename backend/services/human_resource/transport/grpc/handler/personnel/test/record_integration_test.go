//go:build integration

package test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/personnel"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type recordFoundationStub struct {
	partyErr error
}

func (s *recordFoundationStub) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	return nil
}
func (s *recordFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (s *recordFoundationStub) VerifyLegalEntity(context.Context, string) error { return nil }
func (s *recordFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (s *recordFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (s *recordFoundationStub) VerifyParty(context.Context, string, bool) error {
	return s.partyErr
}

func TestPersonnelRecordAndCollectiveContractRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(db)
	if err != nil {
		t.Fatal(err)
	}
	foundation := &recordFoundationStub{}
	validators, err := security.PersonnelValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	recordValidators, err := security.PersonnelRecordValidators(repos, foundation)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range recordValidators {
		validators[k] = v
	}
	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	personID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	cbaID, _ := ids.NewID()
	docRefID, _ := ids.NewID()
	docVerID, _ := ids.NewID()
	contactPartyID, _ := ids.NewID()
	actorID := "00000000-0000-4000-8000-000000000001"

	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: workerScopePolicy{}, Integrity: security.NewIntegrity(validators),
		},
		Worker: repos.Worker, EmploymentContract: repos.EmploymentContract,
		WorkerAssignment: repos.WorkerAssignment, WorkerLifecycleEvent: repos.WorkerLifecycleEvent,
		WorkerDocument: repos.WorkerDocument, EmergencyContact: repos.EmergencyContact,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: actorID, Authenticated: true}), req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-personnel-record-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewPersonnelServiceClient(conn)

	var docID, contactID, contractID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_employment_contracts WHERE id = ?", contractID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", contractID},
			{"DELETE FROM hr_worker_documents WHERE id = ?", docID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", docID},
			{"DELETE FROM hr_emergency_contacts WHERE id = ?", contactID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", contactID},
			{"DELETE FROM hr_collective_agreements WHERE id = ?", cbaID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if stmt.v == "" {
				continue
			}
			_ = db.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error
		}
	})

	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?)",
		workerID, "REC-W-"+workerID, personID, "employee", "active", legalID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_collective_agreements (id, agreement_code, agreement_version, legal_entity_reference_id, representative_party_reference_id, jurisdiction_code, effective_from, effective_to, signed_at, document_version_reference_id, content_digest) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		cbaID, "CBA-"+cbaID, 1, legalID, contactPartyID, "ID", "2026-01-01", "2027-12-31", time.Now().UTC(), docVerID, []byte("digest"),
	).Error; err != nil {
		t.Fatal(err)
	}

	docType, verStatus, retClass, accClass := "national_id", "pending", "statutory", "restricted"
	docRes, err := client.CreateWorkerDocument(ctx, &pb.CreateWorkerDocumentRequest{Input: &pb.WorkerDocumentInput{
		WorkerId: &workerID, DocumentType: &docType, DocumentReferenceId: &docRefID,
		DocumentVersionReferenceId: &docVerID, ContentDigest: []byte("sha"), VerificationStatus: &verStatus,
		RetentionClass: &retClass, AccessClass: &accClass,
	}})
	if err != nil {
		t.Fatalf("create worker document: %v", err)
	}
	docID = docRes.GetItem().GetId()

	newVersionID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	versionUpdate := &pb.UpdateWorkerDocumentRequest{
		Id: docID, Input: &pb.WorkerDocumentInput{DocumentVersionReferenceId: &newVersionID},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"document_version_reference_id"}},
	}
	if _, err := client.UpdateWorkerDocument(ctx, versionUpdate); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("replace document version without digest = %v, want InvalidArgument", err)
	}
	versionUpdate.Input.ContentDigest = []byte("updated-digest")
	versionUpdate.UpdateMask.Paths = append(versionUpdate.UpdateMask.Paths, "content_digest")
	updatedDocument, err := client.UpdateWorkerDocument(ctx, versionUpdate)
	if err != nil || updatedDocument.GetItem().GetDocumentVersionReferenceId() != newVersionID ||
		string(updatedDocument.GetItem().GetContentDigest()) != "updated-digest" {
		t.Fatalf("replace document version with digest: %+v, %v", updatedDocument, err)
	}

	relCode := "spouse"
	prio := int32(1)
	isActive := true
	if result, err := client.CreateEmergencyContact(ctx, &pb.CreateEmergencyContactRequest{Input: &pb.EmergencyContactInput{
		WorkerId: &workerID, ContactPartyReferenceId: &contactPartyID, RelationshipCode: &relCode,
		PriorityOrder: &prio, AuthorizationConfirmedAt: timestamppb.New(time.Now().Add(time.Hour)), IsActive: &isActive,
	}}); status.Code(err) != codes.InvalidArgument {
		contactID = result.GetItem().GetId()
		t.Fatalf("create contact with future authorization = %v, want InvalidArgument", err)
	}
	contactRes, err := client.CreateEmergencyContact(ctx, &pb.CreateEmergencyContactRequest{Input: &pb.EmergencyContactInput{
		WorkerId: &workerID, ContactPartyReferenceId: &contactPartyID, RelationshipCode: &relCode,
		PriorityOrder: &prio, AuthorizationConfirmedAt: timestamppb.Now(), IsActive: &isActive,
	}})
	if err != nil {
		t.Fatalf("create emergency contact: %v", err)
	}
	contactID = contactRes.GetItem().GetId()
	if _, err := client.CreateEmergencyContact(ctx, &pb.CreateEmergencyContactRequest{Input: &pb.EmergencyContactInput{
		WorkerId: &workerID, ContactPartyReferenceId: &contactPartyID, RelationshipCode: &relCode,
		PriorityOrder: &prio, AuthorizationConfirmedAt: timestamppb.Now(), IsActive: &isActive,
	}}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate emergency contact status = %v, want AlreadyExists", status.Code(err))
	}

	replacementPartyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	replaced, err := client.UpdateEmergencyContact(ctx, &pb.UpdateEmergencyContactRequest{
		Id: contactID, Input: &pb.EmergencyContactInput{ContactPartyReferenceId: &replacementPartyID},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"contact_party_reference_id"}},
	})
	if err != nil || replaced.GetItem().GetContactPartyReferenceId() != replacementPartyID ||
		replaced.GetItem().GetAuthorizationConfirmedAt() != nil {
		t.Fatalf("replace emergency contact without carrying forward old authorization: %+v, %v", replaced, err)
	}

	conNum, conType, conStatus, jurisdiction, country := "CON-CBA-"+workerID, "collective_agreement", "draft", "ID", "ID"
	contractRes, err := client.CreateEmploymentContract(ctx, &pb.CreateEmploymentContractRequest{Input: &pb.EmploymentContractInput{
		WorkerId: &workerID, ContractNumber: &conNum, ContractType: &conType, ContractStatus: &conStatus,
		EmployingLegalEntityReferenceId: &legalID, EffectiveFrom: stringPointer("2026-06-01"),
		WorkCountryCode: &country, GoverningJurisdictionCode: &jurisdiction, CollectiveAgreementId: &cbaID,
	}})
	if err != nil {
		t.Fatalf("create contract linked to collective agreement: %v", err)
	}
	contractID = contractRes.GetItem().GetId()

	if _, err := client.DeleteWorkerDocument(ctx, &common.IDRequest{Id: docID}); err != nil {
		t.Fatalf("delete pending document: %v", err)
	}
	docID = ""
	if _, err := client.DeleteEmergencyContact(ctx, &common.IDRequest{Id: contactID}); err != nil {
		t.Fatalf("delete emergency contact: %v", err)
	}
	contactID = ""
}

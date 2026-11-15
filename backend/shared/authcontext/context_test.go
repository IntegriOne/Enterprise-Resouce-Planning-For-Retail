package authcontext

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"
)

func TestCallerContextRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now()
	secret := []byte("super-secret-internal-gateway-key")
	want := CallerContext{
		PrincipalID:   "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
		SessionID:     "11111111-2222-4333-8444-555555555555",
		UserKind:      "system",
		AuthEpoch:     3,
		PermRevision:  7,
		RequestID:     "req-abc-123",
		ClientIP:      "203.0.113.10",
		UserAgent:     "IntegriOne-Test/1.0",
		Authenticated: true,
	}

	outCtx := AppendToOutgoingGRPCContext(context.Background(), want, secret, now)
	md, ok := metadata.FromOutgoingContext(outCtx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	inCtx := metadata.NewIncomingContext(context.Background(), md)

	got, found, err := ExtractFromIncomingGRPCContext(inCtx, secret, now)
	if err != nil {
		t.Fatalf("ExtractFromIncomingGRPCContext() error = %v", err)
	}
	if !found || got != want {
		t.Fatalf("got (%+v, %v), want (%+v, true)", got, found, want)
	}
}

func TestGatewayProofRejectsTamperingAndExpiry(t *testing.T) {
	t.Parallel()

	now := time.Now()
	secret := []byte("super-secret-internal-gateway-key")
	caller := CallerContext{
		PrincipalID:   "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
		SessionID:     "11111111-2222-4333-8444-555555555555",
		UserKind:      "system",
		AuthEpoch:     1,
		PermRevision:  2,
		Authenticated: true,
	}

	outCtx := AppendToOutgoingGRPCContext(context.Background(), caller, secret, now)
	md, _ := metadata.FromOutgoingContext(outCtx)

	// Uji pemalsuan principal ID tanpa memperbarui HMAC
	tampered := md.Copy()
	tampered.Set(MetaPrincipalID, "99999999-2222-4333-8444-555555555555")
	_, _, err := ExtractFromIncomingGRPCContext(metadata.NewIncomingContext(context.Background(), tampered), secret, now)
	if !errors.Is(err, ErrInvalidGatewayProof) {
		t.Fatalf("expected ErrInvalidGatewayProof on tampered principal, got %v", err)
	}

	// Uji kedaluwarsa melewati batas toleransi replay 60 detik
	_, _, err = ExtractFromIncomingGRPCContext(metadata.NewIncomingContext(context.Background(), md), secret, now.Add(120*time.Second))
	if !errors.Is(err, ErrExpiredGatewayProof) {
		t.Fatalf("expected ErrExpiredGatewayProof after window, got %v", err)
	}

	// Uji bypass proof dengan hanya mengirim metadata non-identitas (mis. x-request-id) tanpa x-ig-gw-proof
	unsignedNonIdentity := metadata.Pairs(
		MetaRequestID, "req-spoof-1",
		MetaClientIP, "198.51.100.2",
	)
	_, _, err = ExtractFromIncomingGRPCContext(metadata.NewIncomingContext(context.Background(), unsignedNonIdentity), secret, now)
	if !errors.Is(err, ErrInvalidGatewayProof) {
		t.Fatalf("expected ErrInvalidGatewayProof on unsigned request metadata, got %v", err)
	}

	// Uji tanda tangan multi-hop X-Forwarded-For sebelum normalisasi hop pertama
	multiHopCaller := CallerContext{
		RequestID: "req-multi-hop",
		ClientIP:  "198.51.100.7, 10.0.0.2",
	}
	multiOut := AppendToOutgoingGRPCContext(context.Background(), multiHopCaller, secret, now)
	multiMD, _ := metadata.FromOutgoingContext(multiOut)
	gotMulti, foundMulti, err := ExtractFromIncomingGRPCContext(metadata.NewIncomingContext(context.Background(), multiMD), secret, now)
	if err != nil {
		t.Fatalf("ExtractFromIncomingGRPCContext(multiHop) error = %v", err)
	}
	if !foundMulti || gotMulti.ClientIP != "198.51.100.7" {
		t.Fatalf("expected normalized first ClientIP 198.51.100.7, got (%+v, %v)", gotMulti, foundMulti)
	}
}

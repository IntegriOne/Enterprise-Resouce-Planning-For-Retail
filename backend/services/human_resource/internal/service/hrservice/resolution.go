package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type ResolutionCheck struct {
	VersionID, RequestID, WorkerID, PrincipalID string
}

// ResolutionVerifier memeriksa bukti DRM immutable yang terikat pada permintaan HR.
type ResolutionVerifier interface {
	VerifyServiceRequestResolution(context.Context, ResolutionCheck) error
}

// ResolveServiceRequest mencatat resolusi dengan dokumen terverifikasi dari PIC.
func (s *Service) ResolveServiceRequest(ctx context.Context, id, versionID string) (*entity.ServiceRequest, error) {
	if s == nil || operation.RequireDependency(s.resolution) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("resolution_document_version_reference_id", versionID); err != nil {
		return nil, err
	}
	return s.serviceRequest.Command(ctx, id, "resolve", func(txCtx context.Context, request *entity.ServiceRequest, actor operation.Actor, now time.Time) (*entity.ServiceRequest, error) {
		if actor.Kind != "human" || request.AssignedPrincipalID == nil || !strings.EqualFold(actor.PrincipalID, *request.AssignedPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if request.Status != "in_progress" || request.ResolvedAt != nil || request.ResolutionDocumentVersionReferenceID != nil ||
			now.Before(request.OpenedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if err := s.resolution.VerifyServiceRequestResolution(txCtx, ResolutionCheck{
			VersionID: versionID, RequestID: request.ID, WorkerID: request.WorkerID, PrincipalID: actor.PrincipalID,
		}); err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		request.Status = "resolved"
		request.ResolvedAt = &now
		request.ResolutionDocumentVersionReferenceID = &versionID
		return request, nil
	})
}

// CloseServiceRequest mencatat konfirmasi pengusul atas permintaan yang sudah diselesaikan.
func (s *Service) CloseServiceRequest(ctx context.Context, id string) (*entity.ServiceRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.Command(ctx, id, "close", func(_ context.Context, request *entity.ServiceRequest, actor operation.Actor, now time.Time) (*entity.ServiceRequest, error) {
		if actor.Kind != "human" || !strings.EqualFold(actor.PrincipalID, request.RequestedByPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if request.Status != "resolved" || request.ResolvedAt == nil || request.ResolvedAt.After(now) ||
			request.ResolutionDocumentVersionReferenceID == nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		request.Status = "closed"
		return request, nil
	})
}

// CancelServiceRequest membatalkan tiket terbuka atas permintaan pengusul.
func (s *Service) CancelServiceRequest(ctx context.Context, id string) (*entity.ServiceRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.Command(ctx, id, "cancel", func(_ context.Context, request *entity.ServiceRequest, actor operation.Actor, now time.Time) (*entity.ServiceRequest, error) {
		if actor.Kind != "human" || !strings.EqualFold(actor.PrincipalID, request.RequestedByPrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if request.Status != "open" || request.AssignedPrincipalID != nil || request.ResolvedAt != nil ||
			request.ResolutionDocumentVersionReferenceID != nil || now.Before(request.OpenedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		request.Status = "cancelled"
		return request, nil
	})
}

package human_resource

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const positionReadPermission = "human_resource.positions.read"

// Position membaca detail posisi berdasarkan ID.
func (a *Adapter) Position(ctx context.Context, rawID string) (*model.Position, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, positionReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "position")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid position ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetPositionByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "position")
	}
	position := response.GetItem()
	if position.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource position response", nil)
	}
	return toModelPosition(position)
}

// Positions mendaftar posisi dengan izin baca tersendiri.
func (a *Adapter) Positions(ctx context.Context, limit, offset int, rawLegalEntityID *string) (*Page[model.Position], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, positionReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid position pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var response *workforce.PositionsResponse
	var legalEntityID string
	if rawLegalEntityID == nil {
		response, err = a.workforce.GetAllPositions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	} else {
		id, parseErr := parseNodeID(*rawLegalEntityID, "legal-entity")
		if parseErr != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid legal entity ID", nil)
		}
		legalEntityID = id.String()
		response, err = a.workforce.FindPositionByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{
			Value: &legalEntityID, Limit: int32(limit), Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, mapReadError(err, "position")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource position page", nil)
	}
	result := &Page[model.Position]{
		Items: make([]*model.Position, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, position := range items {
		if rawLegalEntityID != nil && position.GetLegalEntityReferenceId() != legalEntityID {
			return nil, middleware.ErrUnavailable("invalid human resource position legal entity match", nil)
		}
		item, err := toModelPosition(position)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPosition(row *workforce.Position) (*model.Position, error) {
	for _, raw := range []string{row.GetId(), row.GetJobId(), row.GetOrganizationUnitReferenceId(), row.GetLegalEntityReferenceId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource position reference", nil)
		}
	}
	return &model.Position{
		ID: nodeID("position", row.GetId()), RawID: row.GetId(), PositionCode: row.GetPositionCode(), PositionTitle: row.GetPositionTitle(), JobID: row.GetJobId(),
		OrganizationUnitReferenceID: row.GetOrganizationUnitReferenceId(), LegalEntityReferenceID: row.GetLegalEntityReferenceId(),
		ReportsToPositionID: row.ReportsToPositionId, PositionStatus: row.GetPositionStatus(),
		HeadcountAuthorized: row.GetHeadcountAuthorized(), WorkLocationReferenceID: row.WorkLocationReferenceId,
		ValidFrom: row.GetValidFrom(), ValidTo: row.ValidTo,
	}, nil
}

// ActivatePosition mengaktifkan posisi terencana setelah HRM memeriksa prasyaratnya.
func (a *Adapter) ActivatePosition(ctx context.Context, rawID string) (*model.Position, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.positions.activate")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "position")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid position ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.ActivatePosition(rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("position does not meet activation requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "position")
	}
	position := response.GetItem()
	if position.GetId() != id.String() || position.GetPositionStatus() != "active" {
		return nil, middleware.ErrUnavailable("invalid human resource position response", nil)
	}
	return toModelPosition(position)
}

// CreatePosition mencatat posisi planned dengan referensi yang diverifikasi HRM.
func (a *Adapter) CreatePosition(ctx context.Context, input model.CreatePositionInput) (*model.Position, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.positions.create")
	if err != nil {
		return nil, err
	}
	patch := &workforce.PositionInput{
		PositionCode: &input.PositionCode, PositionTitle: &input.PositionTitle, HeadcountAuthorized: input.HeadcountAuthorized,
		ValidFrom: &input.ValidFrom, ValidTo: input.ValidTo,
	}
	for _, reference := range []struct {
		raw    *string
		target **string
	}{
		{optionalRawNodeID(&input.JobID, "job"), &patch.JobId},
		{optionalRawNodeID(&input.OrganizationUnitReferenceID, "organization-unit"), &patch.OrganizationUnitReferenceId},
		{optionalRawNodeID(&input.LegalEntityReferenceID, "legal-entity"), &patch.LegalEntityReferenceId},
		{optionalRawNodeID(input.ReportsToPositionID, "position"), &patch.ReportsToPositionId},
		{optionalRawNodeID(input.WorkLocationReferenceID, "location"), &patch.WorkLocationReferenceId},
	} {
		if reference.raw == nil {
			continue
		}
		id, err := uuid.Parse(*reference.raw)
		if err != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid position reference", nil)
		}
		value := id.String()
		*reference.target = &value
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.CreatePosition(rpcCtx, &workforce.CreatePositionRequest{Input: patch})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("position conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "position")
	}
	position := response.GetItem()
	if position.GetPositionStatus() != "planned" || position.GetJobId() != patch.GetJobId() ||
		position.GetLegalEntityReferenceId() != patch.GetLegalEntityReferenceId() ||
		position.GetOrganizationUnitReferenceId() != patch.GetOrganizationUnitReferenceId() ||
		position.GetWorkLocationReferenceId() != patch.GetWorkLocationReferenceId() ||
		position.GetReportsToPositionId() != patch.GetReportsToPositionId() {
		return nil, middleware.ErrUnavailable("invalid human resource position response", nil)
	}
	return toModelPosition(position)
}

// UpdatePosition mengoreksi field yang dikirim pada posisi terencana.
func (a *Adapter) UpdatePosition(ctx context.Context, rawID string, input model.UpdatePositionInput) (*model.Position, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.positions.update")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "position")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid position ID", nil)
	}
	patch := &workforce.PositionInput{}
	mask := &fieldmaskpb.FieldMask{}
	for _, field := range []struct {
		path     string
		input    graphql.Omittable[*string]
		target   **string
		nullable bool
	}{
		{"position_title", input.PositionTitle, &patch.PositionTitle, false},
		{"headcount_authorized", input.HeadcountAuthorized, &patch.HeadcountAuthorized, false},
		{"work_location_reference_id", input.WorkLocationReferenceID, &patch.WorkLocationReferenceId, true},
		{"valid_from", input.ValidFrom, &patch.ValidFrom, false},
		{"valid_to", input.ValidTo, &patch.ValidTo, true},
	} {
		value, set := field.input.ValueOK()
		if !set {
			continue
		}
		if value == nil && !field.nullable {
			return nil, middleware.ErrInvalidArgument("required position field cannot be null", nil)
		}
		if field.path == "work_location_reference_id" && value != nil {
			locationID, err := parseNodeID(*value, "location")
			if err != nil || locationID == uuid.Nil {
				return nil, middleware.ErrInvalidArgument("invalid work location ID", nil)
			}
			canonical := locationID.String()
			value = &canonical
		}
		*field.target = value
		mask.Paths = append(mask.Paths, field.path)
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("position update is empty", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.UpdatePosition(rpcCtx, &workforce.UpdatePositionRequest{
		Id: id.String(), Input: patch, UpdateMask: mask,
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("position conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "position")
	}
	position := response.GetItem()
	_, locationChanged := input.WorkLocationReferenceID.ValueOK()
	if position.GetId() != id.String() || position.GetPositionStatus() != "planned" ||
		(locationChanged && position.GetWorkLocationReferenceId() != patch.GetWorkLocationReferenceId()) {
		return nil, middleware.ErrUnavailable("invalid human resource position response", nil)
	}
	return toModelPosition(position)
}

// DeletePosition meneruskan penghapusan posisi terencana ke HRM.
func (a *Adapter) DeletePosition(ctx context.Context, id string) (*bool, error) {
	return a.deleteWorkforceRecord(ctx, id, "human_resource.positions.delete", "position", workforce.WorkforcePlanningServiceClient.DeletePosition)
}

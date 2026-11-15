package iam

import (
	"strings"

	"github.com/google/uuid"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

// GIDPrefix adalah awalan Global Object Identifier kanonik IntegriOne sesuai standar Relay.
const GIDPrefix = "gid://integrione/"

// GIDResource mendefinisikan nama sumber daya entitas pada format `gid://integrione/<resource>/<uuid>`.
type GIDResource string

const (
	ResourceUser         GIDResource = "user"
	ResourceUserIdentity GIDResource = "user-identity"
	ResourceSession      GIDResource = "session"
	ResourceRole         GIDResource = "role"
	ResourcePermission   GIDResource = "permission"
	ResourceGroup        GIDResource = "group"
	ResourceUserRole     GIDResource = "user-role"
	ResourceGroupRole    GIDResource = "group-role"
)

// EncodeGID membentuk Global Object Identifier bertipe `gid://integrione/<resource>/<uuid>`.
func EncodeGID(resource GIDResource, rawID string) string {
	clean := strings.TrimSpace(rawID)
	if clean == "" {
		return ""
	}
	return GIDPrefix + string(resource) + "/" + clean
}

// EncodeOptionalGID membentuk pointer Global Object Identifier apabila nilai mentah tidak kosong.
func EncodeOptionalGID(resource GIDResource, rawID *string) *string {
	if rawID == nil {
		return nil
	}
	encoded := EncodeGID(resource, *rawID)
	if encoded == "" {
		return nil
	}
	return &encoded
}

// DecodeNodeGID memecah Global Object Identifier ketat untuk resolusi polimorfik Query.node.
func DecodeNodeGID(raw string) (GIDResource, string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", "", middleware.ErrInvalidArgument("node id is required", nil)
	}
	res, rawID, err := parseStrictGID(trimmed, true)
	if err != nil {
		return "", "", err
	}
	switch res {
	case ResourceUser, ResourceUserIdentity, ResourceSession, ResourceRole, ResourcePermission, ResourceGroup:
		return res, rawID, nil
	default:
		return "", "", middleware.ErrInvalidArgument("unsupported node GID resource", nil)
	}
}

// ParseTypedID menormalisasi argumen ID GraphQL baik berformat GID maupun UUID mentah menjadi UUID kanonik.
func ParseTypedID(raw string, expected GIDResource) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", middleware.ErrInvalidArgument("id is required", nil)
	}
	if strings.HasPrefix(trimmed, "gid://") {
		res, rawID, err := parseStrictGID(trimmed, false)
		if err != nil {
			return "", err
		}
		if expected != "" && res != expected {
			return "", middleware.ErrInvalidArgument("unexpected GID resource type", nil)
		}
		return rawID, nil
	}
	if !isSafeRawID(trimmed) {
		return "", middleware.ErrInvalidArgument("invalid id format", nil)
	}
	if parsed, err := uuid.Parse(trimmed); err == nil {
		return parsed.String(), nil
	}
	return trimmed, nil
}

// ParseFlexibleAssignmentID menerima UUID mentah maupun GID penugasan peran untuk pencabutan assignment.
func ParseFlexibleAssignmentID(raw string, expectedResources ...GIDResource) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", middleware.ErrInvalidArgument("assignment_id is required", nil)
	}
	if strings.HasPrefix(trimmed, "gid://") {
		res, rawID, err := parseStrictGID(trimmed, false)
		if err != nil {
			return "", err
		}
		if len(expectedResources) > 0 {
			matched := false
			for _, candidate := range expectedResources {
				if res == candidate {
					matched = true
					break
				}
			}
			if !matched {
				return "", middleware.ErrInvalidArgument("unexpected assignment GID resource type", nil)
			}
		}
		return rawID, nil
	}
	if !isSafeRawID(trimmed) {
		return "", middleware.ErrInvalidArgument("invalid assignment_id format", nil)
	}
	if parsed, err := uuid.Parse(trimmed); err == nil {
		return parsed.String(), nil
	}
	return trimmed, nil
}

// ParseOptionalScopeID menormalisasi scope_id opsional baik berupa GID lintas domain maupun UUID mentah.
func ParseOptionalScopeID(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	if strings.HasPrefix(trimmed, "gid://") {
		_, rawID, err := parseStrictGID(trimmed, false)
		if err != nil {
			return nil, err
		}
		return &rawID, nil
	}
	if !isSafeRawID(trimmed) {
		return nil, middleware.ErrInvalidArgument("invalid scope_id format", nil)
	}
	if parsed, err := uuid.Parse(trimmed); err == nil {
		normalized := parsed.String()
		return &normalized, nil
	}
	return &trimmed, nil
}

func parseStrictGID(trimmed string, requireUUID bool) (GIDResource, string, error) {
	if !strings.HasPrefix(trimmed, GIDPrefix) {
		return "", "", middleware.ErrInvalidArgument("invalid global object ID scheme", nil)
	}
	rest := strings.TrimPrefix(trimmed, GIDPrefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		return "", "", middleware.ErrInvalidArgument("invalid global object ID structure", nil)
	}
	resource := GIDResource(parts[0])
	rawID := parts[1]
	if resource == "" || rawID == "" || !isSafeRawID(string(resource)) || !isSafeRawID(rawID) {
		return "", "", middleware.ErrInvalidArgument("invalid global object ID segments", nil)
	}
	parsed, err := uuid.Parse(rawID)
	if err != nil {
		if requireUUID {
			return "", "", middleware.ErrInvalidArgument("global object ID identifier must be a valid UUID", err)
		}
		return resource, rawID, nil
	}
	return resource, parsed.String(), nil
}

func isSafeRawID(val string) bool {
	if val == "" || strings.ContainsAny(val, " \t\r\n/?#\\") {
		return false
	}
	return true
}

package enterprise_foundation

import (
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
)

// GIDPrefix menentukan awalan kanonik Relay Global Object ID IntegriOne.
const GIDPrefix = "gid://integrione/"

// GIDResource mendefinisikan tipe entitas yang dapat dibungkus dalam Relay Global ID Enterprise Foundation.
type GIDResource string

const (
	ResourceAddress          GIDResource = "address"
	ResourceLegalEntity      GIDResource = "legal-entity"
	ResourceLocation         GIDResource = "location"
	ResourceOrganizationUnit GIDResource = "organization-unit"
	ResourceParty            GIDResource = "party"
	ResourcePartyRole        GIDResource = "party-role"
	ResourcePartyContact     GIDResource = "party-contact"
	ResourcePartyAddress     GIDResource = "party-address"
	ResourceIAMUser          GIDResource = "user"
	ResourceWorker           GIDResource = "hrm-worker"
)

// IsEFNodeID melaporkan apakah sebuah ID bertipe Global Object ID milik domain Enterprise Foundation.
func IsEFNodeID(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, GIDPrefix) {
		return false
	}
	rest := strings.TrimPrefix(trimmed, GIDPrefix)
	slashIdx := strings.IndexByte(rest, '/')
	if slashIdx <= 0 {
		return false
	}
	switch GIDResource(rest[:slashIdx]) {
	case ResourceAddress,
		ResourceLegalEntity,
		ResourceLocation,
		ResourceOrganizationUnit,
		ResourceParty,
		ResourcePartyRole,
		ResourcePartyContact,
		ResourcePartyAddress:
		return true
	default:
		return false
	}
}

// EncodeGID membentuk Relay Global Object ID (`gid://integrione/<resource>/<rawID>`).
func EncodeGID(resource GIDResource, rawID string) string {
	trimmed := strings.TrimSpace(rawID)
	if trimmed == "" {
		return ""
	}
	return GIDPrefix + string(resource) + "/" + trimmed
}

// EncodeOptionalGID membentuk pointer GID opsional apabila ID sumber tersedia.
func EncodeOptionalGID(resource GIDResource, rawID string) *string {
	trimmed := strings.TrimSpace(rawID)
	if trimmed == "" {
		return nil
	}
	encoded := EncodeGID(resource, trimmed)
	return &encoded
}

// ParseTypedID memvalidasi dan menormalkan argumen ID menjadi UUID mentah dari GID bertipe yang diharapkan maupun UUID mentah.
func ParseTypedID(raw string, expected GIDResource) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", middleware.ErrInvalidArgument("identifier is required", nil)
	}
	if strings.HasPrefix(trimmed, GIDPrefix) {
		res, idVal, err := parseStrictGID(trimmed, true)
		if err != nil {
			return "", err
		}
		if res != expected {
			return "", middleware.ErrInvalidArgument("unexpected global object ID resource type", nil)
		}
		return idVal, nil
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return "", middleware.ErrInvalidArgument("identifier must be a valid UUID or Global ID", err)
	}
	return parsed.String(), nil
}

// ParseOptionalTypedID memvalidasi ID opsional bertipe tertentu apabila bernilai non-kosong.
func ParseOptionalTypedID(raw *string, expected GIDResource) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := ParseTypedID(trimmed, expected)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// ParseOptionalUUIDOrGID memvalidasi UUID mentah atau GID lintas layanan (seperti iam_user_id atau head_person_id).
func ParseOptionalUUIDOrGID(raw *string, allowedResources ...GIDResource) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := ParseUUIDOrGID(trimmed, allowedResources...)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// ParseUUIDOrGID memvalidasi identifier wajib sebagai UUID mentah atau GID dari daftar resource yang diizinkan.
func ParseUUIDOrGID(raw string, allowedResources ...GIDResource) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", middleware.ErrInvalidArgument("identifier is required", nil)
	}
	if strings.HasPrefix(trimmed, GIDPrefix) {
		res, idVal, err := parseStrictGID(trimmed, true)
		if err != nil {
			return "", err
		}
		if len(allowedResources) > 0 {
			allowed := false
			for _, candidate := range allowedResources {
				if res == candidate {
					allowed = true
					break
				}
			}
			if !allowed {
				return "", middleware.ErrInvalidArgument("unexpected global object ID resource type", nil)
			}
		}
		return idVal, nil
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return "", middleware.ErrInvalidArgument("identifier must be a valid UUID or Global ID", err)
	}
	return parsed.String(), nil
}

// ParseTypedIDList memvalidasi daftar ID menjadi potongan UUID mentah unik atau terurut.
func ParseTypedIDList(ids []string, expected GIDResource) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		cleanID, err := ParseTypedID(raw, expected)
		if err != nil {
			return nil, err
		}
		out = append(out, cleanID)
	}
	return out, nil
}

// DecodeNodeGID memecah Global Object ID Enterprise Foundation untuk Query.node.
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
	case ResourceAddress,
		ResourceLegalEntity,
		ResourceLocation,
		ResourceOrganizationUnit,
		ResourceParty,
		ResourcePartyRole,
		ResourcePartyContact,
		ResourcePartyAddress:
		return res, rawID, nil
	default:
		return "", "", middleware.ErrInvalidArgument("unsupported node GID resource", nil)
	}
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
	if resource == "" || rawID == "" || !isSafeSegment(string(resource)) || !isSafeSegment(rawID) {
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

func isSafeSegment(val string) bool {
	if val == "" || len(val) > 128 {
		return false
	}
	for i := 0; i < len(val); i++ {
		c := val[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			continue
		}
		return false
	}
	return true
}

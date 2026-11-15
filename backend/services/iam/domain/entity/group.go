package entity

import (
	"strings"
	"time"
	"unicode/utf8"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// Group merepresentasikan kelompok pengguna untuk penetapan peran kolektif pada tabel groups.
type Group struct {
	ID          string
	Code        string
	Name        string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewGroup membentuk grup baru dalam keadaan aktif setelah memvalidasi kode dan nama grup.
func NewGroup(id, code, name, description string, now time.Time) (*Group, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	normalizedCode, err := ValidateEntityCode(code, MaxGroupCodeLength, iamerrors.ErrInvalidGroupCode)
	if err != nil {
		return nil, err
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || utf8.RuneCountInString(trimmedName) > MaxGroupNameLength {
		return nil, iamerrors.NewValidation("name", iamerrors.Code("iam.group.invalid_name"), "group name must be non-empty and within length bounds")
	}

	ts := normalizeLocalTime(now)
	return &Group{
		ID:          strings.TrimSpace(id),
		Code:        normalizedCode,
		Name:        trimmedName,
		Description: strings.TrimSpace(description),
		IsActive:    true,
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}, nil
}

// UpdateDetails memperbarui nama tampilan dan deskripsi grup.
func (g *Group) UpdateDetails(name, description string, now time.Time) error {
	if g == nil {
		return iamerrors.ErrGroupNotFound
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || utf8.RuneCountInString(trimmedName) > MaxGroupNameLength {
		return iamerrors.NewValidation("name", iamerrors.Code("iam.group.invalid_name"), "group name must be non-empty and within length bounds")
	}
	g.Name = trimmedName
	g.Description = strings.TrimSpace(description)
	g.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Activate mengaktifkan kembali grup pengguna.
func (g *Group) Activate(now time.Time) {
	if g == nil {
		return
	}
	g.IsActive = true
	g.UpdatedAt = normalizeLocalTime(now)
}

// Deactivate menonaktifkan grup agar pewarisan perannya berhenti berlaku.
func (g *Group) Deactivate(now time.Time) {
	if g == nil {
		return
	}
	g.IsActive = false
	g.UpdatedAt = normalizeLocalTime(now)
}

// GroupMember merepresentasikan keanggotaan pengguna di dalam grup pada tabel group_members.
type GroupMember struct {
	GroupID   string
	UserID    string
	CreatedAt time.Time
}

// NewGroupMember membentuk entitas keanggotaan grup dengan validasi UUID grup dan pengguna.
func NewGroupMember(groupID, userID string, now time.Time) (*GroupMember, error) {
	if err := ValidateUUID("group_id", groupID); err != nil {
		return nil, err
	}
	if err := ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	return &GroupMember{
		GroupID:   strings.TrimSpace(groupID),
		UserID:    strings.TrimSpace(userID),
		CreatedAt: normalizeLocalTime(now),
	}, nil
}

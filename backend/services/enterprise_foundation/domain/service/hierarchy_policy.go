package service

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// DefaultMaxHierarchyDepth membatasi kedalaman maksimum pohon badan hukum, lokasi, dan unit organisasi untuk mencegah siklus atau rantai tak terbatas.
const DefaultMaxHierarchyDepth = 16

// ParentLookupFunc mengambil ID induk dari sebuah node hierarki berdasarkan ID-nya.
type ParentLookupFunc func(ctx context.Context, id string) (*string, error)

// HierarchyPolicy menegakkan konsistensi hierarki pohon internal serta batasan dependensi antar entitas internal.
type HierarchyPolicy struct {
	maxDepth int
}

// NewDefaultHierarchyPolicy membentuk HierarchyPolicy dengan batas kedalaman bawaan DefaultMaxHierarchyDepth.
func NewDefaultHierarchyPolicy() HierarchyPolicy {
	return HierarchyPolicy{maxDepth: DefaultMaxHierarchyDepth}
}

// NewHierarchyPolicy membentuk HierarchyPolicy dengan batas kedalaman khusus yang aman (jatuh ke bawaan bila non-positif).
func NewHierarchyPolicy(maxDepth int) HierarchyPolicy {
	if maxDepth <= 0 {
		maxDepth = DefaultMaxHierarchyDepth
	}
	return HierarchyPolicy{maxDepth: maxDepth}
}

// MaxDepth mengembalikan batas kedalaman hierarki yang dikonfigurasi.
func (p HierarchyPolicy) MaxDepth() int {
	if p.maxDepth <= 0 {
		return DefaultMaxHierarchyDepth
	}
	return p.maxDepth
}

// ValidateParentChain menelusuri rantai leluhur untuk mencegah tautan diri sendiri, siklus multi-lompatan, dan kedalaman pohon berlebih.
func (p HierarchyPolicy) ValidateParentChain(
	ctx context.Context,
	targetID string,
	proposedParentID *string,
	lookup ParentLookupFunc,
	selfErr error,
	cycleErr error,
) error {
	if proposedParentID == nil {
		return nil
	}
	if err := entity.ValidateUUID("id", targetID); err != nil {
		return err
	}
	if err := entity.ValidateUUID("parent_id", *proposedParentID); err != nil {
		return err
	}

	normTarget := entity.NormalizeUUID(targetID)
	curParent := entity.NormalizeUUID(*proposedParentID)
	if curParent == normTarget {
		return selfErr
	}
	if lookup == nil {
		return eferrors.WithField(eferrors.ErrInvalidInput, "parent_lookup", "parent lookup callback is required")
	}

	visited := map[string]struct{}{
		normTarget: {},
	}
	maxDepth := p.MaxDepth()
	for depth := 1; depth <= maxDepth; depth++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, seen := visited[curParent]; seen {
			return cycleErr
		}
		visited[curParent] = struct{}{}

		nextParent, err := lookup(ctx, curParent)
		if err != nil {
			return err
		}
		if nextParent == nil {
			return nil
		}
		trimmedNext := entity.NormalizeUUID(*nextParent)
		if trimmedNext == "" {
			return nil
		}
		if !entity.IsValidUUID(trimmedNext) {
			return cycleErr
		}
		curParent = trimmedNext
	}

	return cycleErr
}

// EnsureSameLegalEntityForLocation memverifikasi bahwa lokasi induk ada, tidak dihapus lunak, dan berada di bawah badan hukum yang sama.
func (HierarchyPolicy) EnsureSameLegalEntityForLocation(entityID string, parent *entity.Location) error {
	if parent == nil {
		return nil
	}
	if parent.IsDeleted() {
		return eferrors.ErrLocationDeleted
	}
	if !strings.EqualFold(entity.NormalizeUUID(entityID), entity.NormalizeUUID(parent.LegalEntityID)) {
		return eferrors.ErrLocationParentMismatch
	}
	return nil
}

// EnsureSameLegalEntityForOrgUnit memverifikasi unit induk dan lokasi utama opsional berada pada badan hukum yang sama dengan unit anak.
func (HierarchyPolicy) EnsureSameLegalEntityForOrgUnit(
	entityID string,
	parentUnit *entity.OrganizationUnit,
	primaryLocation *entity.Location,
) error {
	normEntity := entity.NormalizeUUID(entityID)
	if parentUnit != nil {
		if parentUnit.IsDeleted() {
			return eferrors.ErrOrgUnitDeleted
		}
		if !strings.EqualFold(normEntity, entity.NormalizeUUID(parentUnit.LegalEntityID)) {
			return eferrors.ErrOrgUnitParentMismatch
		}
	}
	if primaryLocation != nil {
		if primaryLocation.IsDeleted() {
			return eferrors.ErrLocationDeleted
		}
		if !strings.EqualFold(normEntity, entity.NormalizeUUID(primaryLocation.LegalEntityID)) {
			return eferrors.ErrOrgUnitLocationMismatch
		}
	}
	return nil
}

// ValidateLegalEntityRemoval memvalidasi bahwa badan hukum boleh dihapus lunak tanpa meninggalkan entitas utama atau tanggungan aktif.
func (HierarchyPolicy) ValidateLegalEntityRemoval(le *entity.LegalEntity, activeDependentsCount int64) error {
	if le == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if le.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if le.IsPrimaryOperatingEntity {
		return eferrors.ErrPrimaryLegalEntityExists
	}
	if activeDependentsCount > 0 {
		return eferrors.ErrLegalEntityHasDependents
	}
	return nil
}

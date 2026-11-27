package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

// PartyRepository mendefinisikan operasi persistensi murni untuk tabel parties.
type PartyRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.Party, error)
	SelectByIDs(ctx context.Context, ids []string) ([]*entity.Party, error)
	SelectByCode(ctx context.Context, partnerCode string) (*entity.Party, error)
	SelectByTaxIDNumber(ctx context.Context, taxIDNumber string) (*entity.Party, error)
	SelectByEmail(ctx context.Context, email string, page PageRequest) (PageResult[*entity.Party], error)
	SelectList(ctx context.Context, filter PartyFilter) (PageResult[*entity.Party], error)
	CheckCodeExists(ctx context.Context, partnerCode, excludeID string) (bool, error)
	CheckTaxIDExists(ctx context.Context, taxIDNumber, excludeID string) (bool, error)
	Insert(ctx context.Context, party *entity.Party) error
	Update(ctx context.Context, party *entity.Party) error
	SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error
}

// PartyRoleRepository mendefinisikan operasi persistensi murni untuk tabel party_roles.
type PartyRoleRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.PartyRole, error)
	SelectByPartyID(ctx context.Context, partyID string, page PageRequest) (PageResult[*entity.PartyRole], error)
	SelectByPartyAndRole(ctx context.Context, partyID string, role entity.PartyRoleKind) (*entity.PartyRole, error)
	SelectByRoleKind(ctx context.Context, role entity.PartyRoleKind, activeOnly bool, page PageRequest) (PageResult[*entity.PartyRole], error)
	SelectList(ctx context.Context, filter PartyRoleFilter) (PageResult[*entity.PartyRole], error)
	CheckExists(ctx context.Context, partyID string, role entity.PartyRoleKind) (bool, error)
	Insert(ctx context.Context, partyRole *entity.PartyRole) error
	Update(ctx context.Context, partyRole *entity.PartyRole) error
	DeleteByID(ctx context.Context, id string) error
}

// PartyContactRepository mendefinisikan operasi persistensi murni untuk tabel party_contacts dan tautan iam_user_id.
type PartyContactRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.PartyContact, error)
	SelectByPartyID(ctx context.Context, partyID string, page PageRequest) (PageResult[*entity.PartyContact], error)
	SelectPrimaryByPartyID(ctx context.Context, partyID string) (*entity.PartyContact, error)
	SelectByIAMUserID(ctx context.Context, iamUserID string, page PageRequest) (PageResult[*entity.PartyContact], error)
	SelectByPartyAndIAMUserID(ctx context.Context, partyID, iamUserID string) (*entity.PartyContact, error)
	SelectByEmail(ctx context.Context, email string, page PageRequest) (PageResult[*entity.PartyContact], error)
	SelectList(ctx context.Context, filter PartyContactFilter) (PageResult[*entity.PartyContact], error)
	CheckIAMUserExistsInParty(ctx context.Context, partyID, iamUserID, excludeID string) (bool, error)
	UpdateClearPrimaryByPartyExcept(ctx context.Context, partyID, keepID string, updatedAt time.Time) error
	Insert(ctx context.Context, contact *entity.PartyContact) error
	Update(ctx context.Context, contact *entity.PartyContact) error
	SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error
}

// PartyAddressRepository mendefinisikan operasi persistensi murni untuk tabel party_addresses.
type PartyAddressRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.PartyAddress, error)
	SelectByPartyID(ctx context.Context, partyID string, page PageRequest) (PageResult[*entity.PartyAddress], error)
	SelectByPartyAndUsage(ctx context.Context, partyID string, usage entity.AddressUsageKind, page PageRequest) (PageResult[*entity.PartyAddress], error)
	SelectPrimaryByPartyAndUsage(ctx context.Context, partyID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error)
	SelectByAddressID(ctx context.Context, addressID string, page PageRequest) (PageResult[*entity.PartyAddress], error)
	SelectByPartyAddressUsage(ctx context.Context, partyID, addressID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error)
	SelectList(ctx context.Context, filter PartyAddressFilter) (PageResult[*entity.PartyAddress], error)
	UpdateClearPrimaryByPartyUsageExcept(ctx context.Context, partyID string, usage entity.AddressUsageKind, keepID string, updatedAt time.Time) error
	Insert(ctx context.Context, link *entity.PartyAddress) error
	Update(ctx context.Context, link *entity.PartyAddress) error
	DeleteByID(ctx context.Context, id string) error
}

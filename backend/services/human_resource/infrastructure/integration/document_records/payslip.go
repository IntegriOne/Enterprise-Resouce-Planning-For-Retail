// Package documentrecords mengintegrasikan dokumen payroll HRM dengan DRM.
package documentrecords

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/integrione/erp-retail/backend/services/human_resource/contract"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
)

// PublicationVerifier memeriksa konfirmasi penerbitan dari Finance.
type PublicationVerifier interface {
	VerifyPayrollPublication(context.Context, payroll.PayslipPublicationCheck) (time.Time, error)
}

// PayslipStorage menerima PDF privat dan mengembalikan ID versi immutable DRM.
type PayslipStorage interface {
	StorePrivatePayslip(context.Context, PrivatePayslip) (string, error)
}

// PayslipStorageRestorer menulis ulang blob hanya bila digest cocok dengan metadata versi immutable di DRM.
type PayslipStorageRestorer interface {
	RestorePrivatePayslip(context.Context, string, PrivatePayslip) error
}

type PrivatePayslip struct {
	IdempotencyKey, WorkerID, ResultID, PublicationID string
	LegalEntityID                                     string
	PDF                                               []byte
	Digest                                            [sha256.Size]byte
}

type PayslipProcessor struct {
	PublicationVerifier
	storage PayslipStorage
	font    []byte
}

var _ payroll.PayslipPublicationProcessor = (*PayslipProcessor)(nil)

func NewPayslipProcessor(verifier PublicationVerifier, storage PayslipStorage, font []byte) (*PayslipProcessor, error) {
	if operation.RequireDependency(verifier) != nil || operation.RequireDependency(storage) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if len(font) == 0 {
		return nil, hrerrors.Invalid("font", "TrueType font is required")
	}
	return &PayslipProcessor{PublicationVerifier: verifier, storage: storage, font: bytes.Clone(font)}, nil
}

func (p *PayslipProcessor) StorePayslipPDF(ctx context.Context, input payroll.PayslipDocumentInput) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	pdf, err := p.render(input)
	if err != nil {
		return "", err
	}
	if len(pdf) > contract.MaxPayslipPDFBytes {
		return "", hrerrors.Invalid("pdf", "payslip PDF exceeds the supported size")
	}
	document := PrivatePayslip{
		IdempotencyKey: input.IdempotencyKey, WorkerID: input.Result.WorkerID,
		ResultID: input.Result.ID, PublicationID: input.PublicationID,
		LegalEntityID: input.Period.EmployingLegalEntityReferenceID,
		PDF:           pdf, Digest: sha256.Sum256(pdf),
	}
	if input.RestoreVersionID == "" {
		return p.storage.StorePrivatePayslip(ctx, document)
	}
	restorer, ok := p.storage.(PayslipStorageRestorer)
	if !ok || operation.RequireDependency(restorer) != nil {
		return "", hrerrors.ErrUnavailable
	}
	if err := restorer.RestorePrivatePayslip(ctx, input.RestoreVersionID, document); err != nil {
		return "", err
	}
	return input.RestoreVersionID, nil
}

func (p *PayslipProcessor) render(input payroll.PayslipDocumentInput) ([]byte, error) {
	if input.PublishedAt.IsZero() {
		return nil, hrerrors.Invalid("published_at", "publication timestamp is required")
	}
	if strings.TrimSpace(input.WorkerNumber) == "" {
		return nil, hrerrors.Invalid("worker_number", "worker number is required")
	}
	document := fpdf.New("P", "mm", "A4", "")
	document.SetMargins(18, 16, 18)
	document.SetAutoPageBreak(true, 16)
	document.SetCatalogSort(true)
	document.SetCreationDate(input.PublishedAt.UTC())
	document.SetModificationDate(input.PublishedAt.UTC())
	document.SetTitle("Slip gaji", true)
	document.AddUTF8FontFromBytes("payslip", "", p.font)
	document.AddPage()
	document.SetFont("payslip", "", 17)
	document.MultiCell(0, 10, "Slip gaji", "", "L", false)
	document.SetFontSize(10)
	result, period := input.Result, input.Period
	periodRange := period.PeriodStart.Format(time.DateOnly) + " hingga " + period.PeriodEnd.Format(time.DateOnly)
	if code := strings.Join(strings.Fields(period.PeriodCode), " "); code != "" {
		periodRange = fmt.Sprintf("%s (%s)", code, periodRange)
	}
	for _, row := range []string{
		"Periode " + periodRange,
		"Nomor pekerja " + input.WorkerNumber,
		"Diterbitkan " + input.PublishedAt.UTC().Format(time.DateOnly),
		"Mata uang " + result.CurrencyCode,
	} {
		document.MultiCell(0, 6, row, "", "L", false)
	}
	document.Ln(4)
	document.MultiCell(0, 7, "Rincian penghasilan dan potongan", "B", "L", false)
	lines := slices.Clone(input.Lines)
	slices.SortStableFunc(lines, func(a, b entity.PayrollResultLine) int {
		if a.LineSequence == b.LineSequence {
			return strings.Compare(a.ID, b.ID)
		}
		if a.LineSequence < b.LineSequence {
			return -1
		}
		return 1
	})
	for _, line := range lines {
		if line.ComponentTypeSnapshot == "employer_contribution" {
			continue
		}
		label := strings.Join(strings.Fields(line.ComponentNameSnapshot), " ")
		document.MultiCell(0, 6, fmt.Sprintf("%s   %s %s", label, result.CurrencyCode, line.Amount), "", "L", false)
	}
	document.Ln(4)
	for _, total := range []struct {
		label  string
		amount entity.Decimal
	}{
		{"Penghasilan bruto", result.GrossAmount},
		{"Penggantian biaya", result.ReimbursementAmount},
		{"Total potongan", result.TotalDeductions},
		{"Gaji bersih", result.NetPayAmount},
	} {
		document.MultiCell(0, 7, fmt.Sprintf("%s   %s %s", total.label, result.CurrencyCode, total.amount), "", "L", false)
	}
	var buffer bytes.Buffer
	if err := document.Output(&buffer); err != nil {
		return nil, fmt.Errorf("render payslip: %w", err)
	}
	return buffer.Bytes(), nil
}

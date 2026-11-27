package documentrecords

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
)

func TestPayslipRendersReadableDeterministicPDF(t *testing.T) {
	fontPath := os.Getenv("HRM_PAYSLIP_TEST_FONT")
	if fontPath == "" {
		t.Skip("HRM_PAYSLIP_TEST_FONT is not configured")
	}
	font, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	processor := &PayslipProcessor{font: font}
	start := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	input := payroll.PayslipDocumentInput{
		PublicationID: "internal-publication", WorkerNumber: "TEST-EMPLOYEE", PublishedAt: start.AddDate(0, 0, 24),
		Result: entity.PayrollResult{
			WorkerID: "internal-worker", CurrencyCode: "IDR", GrossAmount: "200000",
			ReimbursementAmount: "0", TotalDeductions: "10000", NetPayAmount: "190000",
		},
		Period: entity.PayrollPeriod{PeriodCode: "NOV-2026", PeriodStart: start, PeriodEnd: start.AddDate(0, 0, 29), EmployingLegalEntityReferenceID: "internal-entity"},
		Lines: []entity.PayrollResultLine{
			{ComponentNameSnapshot: "Tunjangan café", Amount: "200000", LineSequence: 1},
			{ComponentNameSnapshot: "Potongan", Amount: "10000", LineSequence: 2},
			{ComponentNameSnapshot: "Kontribusi perusahaan", ComponentTypeSnapshot: "employer_contribution", Amount: "32100", LineSequence: 3},
		},
	}
	first, err := processor.render(input)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(input.Lines)
	second, err := processor.render(input)
	if err != nil || !bytes.Equal(first, second) || !bytes.HasPrefix(first, []byte("%PDF-")) {
		t.Fatalf("unexpected PDF output: stable=%t, err=%v", bytes.Equal(first, second), err)
	}
	pdfPath := filepath.Join(t.TempDir(), "fixture.pdf")
	if err := os.WriteFile(pdfPath, first, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.CommandContext(t.Context(), "pdftotext", pdfPath, "-").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Periode NOV-2026 (2026-11-01 hingga 2026-11-30)", "Nomor pekerja TEST-EMPLOYEE", "Tunjangan café", "Potongan IDR 10000", "Gaji bersih", "190000", "Diterbitkan 2026-11-25"} {
		if !strings.Contains(string(output), expected) {
			t.Errorf("PDF text is missing %q", expected)
		}
	}
	if strings.Contains(string(output), "internal-") {
		t.Fatal("employee PDF disclosed an internal reference")
	}
	if strings.Contains(string(output), "Kontribusi perusahaan") || strings.Contains(string(output), "32100") {
		t.Fatal("employee PDF disclosed an employer-only contribution")
	}
}

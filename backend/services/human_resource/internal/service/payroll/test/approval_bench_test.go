package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func BenchmarkServiceApprovePayrollRun(b *testing.B) {
	h, svc, _, runs, results, lines := setupPayroll(b)
	run, result := calculatedRun(), fixture.PayrollResult()
	results.Rows[result.ID] = result
	for _, line := range calculatedLines() {
		lines.Rows[line.ID] = line
	}
	b.ReportAllocs()
	for b.Loop() {
		// Reset fixture dikeluarkan dari pengukuran agar jumlah record selalu tetap.
		b.StopTimer()
		runs.Rows[run.ID] = run
		h.Events = nil
		b.StartTimer()
		if _, err := svc.ApprovePayrollRun(b.Context(), run.ID); err != nil {
			b.Fatal(err)
		}
	}
}

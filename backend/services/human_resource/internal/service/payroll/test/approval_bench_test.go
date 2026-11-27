package service_test

import (
	"testing"
)

func BenchmarkServiceApprovePayrollRun(b *testing.B) {
	h, svc, _, runs, results, lines := setupPayroll(b, 6)
	run, result := calculatedRun(), calculatedResult()
	run.Status = "review"
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

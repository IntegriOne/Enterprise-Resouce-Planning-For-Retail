// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"context"
	"testing"
)

func BenchmarkServiceList(b *testing.B) {
	b.Run("scoped-page-1", func(b *testing.B) {
		_, svc, _ := setup(b)
		ctx := context.Background()
		if _, err := svc.CreateDevelopmentPlan(ctx, input()); err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		for b.Loop() {
			page, err := svc.GetAllDevelopmentPlans(ctx, 20, 0)
			if err != nil || len(page.Items) != 1 {
				b.Fatalf("page: %v", err)
			}
		}
	})
}

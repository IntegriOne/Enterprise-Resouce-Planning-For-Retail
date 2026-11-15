// Package main menjalankan generator gqlgen dengan plugin pemecah model Hybrid A1.
package main

import (
	"fmt"
	"os"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/integrione/erp-retail/backend/gateway/internal/gqltools"
)

func main() {
	cfg, err := config.LoadConfigFromDefaultLocations()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gqlgen config: %v\n", err)
		os.Exit(1)
	}

	if err := api.Generate(cfg, gqltools.UseSplitModelGen()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate graphql artifacts: %v\n", err)
		os.Exit(1)
	}
}

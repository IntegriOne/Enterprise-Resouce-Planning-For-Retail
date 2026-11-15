module github.com/integrione/erp-retail/backend/shared/authsession

go 1.27.4

require (
	github.com/integrione/erp-retail/backend/shared/authz v0.0.0
	github.com/redis/go-redis/v9 v9.22.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

replace github.com/integrione/erp-retail/backend/shared/authz => ../authz

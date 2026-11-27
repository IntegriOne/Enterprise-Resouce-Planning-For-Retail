module github.com/integrione/erp-retail/backend/services/human_resource

go 1.27.4

require (
	github.com/integrione/erp-retail/backend/services/iam v0.0.0
	github.com/integrione/erp-retail/backend/shared/authcontext v0.0.0
	github.com/integrione/erp-retail/backend/shared/authsession v0.0.0
	github.com/integrione/erp-retail/backend/shared/authz v0.0.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261110000943-09e21b237f2c
	google.golang.org/grpc v1.85.0
	google.golang.org/protobuf v1.36.12
	gorm.io/driver/postgres v1.6.3
	gorm.io/gorm v1.31.2
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.13.0 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/redis/go-redis/v9 v9.22.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace github.com/integrione/erp-retail/backend/shared/authcontext => ../../shared/authcontext

replace github.com/integrione/erp-retail/backend/shared/authsession => ../../shared/authsession

replace github.com/integrione/erp-retail/backend/shared/authz => ../../shared/authz

replace github.com/integrione/erp-retail/backend/services/iam => ../iam

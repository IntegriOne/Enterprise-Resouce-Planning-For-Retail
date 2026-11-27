module github.com/integrione/erp-retail/backend/services/iam

go 1.27.4

require (
	github.com/aws/aws-sdk-go-v2 v1.48.0
	github.com/aws/aws-sdk-go-v2/config v1.33.18
	github.com/aws/aws-sdk-go-v2/credentials v1.21.0
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.79.9
	github.com/aws/smithy-go v1.28.9
	github.com/integrione/erp-retail/backend/services/human_resource v0.0.0
	github.com/integrione/erp-retail/backend/shared/authcontext v0.0.0
	github.com/integrione/erp-retail/backend/shared/authsession v0.0.0
	github.com/integrione/erp-retail/backend/shared/authz v0.0.0
	github.com/jackc/pgx/v5 v5.13.0
	github.com/redis/go-redis/v9 v9.22.0
	golang.org/x/crypto v0.58.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261110000943-09e21b237f2c
	google.golang.org/grpc v1.85.0
	google.golang.org/protobuf v1.36.12
	gorm.io/driver/postgres v1.6.3
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.2
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.20.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.39.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.44.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.52.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/integrione/erp-retail/backend/services/enterprise_foundation v0.0.0
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace (
	github.com/integrione/erp-retail/backend/shared/authcontext => ../../shared/authcontext
	github.com/integrione/erp-retail/backend/shared/authsession => ../../shared/authsession
	github.com/integrione/erp-retail/backend/shared/authz => ../../shared/authz
)

replace github.com/integrione/erp-retail/backend/services/human_resource => ../human_resource

replace github.com/integrione/erp-retail/backend/services/enterprise_foundation => ../enterprise_foundation

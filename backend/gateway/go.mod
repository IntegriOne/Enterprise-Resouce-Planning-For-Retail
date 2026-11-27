module github.com/integrione/erp-retail/backend/gateway

go 1.27.4

require (
	github.com/99designs/gqlgen v0.17.97
	github.com/gofiber/fiber/v3 v3.7.0
	github.com/google/uuid v1.6.0
	github.com/integrione/erp-retail/backend/services/enterprise_foundation v0.0.0
	github.com/integrione/erp-retail/backend/services/human_resource v0.0.0
	github.com/integrione/erp-retail/backend/services/iam v0.0.0
	github.com/integrione/erp-retail/backend/shared/authcontext v0.0.0
	github.com/integrione/erp-retail/backend/shared/authsession v0.0.0
	github.com/integrione/erp-retail/backend/shared/authz v0.0.0
	github.com/integrione/erp-retail/backend/shared/rbac v0.0.0
	github.com/redis/go-redis/v9 v9.22.0
	github.com/vektah/gqlparser/v2 v2.5.63
	golang.org/x/sync v0.25.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261110000943-09e21b237f2c
	google.golang.org/grpc v1.85.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.15 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/gofiber/schema v1.9.3 // indirect
	github.com/gofiber/utils/v2 v2.6.4 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/molecule-man/go-brrr v1.3.1 // indirect
	github.com/philhofer/fwd v1.2.1 // indirect
	github.com/sosodev/duration v1.4.0 // indirect
	github.com/tinylib/msgp v1.6.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.78.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	gorm.io/gorm v1.31.2 // indirect
)

replace (
	github.com/integrione/erp-retail/backend/services/enterprise_foundation => ../services/enterprise_foundation
	github.com/integrione/erp-retail/backend/services/human_resource => ../services/human_resource
	github.com/integrione/erp-retail/backend/services/iam => ../services/iam
	github.com/integrione/erp-retail/backend/shared/authcontext => ../shared/authcontext
	github.com/integrione/erp-retail/backend/shared/authsession => ../shared/authsession
	github.com/integrione/erp-retail/backend/shared/authz => ../shared/authz
)

replace github.com/integrione/erp-retail/backend/shared/rbac => ../shared/rbac

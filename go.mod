module github.com/dolthub/go-mysql-server

go 1.22.2

// uses protobuf-go-lite
replace github.com/dolthub/vitess => github.com/aperturerobotics/vitess v0.0.0-20240821040752-39ac045ae8fe // aperture

require (
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/dolthub/flatbuffers/v23 v23.3.3-dh.2
	github.com/dolthub/go-icu-regex v0.0.0-20230524105445-af7e7991c97e
	github.com/dolthub/jsonpath v0.0.2-0.20240227200619-19675ab05c71
	github.com/dolthub/vitess v0.0.0-20240429213844-e8e1b4cd75c4
	github.com/go-kit/kit v0.10.0
	github.com/go-sql-driver/mysql v1.7.2-0.20231213112541-0004702b931d
	github.com/gocraft/dbr/v2 v2.7.2
	github.com/google/uuid v1.3.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/lestrrat-go/strftime v1.0.4
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
	golang.org/x/sync v0.3.0
	golang.org/x/sys v0.18.0
	golang.org/x/text v0.6.0
	golang.org/x/tools v0.13.0
	gopkg.in/src-d/go-errors.v1 v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/aperturerobotics/json-iterator-lite v1.0.0 // indirect
	github.com/aperturerobotics/protobuf-go-lite v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/tetratelabs/wazero v1.1.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
)

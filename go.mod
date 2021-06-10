module github.com/dolthub/go-mysql-server

go 1.15

replace github.com/dolthub/vitess => github.com/paralin/vitess v0.0.0-20210611010940-f1489325f50b // ext-engines

replace github.com/oliveagle/jsonpath => github.com/dolthub/jsonpath v0.0.0-20210609232853-d49537a30474

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/cespare/xxhash v1.1.0
	github.com/dolthub/sqllogictest/go v0.0.0-20201107003712-816f3ae12d81
	github.com/dolthub/vitess v0.0.0-20210610232639-3424dd4d93a1
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/go-kit/kit v0.9.0
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529
	github.com/google/go-cmp v0.3.0 // indirect
	github.com/google/uuid v1.2.0
	github.com/hashicorp/golang-lru v0.5.3
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lestrrat-go/strftime v1.0.1
	github.com/mitchellh/hashstructure v1.0.0
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/shopspring/decimal v1.2.1-0.20210329231237-501661573f60
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cast v1.3.0
	github.com/src-d/go-oniguruma v1.1.0
	github.com/stretchr/testify v1.7.0
	github.com/tebeka/strftime v0.1.4 // indirect
	golang.org/x/net v0.0.0-20210610132358-84b48f89b13b // indirect
	golang.org/x/sys v0.0.0-20210608053332-aa57babbf139 // indirect
	google.golang.org/grpc v1.27.0 // indirect
	gopkg.in/src-d/go-errors.v1 v1.0.0
)

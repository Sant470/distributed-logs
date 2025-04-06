module github.com/sant470/distlogs

go 1.23.0

toolchain go1.23.3

replace cloud.google.com/go => cloud.google.com/go v0.100.2 // or any newer version that no longer includes compute/metadata

require (
	github.com/casbin/casbin/v2 v2.104.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/stretchr/testify v1.10.0
	github.com/tysonmote/gommap v0.0.3
	go.opencensus.io v0.24.0
	go.uber.org/zap v1.27.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250404141209-ee84b53bf3d0
	google.golang.org/grpc v1.71.1
	google.golang.org/protobuf v1.36.6
)

require (
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

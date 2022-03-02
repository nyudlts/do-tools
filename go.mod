module github.com/dmnyu/do-tools

go 1.17

require (
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/lestrrat-go/libxml2 v0.0.0-20201123224832-e6d9de61b80d // indirect
	github.com/nyudlts/go-aspace v0.3.8-0.20220214182818-e92697d431b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/nyudlts/do-tools/cmd => /usr/local/go/src/github.com/nyudlts/do-tools/cmd
)

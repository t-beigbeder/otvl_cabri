module github.com/t-beigbeder/otvl_cabri/gocode

go 1.21

require (
	filippo.io/age v1.1.1
	github.com/aws/aws-sdk-go v1.53.21
	github.com/google/uuid v1.6.0
	github.com/labstack/echo/v4 v4.12.0
	github.com/muesli/coral v1.0.0
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/spf13/afero v1.11.0
	github.com/tidwall/buntdb v1.3.1
	golang.org/x/crypto v0.24.0
	golang.org/x/sys v0.21.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	github.com/tidwall/gjson v1.17.1 // indirect
	github.com/tidwall/grect v0.1.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/rtred v0.1.2 // indirect
	github.com/tidwall/tinyqueue v0.1.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/term v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
)

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss => ./packages/cabridss

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrisync => ./packages/cabrisync

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrifsu => ./packages/cabrifsu

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/cabritbx => ./packages/cabritbx

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui => ./packages/cabriui

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/em4ht => ./packages/em4ht

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/internal => ./packages/internal

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/joule => ./packages/joule

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/mockfs => ./packages/mockfs

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber => ./packages/plumber

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/testfs => ./packages/testfs

replace github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath => ./packages/ufpath

replace github.com/t-beigbeder/otvl_cabri/gocode/cabri/cmd => ./cabri/cmd

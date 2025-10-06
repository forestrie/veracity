module github.com/datatrails/veracity

go 1.24.0

replace (
	github.com/datatrails/go-datatrails-common => ../go-datatrails-common
	github.com/forestrie/go-merklelog-azure => ../go-merklelog-azure
	github.com/forestrie/go-merklelog-datatrails => ../go-merklelog-datatrails
	github.com/forestrie/go-merklelog-fs => ../go-merklelog-fs
	github.com/forestrie/go-merklelog-provider-testing => ../go-merklelog-provider-testing
	github.com/forestrie/go-merklelog/massifs => ../go-merklelog/massifs
	github.com/forestrie/go-merklelog/mmr => ../go-merklelog/mmr
)

require (
	github.com/datatrails/go-datatrails-common v0.30.0
	github.com/datatrails/go-datatrails-common-api-gen v0.8.0
	github.com/datatrails/go-datatrails-serialization/eventsv1 v0.0.3
	github.com/datatrails/go-datatrails-simplehash v0.2.0
	github.com/forestrie/go-merklelog-azure v0.0.0-20250928182018-06ed158d48af
	github.com/forestrie/go-merklelog-datatrails v0.0.0
	github.com/forestrie/go-merklelog-fs v0.0.0-20250928180927-a4773e335b22
	github.com/forestrie/go-merklelog-provider-testing v0.0.0-00010101000000-000000000000
	github.com/forestrie/go-merklelog/massifs v0.0.2
	github.com/forestrie/go-merklelog/mmr v0.4.0
	github.com/fxamacker/cbor/v2 v2.9.0
	github.com/google/uuid v1.6.0
	github.com/gosuri/uiprogress v0.0.1
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v2 v2.27.7
	github.com/veraison/go-cose v1.3.0
	github.com/zeebo/bencode v1.0.0
	go.uber.org/zap v1.27.0
	golang.org/x/exp v0.0.0-20250911091902-df9299821621
	google.golang.org/protobuf v1.36.9
)

require (
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.17.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.4.1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.24 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.13 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.3.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/gosuri/uilive v0.0.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.23.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/ldclabs/cose/go v0.0.0-20221214142927-d22c1cfc2154 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

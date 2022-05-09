module github.com/terraform-providers/terraform-provider-aws/tools

go 1.15

require (
	github.com/apparentlymart/go-dump v0.0.0-20190214190832-042adf3cf4a0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/bflad/tfproviderdocs v0.8.0
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.35.2
	github.com/hashicorp/go-changelog v0.0.0-20201005170154-56335215ce3a
	github.com/katbyte/terrafmt v0.2.1-0.20200913185704-5ff4421407b4
	github.com/mitchellh/cli v1.1.1 // indirect
	github.com/posener/complete v1.2.1 // indirect
	github.com/terraform-linters/tflint v0.36.2
	github.com/vmihailenco/msgpack v4.0.1+incompatible // indirect
)

replace github.com/katbyte/terrafmt => github.com/gdavison/terrafmt v0.2.1-0.20201026181004-a896893cd6af

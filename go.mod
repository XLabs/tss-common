module github.com/xlabs/tss-lib/v2

go 1.21

toolchain go1.22.5

require google.golang.org/protobuf v1.33.0

require github.com/google/go-cmp v0.5.9 // indirect

replace github.com/agl/ed25519 => github.com/binance-chain/edwards25519 v0.0.0-20200305024217-f36fc4b53d43

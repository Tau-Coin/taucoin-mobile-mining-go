module github.com/Tau-Coin/taucoin-mobile-mining-go

go 1.14

require (
	github.com/Azure/azure-storage-blob-go v0.8.0
	github.com/Tau-Coin/taucoin-go-p2p v0.0.0-20200315064759-d0a404939e87
	github.com/allegro/bigcache v1.2.1
	github.com/apilayer/freegeoip v3.5.0+incompatible
	github.com/aristanetworks/goarista v0.0.0-20200206021550-59c4040ef2d3
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.1
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/docker v1.13.1
	github.com/edsrzf/mmap-go v1.0.0
	github.com/elastic/gosigar v0.10.5
	github.com/fatih/color v1.9.0
	github.com/fjl/gencodec v0.0.0-20191126094850-e283372f291f // indirect
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08
	github.com/go-stack/stack v1.8.0
	github.com/golang/protobuf v1.3.3
	github.com/golang/snappy v0.0.1
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/howeyc/fsnotify v0.9.0 // indirect
	github.com/huin/goupnp v1.0.0
	github.com/influxdata/influxdb v1.7.9
	github.com/ipfs/go-cid v0.0.5
	github.com/ipfs/go-ipfs v0.4.23
	github.com/ipfs/go-ipld-format v0.0.2
	github.com/ipfs/interface-go-ipfs-core v0.2.6
	github.com/jackpal/go-nat-pmp v1.0.2
	github.com/julienschmidt/httprouter v1.3.0
	github.com/karalabe/usb v0.0.0-20191104083709-911d15fe12a9
	github.com/mattn/go-colorable v0.1.4
	github.com/mattn/go-isatty v0.0.12
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/multiformats/go-multihash v0.0.13
	github.com/naoina/toml v0.1.2-0.20170918210437-9fafd6967416
	github.com/olekukonko/tablewriter v0.0.4
	github.com/oschwald/maxminddb-golang v1.6.0 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/peterh/liner v1.1.1-0.20190123174540-a2c9a5303de7
	github.com/prometheus/tsdb v0.10.0
	github.com/rjeczalik/notify v0.9.2
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff
	github.com/rs/cors v1.7.0
	github.com/status-im/keycard-go v0.0.0-20200107115650-f38e9a19958e
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/tyler-smith/go-bip39 v1.0.2
	github.com/wsddn/go-ecdh v0.0.0-20161211032359-48726bab9208
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae
	golang.org/x/text v0.3.2
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
	gopkg.in/olebedev/go-duktape.v3 v3.0.0-20190709231704-1e4459ed25ff
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace github.com/ipfs/go-ipfs => github.com/tau-coin/go-ipfs v0.4.22-0.20200313092758-7b227442e904

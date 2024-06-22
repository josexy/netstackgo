module github.com/josexy/netstackgo

go 1.22.0

toolchain go1.22.4

require (
	golang.org/x/net v0.26.0
	golang.org/x/sys v0.21.0
	golang.org/x/time v0.5.0
	golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173
	// Fixed build on arm64, see: https://github.com/google/gvisor/issues/8237
	gvisor.dev/gvisor v0.0.0-20240622015726-dfeb44ecf5ac
)

require github.com/vishvananda/netlink v1.2.1-beta.2.0.20240524165444-4d4ba1473f21

require (
	github.com/google/btree v1.1.2 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
)

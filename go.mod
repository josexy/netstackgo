module github.com/josexy/netstackgo

go 1.22

require (
	golang.org/x/net v0.23.0
	golang.org/x/sys v0.18.0
	golang.org/x/time v0.5.0
	golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173
	// Fixed build on arm64, see: https://github.com/google/gvisor/issues/8237
	gvisor.dev/gvisor v0.0.0-20240403191413-660ba137b166
)

require (
	github.com/google/btree v1.1.2 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
)

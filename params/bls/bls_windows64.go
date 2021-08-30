// +build windows,amd64

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/x86_64-pc-windows-gnu -lbls_snark_sys -lm -lws2_32 -luserenv
*/
import "C"

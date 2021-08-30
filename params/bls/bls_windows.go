// +build windows,386

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/i686-pc-windows-gnu -lbls_snark_sys -lm -lws2_32 -luserenv
*/
import "C"

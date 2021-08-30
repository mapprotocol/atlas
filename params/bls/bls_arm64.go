// +build linux,arm64

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/aarch64-unknown-linux-gnu -lbls_snark_sys -ldl -lm
*/
import "C"

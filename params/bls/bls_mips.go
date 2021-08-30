// +build linux,mips

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/mips-unknown-linux-gnu -lbls_snark_sys -ldl -lm
*/
import "C"

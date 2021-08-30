// +build linux,mips64

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/mips64-unknown-linux-gnuabi64 -lbls_snark_sys -ldl -lm
*/
import "C"

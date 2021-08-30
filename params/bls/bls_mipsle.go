// +build linux,mipsle

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/mipsel-unknown-linux-gnu -lbls_snark_sys -ldl -lm
*/
import "C"

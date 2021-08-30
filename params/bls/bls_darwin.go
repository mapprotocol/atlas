// +build darwin,386,!ios

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/i686-apple-darwin -lbls_snark_sys -ldl -lm
*/
import "C"

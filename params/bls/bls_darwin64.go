// +build darwin,amd64,!ios

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/x86_64-apple-darwin -lbls_snark_sys -ldl -lm
*/
import "C"

// +build !android,linux,386,!musl

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/i686-unknown-linux-gnu -lbls_snark_sys -ldl -lm -lpthread
*/
import "C"

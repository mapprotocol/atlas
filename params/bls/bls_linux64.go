// +build !android,linux,amd64,!musl

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/x86_64-unknown-linux-gnu -lbls_snark_sys -ldl -lm -lpthread
*/
import "C"

// +build linux,mips64le

package bls

/*
#cgo LDFLAGS: -L${SRCDIR}/../libs/mips64el-unknown-linux-gnuabi64 -lbls_snark_sys -ldl -lm
*/
import "C"

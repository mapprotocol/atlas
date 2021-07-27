package main

import (
	"testing"
	"time"
)

// Monitoring account1 number
func Test_print(t *testing.T) {
	_, from1 := loadprivateCommon(keystore1)
	conn, _ := dialConnCommon()
	for {
		time.Sleep(time.Second * 1)
		queryRegisterInfo(conn, from1, "1:")
	}
}

// Monitoring account2 number
func Test_print2(t *testing.T) {
	_, from1 := loadprivateCommon(keystore1)
	conn, _ := dialConnCommon()
	for {
		time.Sleep(time.Second * 1)
		queryRegisterInfo(conn, from1, "1:")
	}
}

package main

import (
	"fmt"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
)

func withdrawMock(ctx *cli.Context) error {
	debugInfo := debugInfo{}
	debugInfo.preWork(ctx, true)
	debugInfo.withdrawMock(ctx) //change this
	return nil
}

func (d *debugInfo) withdrawMock(ctx *cli.Context) {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			currentEpoch1 := int(currentEpoch)
			for i := 0; i < len(d.step); i++ {
				if d.step[i] == currentEpoch1 {
					currentEpoch1 = i + 1
					break
				}
			}
			switch currentEpoch1 {
			case 1:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.doWithdraw()
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
				return
			}
		}
	}
}

func (d *debugInfo) doWithdraw() {
	fmt.Println("=================DO Withdraw========================")
	conn := d.client
	for k, _ := range d.relayerData {
		fmt.Println("ADDRESS:", d.relayerData[k].from)
		err := d.relayerData[k].withdraw(conn)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (d *debugInfo) withdrawAtDifferentEpoch12() {
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			currentEpoch1 := int(currentEpoch)
			for i := 0; i < len(d.step); i++ {
				if d.step[i] == currentEpoch1 {
					currentEpoch1 = i + 1
					break
				}
			}
			switch currentEpoch1 {
			case 1:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.doWithdraw()
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				fmt.Println("=====================================================")
				d.doWithdraw()
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
				return
			default:
				fmt.Println("over")
			}
		}
	}
}

func (d *debugInfo) withdrawAccordingToDifferentBalance12() {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			currentEpoch1 := int(currentEpoch)
			for i := 0; i < len(d.step); i++ {
				if d.step[i] == currentEpoch1 {
					currentEpoch1 = i + 1
					break
				}
			}
			switch currentEpoch1 {
			case 1:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.changeAllRegisterValue(500)
				d.doWithdraw()
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.changeAllRegisterValue(300)
				d.doWithdraw()
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.changeAllRegisterValue(100)
				d.doWithdraw()
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.changeAllRegisterValue(1000000)
				d.doWithdraw()
				d.atlasBackendCh <- NEXT_STEP
			}
		}
	}
}
func (r *relayerInfo) withdraw(conn *ethclient.Client) error {

	if int(r.registerValue) <= 0 {
		log.Fatal("Value must bigger than 0")
	}
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	value := new(big.Int).Mul(big.NewInt(r.registerValue), baseUnit)

	input := packInput("withdraw", r.from, value)

	sendContractTransaction(conn, r.from, RelayerAddress, new(big.Int).SetInt64(0), r.priKey, input)

	return nil
}

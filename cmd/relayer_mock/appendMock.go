package main

import (
	"fmt"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
)

func appendMock(ctx *cli.Context) error {
	debugInfo := debugInfo{}
	debugInfo.preWork(ctx, true)
	debugInfo.appendMock(ctx) //change this
	return nil
}

func (d *debugInfo) appendMock(ctx *cli.Context) {
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
				d.changeAllRegisterValue(100)
				d.doAppend()
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.changeAllRegisterValue(100)
				d.doAppend()
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.changeAllRegisterValue(100)
				d.doAppend()
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queryDebuginfo(QUERY_RELAYERINFO)
				d.queryDebuginfo(BALANCE)
				d.queryDebuginfo(REGISTER_BALANCE)
				d.doAppend()
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

func (d *debugInfo) doAppend() {
	fmt.Println("=================DO Withdraw========================")
	conn := d.client
	for k, _ := range d.relayerData {
		fmt.Println("ADDRESS:", d.relayerData[k].from)
		d.relayerData[k].Append(conn)
	}
}
func (r *relayerInfo) Append(conn *ethclient.Client) {
	if int(r.registerValue) <= 0 {
		log.Fatal("Value must bigger than 0")
	}
	value := ethToWei(r.registerValue)

	input := packInput("append", r.from, value)

	sendContractTransaction(conn, r.from, RelayerAddress, nil, r.priKey, input)
}

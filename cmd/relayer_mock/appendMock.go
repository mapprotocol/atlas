package main

import (
	"fmt"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
)

func appendMock(ctx *cli.Context) error {
	debugInfo := debugInfo{}
	debugInfo.preWork(ctx, []int{1, 2, 3}, true)
	debugInfo.appendMock(ctx) //change this
	return nil
}

func (d *debugInfo) appendMock(ctx *cli.Context) {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			switch currentEpoch {
			case 1:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.changeAllImpawnValue(100)
				d.doAppend()
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.changeAllImpawnValue(100)
				d.doAppend()
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.changeAllImpawnValue(100)
				d.doAppend()
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.doAppend()
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
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
	if int(r.impawnValue) <= 0 {
		log.Fatal("Value must bigger than 0")
	}
	value := ethToWei(r.impawnValue)

	input := packInput("append", r.from, value)

	sendContractTransaction(conn, r.from, RelayerAddress, nil, r.priKey, input)
}

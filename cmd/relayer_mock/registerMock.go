package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
)

func registerMock(ctx *cli.Context) error {
	debugInfo := debugInfo{}
	debugInfo.preWork(ctx, []int{1, 2, 3}, false)
	debugInfo.registerMock(ctx) //change this
	return nil
}

func (d *debugInfo) registerMock(ctx *cli.Context) {
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
				d.doRegister(ctx)
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.doRegister(ctx)
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

func (d *debugInfo) doRegister(ctx *cli.Context) {
	fmt.Println("=================DO Register========================")
	conn := d.client
	for k, _ := range d.relayerData {
		fmt.Println("ADDRESS:", d.relayerData[k].from)
		register11(ctx, conn, *d.relayerData[k])
	}
}

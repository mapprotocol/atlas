package writer

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Writer struct {
	conn *ethclient.Client
}

func New(conn *ethclient.Client) *Writer {
	return &Writer{
		conn: conn,
	}
}

func (w *Writer) ResolveMessage(m Message) bool {
	switch m.messageType {
	case SolveSendTranstion1:
		txHash, err := sendContractTransaction(w.conn, m.from, m.to, nil, m.priKey, m.input, m.gasLimit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		getResult(w.conn, txHash, true)
		m.DoneCh <- struct{}{}
	case SolveSendTranstion2:
		txHash, err := sendContractTransaction(w.conn, m.from, m.to, m.value, m.priKey, m.input, m.gasLimit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		getResult(w.conn, txHash, true)
		m.DoneCh <- struct{}{}
	case SolveQueryResult3:
		w.handleUnpackMethodSolveType3(m)
		m.DoneCh <- struct{}{}
	case SolveQueryResult4:
		w.handleUnpackMethodSolveType4(m)
		m.DoneCh <- struct{}{}
	default:
	}
	return true
}

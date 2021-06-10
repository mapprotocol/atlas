package ethereum

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/snowfork/polkadot-ethereum/relayer/crypto/secp256k1"
)

type Connection struct {
	endpoint string
	kp       *secp256k1.Keypair
	client   *ethclient.Client
	log      *logrus.Entry
}

func NewConnection(endpoint string, kp *secp256k1.Keypair, log *logrus.Entry) *Connection {
	return &Connection{
		endpoint: endpoint,
		kp:       kp,
		log:      log,
	}
}

func (c *Connection) Connect(ctx context.Context) error {
	client, err := ethclient.Dial(c.endpoint)
	if err != nil {
		return err
	}

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return err
	}

	c.log.WithFields(logrus.Fields{
		"endpoint": c.endpoint,
		"chainID":  chainID,
	}).Info("Connected to chain")

	c.client = client

	return nil
}

func (c *Connection) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *Connection) GetClient() *ethclient.Client {
	return c.client
}

func (c *Connection) GetKP() *secp256k1.Keypair {
	return c.kp
}

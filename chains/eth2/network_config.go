package eth2

import "fmt"

type NetworkConfig struct {
	GenesisValidatorsRoot [32]byte
	BellatrixForkVersion  ForkVersion
	BellatrixForkEpoch    uint64
	CapellaForkVersion    ForkVersion
	CapellaForkEpoch      uint64
}

func newNetworkConfig(chainID uint64) (*NetworkConfig, error) {
	switch chainID {
	case 1: // Mainnet
		return &NetworkConfig{
			GenesisValidatorsRoot: [32]byte{
				0x4b, 0x36, 0x3d, 0xb9, 0x4e, 0x28, 0x61, 0x20, 0xd7, 0x6e, 0xb9, 0x05, 0x34,
				0x0f, 0xdd, 0x4e, 0x54, 0xbf, 0xe9, 0xf0, 0x6b, 0xf3, 0x3f, 0xf6, 0xcf, 0x5a,
				0xd2, 0x7f, 0x51, 0x1b, 0xfe, 0x95,
			},
			BellatrixForkVersion: [4]byte{0x02, 0x00, 0x00, 0x00},
			BellatrixForkEpoch:   144896,
			CapellaForkVersion:   [4]byte{0x03, 0x00, 0x00, 0x00},
			CapellaForkEpoch:     194048,
		}, nil
	case 5: // Goerli
		return &NetworkConfig{
			GenesisValidatorsRoot: [32]byte{
				0x04, 0x3d, 0xb0, 0xd9, 0xa8, 0x38, 0x13, 0x55, 0x1e, 0xe2, 0xf3, 0x34, 0x50,
				0xd2, 0x37, 0x97, 0x75, 0x7d, 0x43, 0x09, 0x11, 0xa9, 0x32, 0x05, 0x30, 0xad,
				0x8a, 0x0e, 0xab, 0xc4, 0x3e, 0xfb,
			},
			BellatrixForkVersion: [4]byte{0x02, 0x00, 0x10, 0x20},
			BellatrixForkEpoch:   112260,
			CapellaForkVersion:   [4]byte{0x03, 0x00, 0x10, 0x20},
			CapellaForkEpoch:     162304,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported network chain ID %d", chainID)
	}
}

// Return the fork version at the given epoch
func (nc *NetworkConfig) computeForkVersion(epoch uint64) *ForkVersion {
	if epoch >= nc.CapellaForkEpoch {
		return &nc.CapellaForkVersion
	}

	if epoch >= nc.BellatrixForkEpoch {
		return &nc.BellatrixForkVersion
	}

	return nil
}

// Return the fork version at the given epoch
func (nc *NetworkConfig) computeForkVersionBySlot(slot uint64) *ForkVersion {
	return nc.computeForkVersion(computeEpochAtSlot(slot))
}

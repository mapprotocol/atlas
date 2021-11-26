package reserve

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	ErrTobinTaxZeroDenominator  = errors.New("tobin tax denominator equal to zero")
	ErrTobinTaxInvalidNumerator = errors.New("tobin tax numerator greater than denominator")
)

type Ratio struct {
	numerator, denominator *big.Int
}

func (r *Ratio) Apply(value *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(r.numerator, value), r.denominator)
}

func TobinTax(vmRunner vm.EVMRunner, sender common.Address) (tax Ratio, reserveAddress common.Address, err error) {

	reserveAddress, err = contracts.GetRegisteredAddress(vmRunner, params.ReserveRegistryId)
	if err != nil {
		return Ratio{}, params.ZeroAddress, err
	}

	ret, err := vmRunner.ExecuteFrom(sender, reserveAddress, params.TobinTaxFunctionSelector, params.MaxGasForGetOrComputeTobinTax, big.NewInt(0))
	if err != nil {
		return Ratio{}, params.ZeroAddress, err
	}

	// Expected size of ret is 64 bytes because getOrComputeTobinTax() returns two uint256 values,
	// each of which is equivalent to 32 bytes
	if binary.Size(ret) != 64 {
		return Ratio{}, params.ZeroAddress, errors.New("length of tobin tax not equal to 64 bytes")
	}
	numerator := new(big.Int).SetBytes(ret[0:32])
	denominator := new(big.Int).SetBytes(ret[32:64])
	if denominator.Cmp(common.Big0) == 0 {
		return Ratio{}, params.ZeroAddress, ErrTobinTaxZeroDenominator
	}
	if numerator.Cmp(denominator) == 1 {
		return Ratio{}, params.ZeroAddress, ErrTobinTaxInvalidNumerator
	}
	return Ratio{numerator, denominator}, reserveAddress, nil
}

func ComputeTobinTax(vmRunner vm.EVMRunner, sender common.Address, transferAmount *big.Int) (tax *big.Int, taxRecipient common.Address, err error) {
	return nil, params.ZeroAddress, nil
	//TODO Replace with the following in the future
	taxRatio, recipient, err := TobinTax(vmRunner, sender)
	if err != nil {
		return nil, params.ZeroAddress, err
	}

	return taxRatio.Apply(transferAmount), recipient, nil
}

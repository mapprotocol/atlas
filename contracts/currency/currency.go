package currency

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/contracts"
	"github.com/mapprotocol/atlas/contracts/abis"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
)

var (
	medianRateMethod   = contracts.NewRegisteredContractMethod(params.SortedOraclesRegistryId, abis.SortedOracles, "medianRate", params.MaxGasForMedianRate)
	getWhitelistMethod = contracts.NewRegisteredContractMethod(params.FeeCurrencyWhitelistRegistryId, abis.FeeCurrency, "getWhitelist", params.MaxGasForGetWhiteList)
	getBalanceMethod   = contracts.NewMethod(abis.ERC20, "balanceOf", params.MaxGasToReadErc20Balance)
)

// NoopExchangeRate represents an exchange rate of 1 to 1
var NoopExchangeRate = ExchangeRate{common.Big1, common.Big1}

var MAPCurrency = Currency{
	Address:   params.ZeroAddress,
	toMAPRate: NoopExchangeRate,
}

// Currency represent a system currency
// than can be converted to MAP
// Two currencies are deemed equal if they have the same address
type Currency struct {
	Address   common.Address
	toMAPRate ExchangeRate
}

// ToMAP converts an currency's token amount to a MAP amount
func (c *Currency) ToMAP(tokenAmount *big.Int) *big.Int {
	return c.toMAPRate.ToBase(tokenAmount)
}

// FromMAP converts an MAP amount to a currency tokens amount
func (c *Currency) FromMAP(mapAmount *big.Int) *big.Int {
	return c.toMAPRate.FromBase(mapAmount)
}

// CmpToCurrency compares a currency amount to an amount in a different currency
func (c *Currency) CmpToCurrency(currencyAmount *big.Int, sndCurrencyAmount *big.Int, sndCurrency *Currency) int {
	if c == sndCurrency || c.Address == sndCurrency.Address {
		return currencyAmount.Cmp(sndCurrencyAmount)
	}

	// Below code block is basically evaluating this comparison:
	// currencyAmount * c.toMAPRate.denominator / c.toMAPRate.numerator < sndCurrencyAmount * sndCurrency.toMAPRate.denominator / sndCurrency.toMAPRate.numerator
	// It will transform that comparison to this, to remove having to deal with fractional values.
	// currencyAmount * c.toMAPRate.denominator * sndCurrency.toMAPRate.numerator < sndCurrencyAmount * sndCurrency.toMAPRate.denominator * c.toMAPRate.numerator
	leftSide := new(big.Int).Mul(
		currencyAmount,
		new(big.Int).Mul(
			c.toMAPRate.denominator,
			sndCurrency.toMAPRate.numerator,
		),
	)
	rightSide := new(big.Int).Mul(
		sndCurrencyAmount,
		new(big.Int).Mul(
			sndCurrency.toMAPRate.denominator,
			c.toMAPRate.numerator,
		),
	)

	return leftSide.Cmp(rightSide)
}

// ExchangeRate represent the exchangeRate [Base -> Token]
// Follows the equation: 1 base * ExchangeRate = X token
type ExchangeRate struct {
	numerator   *big.Int
	denominator *big.Int
}

// NewExchangeRate creates an exchange rate.
// Requires numerator >=0 && denominator >= 0
func NewExchangeRate(numerator *big.Int, denominator *big.Int) (*ExchangeRate, error) {
	if numerator == nil || common.Big0.Cmp(numerator) >= 0 {
		return nil, contracts.ErrExchangeRateZero
	}
	if denominator == nil || common.Big0.Cmp(denominator) >= 0 {
		return nil, contracts.ErrExchangeRateZero
	}
	return &ExchangeRate{numerator, denominator}, nil
}

// ToBase converts from token to base
func (er *ExchangeRate) ToBase(tokenAmount *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(tokenAmount, er.denominator), er.numerator)
}

// FromGold converts from base to token
func (er *ExchangeRate) FromBase(goldAmount *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(goldAmount, er.numerator), er.denominator)
}

// CurrencyManager provides an interface to access different fee currencies on a given point in time (header,state)
// and doing comparison or fetching exchange rates
//
// It's implements an internal cache to avoid perfoming duplicated EVM calls
type CurrencyManager struct {
	vmRunner vm.EVMRunner

	currencyCache    map[common.Address]*Currency                               // map of exchange rates of the form (MAP, token)
	_getExchangeRate func(vm.EVMRunner, *common.Address) (*ExchangeRate, error) // function to obtain exchange rate from blockchain state
}

// NewManager creates a new CurrencyManager
func NewManager(vmRunner vm.EVMRunner) *CurrencyManager {
	return newManager(GetExchangeRate, vmRunner)
}

func newManager(_getExchangeRate func(vm.EVMRunner, *common.Address) (*ExchangeRate, error), vmRunner vm.EVMRunner) *CurrencyManager {
	return &CurrencyManager{
		vmRunner:         vmRunner,
		currencyCache:    make(map[common.Address]*Currency),
		_getExchangeRate: _getExchangeRate,
	}
}

// GetCurrency retrieves fee currency
func (cc *CurrencyManager) GetCurrency(currencyAddress *common.Address) (*Currency, error) {
	if currencyAddress == nil {
		return &MAPCurrency, nil
	}

	val, ok := cc.currencyCache[*currencyAddress]
	if ok {
		return val, nil
	}

	currencyExchangeRate, err := cc._getExchangeRate(cc.vmRunner, currencyAddress)
	if err != nil {
		return nil, err
	}

	val = &Currency{
		Address:   *currencyAddress,
		toMAPRate: *currencyExchangeRate,
	}

	cc.currencyCache[*currencyAddress] = val

	return val, nil
}

// CmpValues compares values of potentially different currencies
func (cc *CurrencyManager) CmpValues(val1 *big.Int, currencyAddr1 *common.Address, val2 *big.Int, currencyAddr2 *common.Address) int {
	// Short circuit if the fee currency is the same. nil currency => native currency
	if (currencyAddr1 == nil && currencyAddr2 == nil) || (currencyAddr1 != nil && currencyAddr2 != nil && *currencyAddr1 == *currencyAddr2) {
		return val1.Cmp(val2)
	}

	currency1, err1 := cc.GetCurrency(currencyAddr1)
	currency2, err2 := cc.GetCurrency(currencyAddr2)

	if err1 != nil || err2 != nil {
		currency1Output := "nil"
		if currencyAddr1 != nil {
			currency1Output = currencyAddr1.Hex()
		}
		currency2Output := "nil"
		if currencyAddr2 != nil {
			currency2Output = currencyAddr2.Hex()
		}
		log.Warn("Error in retrieving exchange rate.  Will do comparison of two values without exchange rate conversion.", "currency1", currency1Output, "err1", err1, "currency2", currency2Output, "err2", err2)
		return val1.Cmp(val2)
	}

	return currency1.CmpToCurrency(val1, val2, currency2)
}

// GetExchangeRate retrieves currency-to-MAP exchange rate
func GetExchangeRate(vmRunner vm.EVMRunner, currencyAddress *common.Address) (*ExchangeRate, error) {
	if currencyAddress == nil {
		return &NoopExchangeRate, nil
	}

	var returnArray [2]*big.Int

	err := medianRateMethod.Query(vmRunner, &returnArray, currencyAddress)

	if err == contracts.ErrSmartContractNotDeployed {
		log.Warn("Registry address lookup failed", "err", err)
		return &NoopExchangeRate, nil
	} else if err != nil {
		log.Error("medianRate invocation error", "feeCurrencyAddress", currencyAddress.Hex(), "err", err)
		return &NoopExchangeRate, nil
	}

	log.Trace("medianRate invocation success", "feeCurrencyAddress", currencyAddress, "returnArray", returnArray)
	return NewExchangeRate(returnArray[0], returnArray[1])
}

// GetBalanceOf returns an account's balance on a given ERC20 currency
func GetBalanceOf(vmRunner vm.EVMRunner, accountOwner common.Address, contractAddress common.Address) (result *big.Int, err error) {

	log.Trace("GetBalanceOf() Called", "accountOwner", accountOwner.Hex(), "contractAddress", contractAddress)

	err = getBalanceMethod.Bind(contractAddress).Query(vmRunner, &result, accountOwner)

	if err != nil {
		log.Error("GetBalanceOf evm invocation error", "err", err)
	} else {
		log.Trace("GetBalanceOf evm invocation success", "accountOwner", accountOwner.Hex(), "Balance", result.String())
	}

	return result, err
}

// CurrencyWhitelist retrieves the list of currencies that can be used to pay transaction fees
func CurrencyWhitelist(vmRunner vm.EVMRunner) ([]common.Address, error) {
	returnList := []common.Address{}
	return returnList, nil
	//TODO Replace with the following in the future
	err := getWhitelistMethod.Query(vmRunner, &returnList)

	if err == contracts.ErrSmartContractNotDeployed {
		log.Warn("Registry address lookup failed", "err", err)
	} else if err != nil {
		log.Error("getWhitelist invocation failed", "err", err)
	} else {
		log.Trace("getWhitelist invocation success")
	}

	return returnList, err
}

// IsWhitelisted indicates if a currency is whitelisted for transaction fee payments
func IsWhitelisted(vmRunner vm.EVMRunner, feeCurrency *common.Address) bool {
	if feeCurrency == nil {
		return true
	}

	whitelistedCurrencies, err := CurrencyWhitelist(vmRunner)
	if err != nil {
		return true
	}

	for _, addr := range whitelistedCurrencies {
		if addr == *feeCurrency {
			return true
		}
	}
	return false
}

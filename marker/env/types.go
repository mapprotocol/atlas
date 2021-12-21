package env

import (
	"math/big"
)

// Config represents marker environment parameters
type Config struct {
	ChainID  *big.Int       `json:"chainId"`  // chainId identifies the current chain and is used for replay protection
	Accounts AccountsConfig `json:"accounts"` // Accounts configuration for the environment
}

// AccountsConfig represents accounts configuration for the environment
type AccountsConfig struct {
	Mnemonic             string `json:"mnemonic"`            // Accounts mnemonic
	NumValidators        int    `json:"validators"`          // Number of initial validators
	NumDeveloperAccounts int    `json:"developerAccounts"`   // Number of developers accounts
	UseValidatorAsAdmin  bool   `json:"useValidatorAsAdmin"` // Whether to use the first validator as the admin (for compatibility with monorepo)
}

// AdminAccount returns the environment's admin account
func (ac *AccountsConfig) AdminAccount() *Account {
	at := AdminAT
	if ac.UseValidatorAsAdmin {
		at = ValidatorAT
	}
	acc, err := DeriveAccount(ac.Mnemonic, at, 0)
	if err != nil {
		panic(err)
	}
	return acc
}

// DeveloperAccounts returns the environment's developers accounts
func (ac *AccountsConfig) DeveloperAccounts() []Account {
	accounts, err := DeriveAccountList(ac.Mnemonic, DeveloperAT, ac.NumDeveloperAccounts)
	if err != nil {
		panic(err)
	}
	return accounts
}

// Account retrieves the account corresponding to the (accountType, idx)
func (ac *AccountsConfig) Account(accType AccountType, idx int) (*Account, error) {
	return DeriveAccount(ac.Mnemonic, accType, idx)
}

// ValidatorAccounts returns the environment's validators accounts
func (ac *AccountsConfig) ValidatorAccounts() []Account {
	accounts, err := DeriveAccountList(ac.Mnemonic, ValidatorAT, ac.NumValidators)
	if err != nil {
		panic(err)
	}
	return accounts
}

// ValidatorAccounts returns the environment's validators accounts
func (ac *AccountsConfig) TxFeeRecipientAccounts() []Account {
	accounts, err := DeriveAccountList(ac.Mnemonic, TxFeeRecipientAT, ac.NumValidators)
	if err != nil {
		panic(err)
	}
	return accounts
}

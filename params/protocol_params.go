package params

const (
	DefaultGasLimit uint64 = 20000000 // Gas limit of the blocks before BlockchainParams contract is loaded.

	thousand = 1000
	million  = 1000 * 1000

	// Default intrinsic gas cost of transactions paying for gas in alternative currencies.
	// Calculated to estimate 1 balance read, 1 debit, and 4 credit transactions.
	IntrinsicGasForAlternativeFeeCurrency uint64 = 50 * thousand

	// Contract communication gas limits
	MaxGasForCalculateTargetEpochPaymentAndRewards uint64 = 2 * million
	MaxGasForCommitments                           uint64 = 2 * million
	MaxGasForComputeCommitment                     uint64 = 2 * million
	MaxGasForBlockRandomness                       uint64 = 2 * million
	MaxGasForDebitGasFeesTransactions              uint64 = 1 * million
	MaxGasForCreditGasFeesTransactions             uint64 = 1 * million
	MaxGasForDistributeEpochPayment                uint64 = 1 * million
	MaxGasForDistributeVoterEpochRewards           uint64 = 1 * million
	MaxGasForElectValidators                       uint64 = 50 * million
	MaxGasForElectNValidatorSigners                uint64 = 50 * million
	MaxGasForActiveAllPending                      uint64 = 5000 * million
	MaxGasForGetAddressFor                         uint64 = 100 * million
	MaxGasForGetElectableValidators                uint64 = 100 * thousand
	MaxGasForGetEligibleValidatorsVoteTotals       uint64 = 1 * million
	MaxGasForGetGasPriceMinimum                    uint64 = 2 * million
	MaxGasForGetGroupEpochRewards                  uint64 = 500 * thousand
	MaxGasForGetPledgeMultiplierInReward           uint64 = 50 * thousand
	MaxGasForGetOrComputeTobinTax                  uint64 = 1 * million
	MaxGasForGetRegisteredValidators               uint64 = 2 * million
	MaxGasForGetValidator                          uint64 = 100 * thousand
	MaxGasForGetWhiteList                          uint64 = 200 * thousand
	MaxGasForGetTransferWhitelist                  uint64 = 2 * million
	MaxGasForIncreaseSupply                        uint64 = 50 * thousand
	MaxGasForIsFrozen                              uint64 = 20 * thousand
	MaxGasForMedianRate                            uint64 = 100 * thousand
	MaxGasForReadBlockchainParameter               uint64 = 40 * thousand // ad-hoc measurement is ~26k
	MaxGasForRevealAndCommit                       uint64 = 2 * million
	MaxGasForUpdateGasPriceMinimum                 uint64 = 2 * million
	MaxGasForUpdateTargetVotingYield               uint64 = 2 * million
	MaxGasForUpdateValidatorScore                  uint64 = 1 * million
	MaxGasForTotalSupply                           uint64 = 50 * thousand
	MaxGasForMintGas                               uint64 = 5 * million
	MaxGasToReadErc20Balance                       uint64 = 100 * thousand
	MaxGasForIsReserveLow                          uint64 = 1 * million
	MaxGasForGetCommunityPartnerSettingPartner     uint64 = 100 * thousand
	MaxGasForGetMgrMaintainerAddress               uint64 = 100 * thousand

	////////////////////////////////////////////////////////////////////////////////////////////////
	CallValueTransferGas uint64 = 9000  // Paid for CALL when the value transfer is non-zero.
	CallNewAccountGas    uint64 = 25000 // Paid for CALL when the destination address didn't exist prior.
	Sha3Gas              uint64 = 30    // Once per SHA3 operation.
	Sha3WordGas          uint64 = 6     // Once per word of the SHA3 operation's data.
	// Precompiled contract gas prices
	EcrecoverGas        uint64 = 3000 // Elliptic curve sender recovery gas price
	Sha256BaseGas       uint64 = 60   // Base price for a SHA256 operation
	Sha256PerWordGas    uint64 = 12   // Per-word price for a SHA256 operation
	Ripemd160BaseGas    uint64 = 600  // Base price for a RIPEMD160 operation
	Ripemd160PerWordGas uint64 = 120  // Per-word price for a RIPEMD160 operation
	IdentityBaseGas     uint64 = 15   // Base price for a data copy operation
	IdentityPerWordGas  uint64 = 3    // Per-work price for a data copy operation
	ModExpQuadCoeffDiv  uint64 = 20   // Divisor for the quadratic particle of the big int modular exponentiation

	Bn256AddGasByzantium             uint64 = 500    // Byzantium gas needed for an elliptic curve addition
	Bn256AddGasIstanbul              uint64 = 150    // Gas needed for an elliptic curve addition
	Bn256ScalarMulGasByzantium       uint64 = 40000  // Byzantium gas needed for an elliptic curve scalar multiplication
	Bn256ScalarMulGasIstanbul        uint64 = 6000   // Gas needed for an elliptic curve scalar multiplication
	Bn256PairingBaseGasByzantium     uint64 = 100000 // Byzantium base price for an elliptic curve pairing check
	Bn256PairingBaseGasIstanbul      uint64 = 45000  // Base price for an elliptic curve pairing check
	Bn256PairingPerPointGasByzantium uint64 = 80000  // Byzantium per-point price for an elliptic curve pairing check
	Bn256PairingPerPointGasIstanbul  uint64 = 34000  // Per-point price for an elliptic curve pairing check
	// Atlas precompiled contracts
	FractionMulExpGas           uint64 = 50     // Cost of performing multiplication and exponentiation of fractions to an exponent of up to 10^3.
	ProofOfPossessionGas        uint64 = 350000 // Cost of verifying a BLS proof of possession.
	GetValidatorGas             uint64 = 1000   // Cost of reading a validator's address.
	GetValidatorBLSGas          uint64 = 1000   // Cost of reading a validator's BLS public key.
	GetEpochSizeGas             uint64 = 10     // Cost of querying the number of blocks in an epoch.
	GetBlockNumberFromHeaderGas uint64 = 10     // Cost of decoding a block header.
	HashHeaderGas               uint64 = 10     // Cost of hashing a block header.
	GetParentSealBitmapGas      uint64 = 100    // Cost of reading the parent seal bitmap from the chain.
	// May take a bit more time with 100 validators, need to bench that
	GetVerifiedSealBitmapGas uint64 = 350000           // Cost of verifying the seal on a given RLP encoded header.
	Ed25519VerifyGas         uint64 = 1500             // Gas needed for and Ed25519 signature verification
	Sha2_512BaseGas          uint64 = Sha256BaseGas    // Base price for a Sha2-512 operation
	Sha2_512PerWordGas       uint64 = Sha256PerWordGas // Per-word price for a Sha2-512 operation

	Sha3_256BaseGas     uint64 = Sha3Gas     // Base price for a Sha3-256 operation
	Sha3_256PerWordGas  uint64 = Sha3WordGas // Per-word price for a sha3-256 operation
	Sha3_512BaseGas     uint64 = Sha3Gas     // Base price for a Sha3-512 operation
	Sha3_512PerWordGas  uint64 = Sha3WordGas // Per-word price for a Sha3-512 operation
	Keccak512BaseGas    uint64 = Sha3Gas     // Per-word price for a Keccak512 operation
	Keccak512PerWordGas uint64 = Sha3WordGas // Base price for a Keccak512 operation

	Blake2sBaseGas    uint64 = Sha256BaseGas    // Per-word price for a Blake2s operation
	Blake2sPerWordGas uint64 = Sha256PerWordGas // Base price for a Blake2s
	InvalidCip20Gas   uint64 = 200              // Price of attempting to access an unsupported CIP20 hash function

	Bls12377G1AddGas          uint64 = 600   // Price for BLS12-377 elliptic curve G1 point addition
	Bls12377G1MulGas          uint64 = 12000 // Price for BLS12-377 elliptic curve G1 point scalar multiplication
	Bls12377G2AddGas          uint64 = 4500  // Price for BLS12-377 elliptic curve G2 point addition
	Bls12377G2MulGas          uint64 = 55000 // Price for BLS12-377 elliptic curve G2 point scalar multiplication
	Bls12377PairingBaseGas    uint64 = 65000 // Base gas price for BLS12-377 elliptic curve pairing check
	Bls12377PairingPerPairGas uint64 = 55000 // Per-point pair gas price for BLS12-377 elliptic curve pairing check

	Bls12381G1AddGas          uint64 = 600   // Price for BLS12-381 elliptic curve G1 point addition
	Bls12381G1MulGas          uint64 = 12000 // Price for BLS12-381 elliptic curve G1 point scalar multiplication
	Bls12381G2AddGas          uint64 = 800   // Price for BLS12-381 elliptic curve G2 point addition
	Bls12381G2MulGas          uint64 = 45000 // Price for BLS12-381 elliptic curve G2 point scalar multiplication
	Bls12381PairingBaseGas    uint64 = 65000 // Base gas price for BLS12-381 elliptic curve pairing check
	Bls12381PairingPerPairGas uint64 = 43000 // Per-point pair gas price for BLS12-381 elliptic curve pairing check
	Bls12381MapG1Gas          uint64 = 5500  // Gas price for BLS12-381 mapping field element to G1 operation
	Bls12381MapG2Gas          uint64 = 75000 // Gas price for BLS12-381 mapping field element to G2 operation

	MaxCodeSize = 49152 // Maximum bytecode to permit for a contract
)

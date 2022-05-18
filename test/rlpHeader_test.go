package test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/core/types"
	blscrypto "github.com/mapprotocol/atlas/helper/bls"
	"golang.org/x/crypto/sha3"
	"io"
	"log"
	"math/big"
	"math/rand"
	"testing"
)

func TestData(t *testing.T) {
	//url := fmt.Sprintf("http://127.0.0.1:7445")
	//ip := "127.0.0.1" //utils.RPCListenAddrFlag.Name)
	//port := 7415                //utils.RPCPortFlag.Name)
	//url := fmt.Sprintf("http://%s", fmt.Sprintf("%s:%d", ip, port))
	url := fmt.Sprintf("https://poc2-rpc.maplabs.io")
	conn, err := Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the Atlaschain client: %v", err)
	}
	Header, _ := conn.HeaderByNumber(context.Background(), big.NewInt(50))
	if Header == nil {
		t.Fatal("header is nil")
	}
	//fmt.Println("ParentHash",Header.ParentHash)
	//fmt.Println("Coinbase",Header.Coinbase)
	//fmt.Println("Root",Header.Root)
	//fmt.Println("TxHash",Header.TxHash)
	//fmt.Println("ReceiptHash",Header.ReceiptHash)
	//fmt.Printf("Bloom:%x\n",Header.Bloom)
	//fmt.Println("Number",Header.Number)
	//fmt.Println("GasLimit",Header.GasLimit)
	//fmt.Println("GasUsed",Header.GasUsed)
	//fmt.Println("Time",Header.Time)
	//fmt.Printf("Extra:%x\n",Header.Extra)
	//fmt.Printf("Extra[32:]:%x\n",Header.Extra[32:])
	//fmt.Println("MixDigest",Header.MixDigest)
	//fmt.Printf("Nonce:%x\n",Header.Nonce)
	//fmt.Println("BaseFee",Header.BaseFee)
	//fmt.Println("================================================")

	//fmt.Println("sigHash2(Header):",sigHash2(Header))
	HB, _ := rlp.EncodeToBytes(&Header)
	fmt.Printf("rlpHeader: %x \n", HB)
	fmt.Println("sigHash(Header):", sigHash(Header))
	fmt.Println("hash:", Header.Hash())

	//HB, _ := rlp.EncodeToBytes(&Header)
	//fmt.Printf("Header Bytes:%x\n",HB)
	//fmt.Println("Header.Coinbase:",Header.Coinbase)
	//fmt.Println("Header.Hash:",Header.Hash())
	//fmt.Printf("extra:%x\n",Header.Extra[32:])
	//extra, err := types.ExtractIstanbulExtra(Header)
	//fmt.Println("extra addr:",extra.AddedValidators)
	//fmt.Println("extra pk:",extra.AddedValidatorsPublicKeys)
	//for i:=0;i< len(extra.AddedValidatorsPublicKeys);i++{
	//	fmt.Printf("%d pk:%x\n",i,extra.AddedValidatorsPublicKeys[i])
	//}
	//fmt.Printf("extra seal: %x \n",extra.Seal)
	//fmt.Println("extra removed:",extra.RemovedValidators.Bits())
	//fmt.Println("extra blsSeal:",extra.AggregatedSeal.String())
	//fmt.Println("extra parentBlsSeal:",extra.ParentAggregatedSeal.String())
	//
	//addr, err := istanbul.GetSignatureAddress(sigHash(Header).Bytes(), extra.Seal)
	//fmt.Printf("verify params: %x %x \n",sigHash(Header).Bytes(), extra.Seal)
	//fmt.Println("verify sign",addr,Header.Coinbase,err)
}

func sigHash2(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	// Clean seal is required for calculating proposer seal.
	//rlp.Encode(hasher, header) //types.IstanbulFilteredHeader(header, false)//Hash(HeaderWithExtra)
	//rlp.Encode(hasher, types.IstanbulFilteredHeader(header, false))//sigHash()
	rlp.Encode(hasher, types.IstanbulFilteredHeader(header, true)) //Header.Hash()
	hasher.Sum(hash[:0])
	return hash
}

func TestRlp(t *testing.T) {
	var addrs []common.Address
	var publicKeys []blscrypto.SerializedPublicKey

	addr, pk := makeKey()
	addrs = append(addrs, addr)
	publicKeys = append(publicKeys, pk)
	fmt.Printf("pk1:%x\n", pk)
	addr, pk = makeKey()
	addrs = append(addrs, addr)
	publicKeys = append(publicKeys, pk)
	fmt.Printf("pk2:%x\n", pk)
	fmt.Println("pks,pk", len(publicKeys), len(pk))
	ist := types.IstanbulExtra{
		AddedValidators:           addrs,
		AddedValidatorsPublicKeys: publicKeys,
		RemovedValidators:         nil, // big.NewInt(1),
		Seal:                      nil, //common.FromHex("0x9f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc797016703885901"),
		AggregatedSeal:            types.IstanbulAggregatedSeal{big.NewInt(1000000000), common.FromHex("0x15c8954e105e86ea231c0af668416d4a6260da9bde72047e1af44828cb5d6cc571ce32787b4ef3850685425072f36600"), big.NewInt(1)},
		ParentAggregatedSeal:      types.IstanbulAggregatedSeal{}, //types.IstanbulAggregatedSeal{big.NewInt(2),common.FromHex("0x15c8954e105e86ea231c0af668416d4a6260da9bde72047e1af44828cb5d6cc571ce32787b4ef3850685425072f36600"),big.NewInt(2)},
	}
	fmt.Println("=======================================")
	fmt.Println("data AddedValidators", ist.AddedValidators)
	fmt.Println("data AddedValidatorsPublicKeys", ist.AddedValidatorsPublicKeys)
	fmt.Printf("data Seal %x\n", ist.Seal)
	fmt.Println("data RemovedValidators", ist.RemovedValidators)
	fmt.Println("data AggregatedSeal", ist.AggregatedSeal)
	fmt.Println("data ParentAggregatedSeal", ist.ParentAggregatedSeal)
	fmt.Println("=======================================")

	istPayload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		t.Fatal("failed to encode istanbul extra")
	}
	fmt.Println("len:", len(istPayload))
	fmt.Printf("encode byte: %x\n", istPayload)

	var istanbulExtra *IstanbulExtra
	err = rlp.DecodeBytes(istPayload, &istanbulExtra)
	if err != nil {
		t.Fatal("decode err", err)
	}
	fmt.Println("=======================================")
	fmt.Println("decode data AddedValidators", istanbulExtra.AddedValidators)
	fmt.Println("decode data AddedValidatorsPublicKeys", istanbulExtra.AddedValidatorsPublicKeys)
	fmt.Printf("decode data Seal %x\n", istanbulExtra.Seal)
	fmt.Println("decode data RemovedValidators", istanbulExtra.RemovedValidators)
	fmt.Println("decode data AggregatedSeal", istanbulExtra.AggregatedSeal)
	fmt.Println("decode data ParentAggregatedSeal", istanbulExtra.ParentAggregatedSeal)
	fmt.Println("=======================================")
}

type IstanbulExtra struct {
	// AddedValidators are the validators that have been added in the block
	AddedValidators []common.Address
	// AddedValidatorsPublicKeys are the BLS public keys for the validators added in the block
	AddedValidatorsPublicKeys []blscrypto.SerializedPublicKey
	// RemovedValidators is a bitmap having an active bit for each removed validator in the block
	RemovedValidators *big.Int
	// Seal is an ECDSA signature by the proposer
	Seal []byte
	// AggregatedSeal contains the aggregated BLS signature created via IBFT consensus.
	AggregatedSeal IstanbulAggregatedSeal
	// ParentAggregatedSeal contains and aggregated BLS signature for the previous block.
	ParentAggregatedSeal IstanbulAggregatedSeal
}

type IstanbulAggregatedSeal struct {
	// Bitmap is a bitmap having an active bit for each validator that signed this block
	Bitmap *big.Int
	// Signature is an aggregated BLS signature resulting from signatures by each validator that signed this block
	Signature []byte
	// Round is the round in which the signature was created.
	Round *big.Int
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (ist *IstanbulExtra) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		ist.AddedValidators,
		ist.AddedValidatorsPublicKeys,
		ist.RemovedValidators,
		ist.Seal,
		&ist.AggregatedSeal,
		&ist.ParentAggregatedSeal,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (ist *IstanbulExtra) DecodeRLP(s *rlp.Stream) error {
	var istanbulExtra struct {
		AddedValidators           []common.Address
		AddedValidatorsPublicKeys []blscrypto.SerializedPublicKey
		RemovedValidators         *big.Int
		Seal                      []byte
		AggregatedSeal            IstanbulAggregatedSeal
		ParentAggregatedSeal      IstanbulAggregatedSeal
	}
	if err := s.Decode(&istanbulExtra); err != nil {
		return err
	}
	ist.AddedValidators, ist.AddedValidatorsPublicKeys, ist.RemovedValidators, ist.Seal, ist.AggregatedSeal, ist.ParentAggregatedSeal = istanbulExtra.AddedValidators, istanbulExtra.AddedValidatorsPublicKeys, istanbulExtra.RemovedValidators, istanbulExtra.Seal, istanbulExtra.AggregatedSeal, istanbulExtra.ParentAggregatedSeal
	return nil
}

func makeKey() (common.Address, blscrypto.SerializedPublicKey) {
	privateKey, _ := crypto.GenerateKey()
	blsPrivateKey, _ := blscrypto.CryptoType().ECDSAToBLS(privateKey)
	blsPublicKey, _ := blscrypto.CryptoType().PrivateToPublic(blsPrivateKey)
	bp, _ := blsPublicKey.MarshalText()
	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Println("address:", from)
	fmt.Printf("bls public key: %s\n", bp)
	return from, blsPublicKey
}

func TestRlpEncode(t *testing.T) {
	te := TestEncode{
		testbytes(256),
		testbytes(256),
		common.FromHex("0x9f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc797016703885901"),
	}
	istPayload, err := rlp.EncodeToBytes(&te)
	if err != nil {
		t.Fatal("failed to encode istanbul extra")
	}
	fmt.Println("len:", len(istPayload))
	fmt.Printf("encode bytes %x\n", istPayload)

	///////////////////////////////////////////////////////////
	te1 := TestEncode1{
		testbytes(256),
		testbytes(256),
	}
	istPayload1, err := rlp.EncodeToBytes(&te1)
	if err != nil {
		t.Fatal("failed to encode istanbul extra")
	}
	fmt.Println("len:", len(istPayload1))
	fmt.Printf("encode bytes: %x\n", istPayload1)
	/////////////////////////////////////////////////////////////
	te2 := TestEncode2{
		common.FromHex("0x9f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc797016703885901"),
	}
	istPayload2, err := rlp.EncodeToBytes(&te2)
	if err != nil {
		t.Fatal("failed to encode istanbul extra")
	}
	fmt.Println("len:", len(istPayload2))
	fmt.Printf("encode bytes: %x\n", istPayload2)
	///////////////////////////////////////////////////////
	te1Te2 := common.FromHex("0xf90143b89ff843b8419f625663217f82a8765f8a6277bea12e953cd2adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc7970159019f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019a78585408cc797016703885901a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019a78585408cc7970b89ff843b8419f625663217f82a8765f8a6277bea12e953cd2adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc7970159019f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019a78585408cc797016703885901a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019a78585408cc797080") //append(istPayload1[2:],istPayload2[2:]...)
	//fmt.Printf("te1te2:%x\n",te1Te2)
	//te1Te2 = append(common.FromHex("f8c9"), te1Te2...)
	fmt.Println("len:", len(te1Te2))
	fmt.Printf("encode bytes::%x\n", te1Te2)
	var te3 *TestEncode
	err1 := rlp.DecodeBytes(te1Te2, &te3)
	if err1 != nil {
		t.Fatal("decode err", err1)
	}
	//fmt.Printf("d1:%x\n",testbytes(254))
	//fmt.Printf("d2:%x\n",testbytes(25))
	//fmt.Printf("d3:%x\n",testbytes(25))
}

type TestEncode struct {
	D1 []byte
	D2 []byte
	D3 []byte
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (te *TestEncode) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		te.D1,
		te.D2,
		te.D3,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (te *TestEncode) DecodeRLP(s *rlp.Stream) error {
	var testEncode struct {
		D1 []byte
		D2 []byte
		D3 []byte
	}
	if err := s.Decode(&testEncode); err != nil {
		return err
	}
	te.D1 = testEncode.D1
	te.D2 = testEncode.D2
	te.D3 = testEncode.D3
	return nil
}

/////////////////////////////////////////////////////////////////////
type TestEncode1 struct {
	D1 []byte
	D2 []byte
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (te *TestEncode1) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		te.D1,
		te.D2,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (te *TestEncode1) DecodeRLP(s *rlp.Stream) error {
	var testEncode struct {
		D1 []byte
		D2 []byte
		D3 []byte
	}
	if err := s.Decode(&testEncode); err != nil {
		return err
	}
	te.D1 = testEncode.D1
	te.D2 = testEncode.D2
	return nil
}

////////////////////////////////////////////////////////////////////
type TestEncode2 struct {
	D1 []byte
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (te *TestEncode2) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		te.D1,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (te *TestEncode2) DecodeRLP(s *rlp.Stream) error {
	var testEncode struct {
		D1 []byte
	}
	if err := s.Decode(&testEncode); err != nil {
		return err
	}
	te.D1 = testEncode.D1
	return nil
}

func testbytes(len int) []byte {
	token := make([]byte, len)
	rand.Read(token)
	//return token
	return common.FromHex("0xf843b8419f625663217f82a8765f8a6277bea12e953cd2adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc7970159019f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019a78585408cc797016703885901a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019a78585408cc7970")
	//return common.FromHex("0x9f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc797016703885901")
}

/////////////////////////////////////////////////////////////////

type AggSeal struct {
	Bitmap    *big.Int
	Signature []byte
	Round     *big.Int
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (as *AggSeal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		as.Bitmap,
		as.Signature,
		as.Round,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (as *AggSeal) DecodeRLP(s *rlp.Stream) error {
	var aggSeal struct {
		Bitmap    *big.Int
		Signature []byte
		Round     *big.Int
	}

	if err := s.Decode(&aggSeal); err != nil {
		return err
	}
	as.Bitmap = aggSeal.Bitmap
	as.Signature = aggSeal.Signature
	as.Round = aggSeal.Round
	return nil
}

func TestEncodeAgg(t *testing.T) {
	te := AggSeal{
		big.NewInt(0),
		common.FromHex("0x9f625663217f82a8765f8a6277bea12e953cd219adc9ff967c783946f00bbcae615ae014687384abb019b37cfb507fe418a52558a78585408cc797016703885901"),
		big.NewInt(0),
	}
	istPayload, err := rlp.EncodeToBytes(&te)
	if err != nil {
		t.Fatal("failed to encode istanbul extra")
	}
	fmt.Println("len:", len(istPayload))
	fmt.Printf("encode bytes %x\n", istPayload)
}

func TestYode2(t *testing.T) {
	a := big.NewInt(0)
	b := big.NewInt(9)
	a.SetBytes(b.Bytes())
	fmt.Println(a, b)
}

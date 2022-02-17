package bn256_keep_network

import (
	"encoding/hex"
	"encoding/json"
)

func (secretKey PrivateKey) MarshalJSON() ([]byte, error) {
	if secretKey.p == nil {
		return json.Marshal(nil)
	}
	return secretKey.p.MarshalJSON()
}

func (secretKey *PrivateKey) UnmarshalJSON(data []byte) error {
	priv, err := UnmarshalPrivateKey(data)
	*secretKey = priv
	return err
}

func (publicKey PublicKey) MarshalJSON() ([]byte, error) {
	raw := publicKey.Marshal()
	if raw == nil {
		return json.Marshal(raw)
	}
	return json.Marshal(hex.EncodeToString(raw))
}

func (publicKey *PublicKey) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	raw, err := hex.DecodeString(str)
	if err != nil {
		return err
	}
	sig, err := UnmarshalPublicKey(raw)
	*publicKey = sig
	return err
}

func (signature Signature) MarshalJSON() ([]byte, error) {
	raw := signature.Marshal()
	if raw == nil {
		return json.Marshal(raw)
	}
	return json.Marshal(hex.EncodeToString(raw))
}

func (signature *Signature) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	raw, err := hex.DecodeString(str)
	if err != nil {
		return err
	}
	sig, err := UnmarshalSignature(raw)
	*signature = sig
	return err
}

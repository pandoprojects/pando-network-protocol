package keystore

import (
	"github.com/pborman/uuid"

	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/crypto"
)

type Key struct {
	Id         uuid.UUID
	Address    common.Address
	PrivateKey *crypto.PrivateKey
}

func NewKey(privKey *crypto.PrivateKey) *Key {
	Id := uuid.NewRandom()
	return &Key{
		Id:         Id,
		Address:    privKey.PublicKey().Address(),
		PrivateKey: privKey,
	}
}

func (key *Key) Sign(data common.Bytes) (*crypto.Signature, error) {
	sig, err := key.PrivateKey.Sign(data)
	return sig, err
}

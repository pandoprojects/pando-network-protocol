package keystore

import (
	"io"

	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/crypto"
	"github.com/pandoprojects/pando/wallet/types"
)

// Driver abstracts the functionality of the hardware wallet
type Driver interface {
	Status() (string, error)
	Open(device io.ReadWriter, password string) error
	Close() error
	Heartbeat() error
	Derive(path types.DerivationPath) (common.Address, error)
	SignTx(path types.DerivationPath, txrlp common.Bytes) (common.Address, *crypto.Signature, error)
}

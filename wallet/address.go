package wallet

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	bip32 "github.com/btcsuite/btcutil/hdkeychain"

	"bitlox/btcinfo"
	"github.com/btcsuite/btcd/btcec"
)

type Address struct {
	key         *bip32.ExtendedKey
	hash        *btcutil.AddressPubKeyHash
	Chain       uint32
	ChainIndex  uint32
	BalanceInfo *btcinfo.Address
	Unspent     []*btcinfo.Output
}

func (a *Address) Hash() (*btcutil.AddressPubKeyHash, error) {
	if a.hash != nil {
		return a.hash, nil
	}
	hash, err := a.key.Address(&chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	a.hash = hash
	return hash, nil
}

func (a *Address) Address() (string, error) {
	hash, err := a.Hash()
	if err != nil {
		return "", err
	}
	return hash.EncodeAddress(), nil
}

func (a *Address) String() string {
	pub, err := a.Address()
	if err != nil {
		return "Invalid key"
	}
	return pub
}

func (a *Address) ECPubKey() (*btcec.PublicKey, error) {
	return a.key.ECPubKey()
}

func (a *Address) Balance() btcinfo.Satoshi {
	if a.BalanceInfo == nil {
		return 0
	}
	bal := a.BalanceInfo.Balance - a.BalanceInfo.UnconfirmedSent
	if bal < 0 {
		return 0
	}
	return btcinfo.Satoshi(bal)
}

func (a *Address) UnconfirmedBalance() btcinfo.Satoshi {
	if a.BalanceInfo == nil {
		return 0
	}
	bal := a.BalanceInfo.UnconfirmedBalance
	if bal < 0 {
		return 0
	}
	return btcinfo.Satoshi(bal)
}

package wallet

import (
	bip32 "github.com/btcsuite/btcutil/hdkeychain"

	"bitlox/btcinfo"
	"bitlox/logger"
)

const CHAIN_INDEX_RECEIVE uint32 = 0x00
const CHAIN_INDEX_CHANGE uint32 = 0x01

type Wallet struct {
	xpub      []byte
	masterKey *bip32.ExtendedKey
	chains    map[uint32]*bip32.ExtendedKey
	addresses map[uint32]map[uint32]*Address
}

func WalletFromXpub(xpub []byte) *Wallet {
	addresses := map[uint32]map[uint32]*Address{}
	addresses[CHAIN_INDEX_RECEIVE] = make(map[uint32]*Address)
	addresses[CHAIN_INDEX_CHANGE] = make(map[uint32]*Address)

	wallet := &Wallet{
		xpub:      xpub,
		chains:    make(map[uint32]*bip32.ExtendedKey),
		addresses: addresses,
	}
	return wallet
}

func (w *Wallet) Balance() btcinfo.Satoshi {
	total := btcinfo.Satoshi(0)
	for _, chainAddrs := range w.addresses {
		for _, addr := range chainAddrs {
			for _, output := range addr.Unspent {
				total += output.Value
			}
		}
	}
	return total
}

func (w *Wallet) Addresses(chain uint32) []*Address {
	var (
		addrMap map[uint32]*Address
		ok      bool
	)
	if addrMap, ok = w.addresses[chain]; !ok {
		return make([]*Address, 0)
	}
	addrCount := len(addrMap)
	addrs := make([]*Address, addrCount)
	for chainIndex, addr := range addrMap {
		addrs[chainIndex] = addr
	}
	return addrs
}

func (w *Wallet) MasterKey() (*bip32.ExtendedKey, error) {
	if w.masterKey != nil {
		return w.masterKey, nil
	}
	k, err := bip32.NewKeyFromString(string(w.xpub))
	if err != nil {
		return nil, err
	}
	w.masterKey = k
	return k, nil

}

func (w *Wallet) getChain(index uint32) (*bip32.ExtendedKey, error) {
	if ch, ok := w.chains[index]; ok {
		return ch, nil
	}
	master, err := w.MasterKey()
	if err != nil {
		return nil, err
	}
	k, err := master.Child(index)
	if err != nil {
		return nil, err
	}
	w.chains[index] = k
	return k, nil
}

func (w *Wallet) ReceiveChain() (*bip32.ExtendedKey, error) {
	return w.getChain(CHAIN_INDEX_RECEIVE)
}

func (w *Wallet) ChangeChain() (*bip32.ExtendedKey, error) {
	return w.getChain(CHAIN_INDEX_CHANGE)
}

func (w *Wallet) generateAddress(chain, chainIndex uint32) (*Address, error) {
	if k, ok := w.addresses[chain][chainIndex]; ok {
		return k, nil
	}
	ch, err := w.getChain(chain)
	if err != nil {
		return nil, err
	}
	k, err := ch.Child(chainIndex)
	if err != nil {
		return nil, err
	}
	address := &Address{
		key:        k,
		Unspent:    make([]*btcinfo.Output, 0),
		Chain:      chain,
		ChainIndex: chainIndex,
	}
	w.addresses[chain][chainIndex] = address
	return address, nil

}

func (w *Wallet) ReceiveAddress(chainIndex uint32) (*Address, error) {
	return w.generateAddress(CHAIN_INDEX_RECEIVE, chainIndex)
}

func (w *Wallet) LoadBalance() {
	w.loadAllAddresses()
}

func (w *Wallet) loadAllAddresses() {
	for chain := range w.addresses {
		w.loadAddressChain(chain)
	}
}

func (w *Wallet) loadAddressChain(chain uint32) {
	needsMore := true
	chainIndex := uint32(0)
	for needsMore {
		address, err := w.generateAddress(chain, chainIndex)
		if err != nil {
			logger.Debug("Error making address", chain, chainIndex, err)
			chainIndex += 1
			continue
		}

		addrInfo, err := btcinfo.GetAddress(address.String())
		if err != nil {
			logger.Debug("Error getting address info", chain, chainIndex, err)
			chainIndex += 1
			continue
		}

		address.BalanceInfo = addrInfo
		if addrInfo.Received > 0 || addrInfo.UnconfirmedReceived > 0 {
			unspent, err := btcinfo.GetUnspent(address.String())
			if err != nil {
				logger.Debug("Error getting address unspent", chain, chainIndex, err)
			}
			if unspent != nil {
				address.Unspent = unspent
			}
			chainIndex += 1
		} else {
			needsMore = false
		}
	}
}

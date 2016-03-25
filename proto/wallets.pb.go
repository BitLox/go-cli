package proto

import (
	"bytes"
	"fmt"
)

// WalletInfo contains info about a wallet, duh
type WalletInfo struct {
	Number  int32  `protobuf:"varint,1,req,name=wallet_number"`
	Name    []byte `protobuf:"bytes,2,req,name=wallet_name"`
	UUID    []byte `protobuf:"bytes,3,req,name=wallet_uuid"`
	Version int32  `protobuf:"varint,4,req,name=version"`
}

func (m *WalletInfo) Reset() {
	m = &WalletInfo{}
}

func (m *WalletInfo) String() string {
	return fmt.Sprintf("[%d] %s", m.Number, m.NameString())
}

func (m *WalletInfo) ProtoMessage() {}

func (m *WalletInfo) NameString() string {
	nameBytes := bytes.TrimRight(m.Name, string([]byte{0x00}))
	nameBytes = bytes.TrimSpace(nameBytes)
	return string(nameBytes)
}

// Wallets holds WalletInfo structs when calling list wallets command
type Wallets struct {
	Wallets []*WalletInfo `protobuf:"bytes,1,rep,name=wallet_info"`
}

func (m *Wallets) Reset() {
	m = &Wallets{}
}

func (m *Wallets) String() string {
	return fmt.Sprintf("%d wallets", len(m.Wallets))
}

func (m *Wallets) ProtoMessage() {}

// LoadWallet is an outgoing message to load a wallet on the device for scanning
type LoadWallet struct {
	Number uint32 `protobuf:"varint,1,opt,name=wallet_number"`
}

// Wallet contains the xpub info received from the bitlox
type CurrentWalletXPUB struct {
	Xpub []byte `protobuf:"bytes,1,req,name=xpub"`
}

func (m *CurrentWalletXPUB) Reset() {
	m = &CurrentWalletXPUB{}
}

func (m *CurrentWalletXPUB) String() string {
	return fmt.Sprintf("currentWalletXPUB: %s", m.Xpub)
}

func (m *CurrentWalletXPUB) ProtoMessage() {}

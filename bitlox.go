package bitlox

import (
	gohid "github.com/GeertJohan/go.hid"
	"github.com/golang/protobuf/proto"

	"bitlox/hid"
	"bitlox/logger"
	models "bitlox/proto"
	"bitlox/wallet"

	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"regexp"
)

var ERR_UNRECOGNIZED_RETURN = errors.New("Unrecogized command returned from device")
var ERR_MESSAGE_TOO_LONG = errors.New("Message too long")
var ERR_INVALID_SIG_LENGTH = errors.New("Invalid signature length")
var ERR_INVALID_SIG = errors.New("Invalid signature")

const MAX_MESSAGE_LENGTH = 0xffff

type Device struct {
	dev *gohid.Device
}

func (d *Device) Close() {
	d.dev.Close()
}

func GetDevice() (*Device, error) {
	dev, err := hid.GetDevice(1003, 8271)
	if err != nil {
		return nil, err
	}
	return &Device{dev: dev}, nil
}

func GetWallets(d *Device) ([]*models.WalletInfo, error) {
	err := hid.Write(d.dev, hid.COMMAND_LIST_WALLETS)
	if err != nil {
		return nil, err
	}

	res := new(bytes.Buffer)
	err = hid.Read(d.dev, res)
	if err != nil {
		return nil, err
	}

	_, _, payload := hid.ParseResponse(res)

	wallets := &models.Wallets{}

	err = proto.Unmarshal(payload, wallets)
	if err != nil {
		return nil, err
	}

	return wallets.Wallets, nil
}

func LoadWallet(d *Device, number byte) error {

	cmd := append(hid.PREFIX_LOAD_WALLET, number)

	err := hid.Write(d.dev, cmd)
	if err != nil {
		return err
	}

	res := new(bytes.Buffer)
	err = hid.Read(d.dev, res)
	if err != nil {
		return err
	}

	resCmd, _, payload := hid.ParseResponse(res)

	if resCmd == hid.RESPONSE_SUCCESS {
		return nil
	} else if resCmd == hid.RESPONSE_ERROR {
		failure := &models.Failure{}

		err = proto.Unmarshal(payload, failure)
		if err != nil {
			return err
		}
		return failure
	} else {
		return ERR_UNRECOGNIZED_RETURN
	}
}

func ScanWallet(d *Device) ([]byte, error) {
	err := hid.Write(d.dev, hid.COMMAND_SCAN_WALLET)
	if err != nil {
		return nil, err
	}

	res := new(bytes.Buffer)
	err = hid.Read(d.dev, res)
	if err != nil {
		return nil, err
	}

	_, _, payload := hid.ParseResponse(res)

	xpub := &models.CurrentWalletXPUB{}

	err = proto.Unmarshal(payload, xpub)
	if err != nil {
		return nil, err
	}
	logger.Debug("scan got", xpub)

	return xpub.Xpub, nil
}

func makeAddressHandle(ch uint32, chainIndex uint32) []byte {
	b := []byte{10}
	chain := make([]byte, 4)
	binary.LittleEndian.PutUint32(chain, ch)
	chain = bytes.TrimRight(chain, string(0))
	if len(chain) == 0 {
		chain = []byte{0}
	}
	rootAndChain := append([]byte{8, 0, 16}, chain...)
	rootAndChain = append(rootAndChain, 24)
	index := make([]byte, 4)
	binary.LittleEndian.PutUint32(index, chainIndex)
	index = bytes.TrimRight(index, string(0))
	if len(index) == 0 {
		index = []byte{0}
	}
	data := append(rootAndChain, index...)
	b = append(b, byte(len(data)))
	b = append(b, data...)
	return b
}

var messagePrefix = []byte("Bitcoin Signed Message:\n")
var messageTrim = regexp.MustCompile(`(^[\s\n]+|[\s\n]+$)`)

func getMessageSize(msg []byte) []byte {
	l := len(msg)
	if l < 0xfd { // < 254
		return []byte{byte(l)}
	} else if l < MAX_MESSAGE_LENGTH {
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(l))
		return append([]byte{0xfd}, b...)
	} else {
		panic(ERR_MESSAGE_TOO_LONG)
	}
}

func SignMessage(d *Device, address *wallet.Address, message []byte) ([]byte, error) {

	message = messageTrim.ReplaceAll(message, []byte{})

	if len(message) >= MAX_MESSAGE_LENGTH {
		return nil, ERR_MESSAGE_TOO_LONG
	}

	message = append(getMessageSize(message), message...)
	message = append(messagePrefix, message...)
	message = append(getMessageSize(messagePrefix), message...)

	m := &models.SignMessage{
		Handle:  makeAddressHandle(wallet.CHAIN_INDEX_RECEIVE, address.ChainIndex),
		Message: message,
	}

	mBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	err = hid.WriteVariable(d.dev, hid.PREFIX_SIGN_MESSAGE, mBytes)
	if err != nil {
		return nil, err
	}

	res := new(bytes.Buffer)
	err = hid.Read(d.dev, res)
	if err != nil {
		return nil, err
	}

	resCmd, _, payload := hid.ParseResponse(res)

	// check if the command is "please ACK", if so, send the ack,
	// read, and then parse *that* response
	if resCmd != hid.RESPONSE_PLEASE_ACK {
		return nil, ERR_UNRECOGNIZED_RETURN
	}

	logger.Debug("sending ACK")

	err = hid.Write(d.dev, hid.COMMAND_ACK)
	if err != nil {
		return nil, err
	}

	res = new(bytes.Buffer)
	err = hid.Read(d.dev, res)
	if err != nil {
		return nil, err
	}

	resCmd, _, payload = hid.ParseResponse(res)

	sig := &models.SignatureComplete{}

	err = proto.Unmarshal(payload, sig)
	if err != nil {
		return nil, err
	}

	if resCmd == hid.RESPONSE_MESSAGE_SIGNATURE {
		return processSignedMessage2(address, message, sig.Signature)
	} else if resCmd == hid.RESPONSE_ERROR {
		failure := &models.Failure{}

		err = proto.Unmarshal(payload, failure)
		if err != nil {
			return nil, err
		}
		return nil, failure
	} else {
		return nil, ERR_UNRECOGNIZED_RETURN
	}

}

func doubleSha(b []byte) []byte {
	round1 := sha256.Sum256(b)
	arr := sha256.Sum256(round1[0:32])
	return arr[0:32]
}

func processSignedMessage2(address *wallet.Address, message, derSig []byte) ([]byte, error) {
	sig, err := btcec.ParseDERSignature(derSig, btcec.S256())
	if err != nil {
		logger.Error("sig parse error", err)
		return nil, err
	}

	r := sig.R.Bytes()
	s := sig.S.Bytes()

	derRLen := derSig[3]
	derSLen := derSig[5+derRLen]
	logger.Debug(derSig)
	logger.Debugf("R: der: %d bytes: %d", derRLen, len(r))
	logger.Debugf("S: der: %d bytes: %d", derSLen, len(s))
	signature := []byte{0}
	signature = append(signature, r...)
	signature = append(signature, s...)

	for nV := 31; nV < 34; nV++ {
		signature[0] = byte(nV)
		b64len := base64.StdEncoding.EncodedLen(len(signature))
		b64 := make([]byte, b64len)
		base64.StdEncoding.Encode(b64, signature)
		logger.Debug(nV, string(b64))
		// Validate the signature - this just shows that it was valid at all.
		// we will compare it with the key next.
		expectedMessageHash := doubleSha(append(messagePrefix, message...))

		pk, wasCompressed, err := btcec.RecoverCompact(btcec.S256(), signature, expectedMessageHash)
		logger.Debug("wasCompressed", wasCompressed)
		if err != nil {
			logger.Debug("nV", nV, err)
			continue
		}

		var serializedPK []byte
		serializedPK = pk.SerializeUncompressed()
		addr, err := btcutil.NewAddressPubKey(serializedPK, &chaincfg.MainNetParams)
		if err != nil {
			logger.Debug("nV", nV, err)
			continue
		}

		// Return boolean if addresses match.
		logger.Debugf("%s == %s\n", addr.EncodeAddress(), address.String())
		if addr.EncodeAddress() == address.String() {
			return b64, nil
		} else {
			continue
		}
	}
	return nil, ERR_INVALID_SIG
}

func processSignedMessage(address *wallet.Address, message, signature []byte) ([]byte, error) {
	sig, err := btcec.ParseDERSignature(signature, btcec.S256())
	if err != nil {
		logger.Error("sig parse error", err)
		return nil, err
	}
	key, err := address.ECPubKey()
	if err != nil {
		logger.Error("get pub error", err)
		return nil, err
	}
	if ok := sig.Verify(doubleSha(message), key); !ok {
		return nil, ERR_INVALID_SIG
	}
	return []byte(base64.StdEncoding.EncodeToString(signature)), nil
}

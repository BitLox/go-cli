package btcinfo

import (
	"github.com/btcsuite/btcutil"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"bitlox/logger"
)

const TOSHI_URL = "https://bitcoin.toshi.io/api/v0/addresses"

var UnitBTC = btcutil.AmountBTC
var UnitBits = btcutil.AmountMicroBTC
var UnitSatoshi = btcutil.AmountSatoshi
var UnitMBTC = btcutil.AmountMilliBTC

type Satoshi float64

func (s Satoshi) ToBitcoin() float64 {
	return btcutil.Amount(s).ToBTC()
}

func (s Satoshi) ToBits() float64 {
	return btcutil.Amount(s).ToUnit(btcutil.AmountMicroBTC)
}

func (s Satoshi) ToBitcoinString() string {
	return btcutil.Amount(s).Format(btcutil.AmountBTC)
}

func (s Satoshi) ToFullBitcoinString() string {
	return fmt.Sprintf("%.8f BTC", btcutil.Amount(s).ToBTC())
}

func (s Satoshi) ToBitString() string {
	str := btcutil.Amount(s).Format(btcutil.AmountMicroBTC)
	return strings.Replace(str, "μBTC", "bits", -1)
}

func (s Satoshi) ToUnit(unit btcutil.AmountUnit) float64 {
	return btcutil.Amount(s).ToUnit(unit)
}

func (s Satoshi) Format(unit btcutil.AmountUnit) string {
	str := btcutil.Amount(s).Format(unit)
	return strings.Replace(str, "μBTC", "bits", -1)
}

type Output struct {
	HashStr   string   `json:"transaction_hash"`
	Value     Satoshi  `json:"amount"`
	ScriptStr string   `json:"script_hex"`
	Number    int      `json:"output_index"`
	Addresses []string `json:"addresses"`
}

func (o *Output) Hash() ([]byte, error) {
	hash, err := hex.DecodeString(o.HashStr)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (o *Output) Script() ([]byte, error) {
	script, err := hex.DecodeString(o.ScriptStr)
	if err != nil {
		return nil, err
	}
	return script, nil
}

type Address struct {
	Received            Satoshi `json:"received"`
	Balance             Satoshi `json:"balance"`
	UnconfirmedSent     Satoshi `json:"unconfirmed_sent"`
	UnconfirmedReceived Satoshi `json:"unconfirmed_received"`
	UnconfirmedBalance  Satoshi `json:"unconfirmed_balance"`
}

func doReq(path string, resItem interface{}) error {
	resp, err := http.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, resItem)
	if err != nil {
		logger.Debug("toshi JSON", string(body[44:52]), string(body))
		return err
	}
	return nil
}

func GetAddress(pubkey string) (*Address, error) {
	addr := &Address{}
	err := doReq(TOSHI_URL+"/"+pubkey, addr)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func GetUnspent(pubkey string) ([]*Output, error) {
	unspent := make([]*Output, 0)
	err := doReq(TOSHI_URL+"/"+pubkey+"/unspent_outputs", &unspent)
	if err != nil {
		return nil, err
	}
	if len(unspent) > 0 {
		logger.Debugf("%#v\n", unspent[0])
	}
	return unspent, nil
}

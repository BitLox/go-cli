package main

import (
	"github.com/spf13/cobra"

	"bitlox"
	"bitlox/btcinfo"
	"bitlox/logger"
	"bitlox/wallet"
	"strconv"
)

var UNIT = btcinfo.UnitBTC

// command line flags
var (
	walletNumber int
	verbose      bool
	debug        bool
	unit         string
	address      string
	chainIndex   int
)

// global vars to store things
var w *wallet.Wallet
var dev *bitlox.Device

func main() {

	appCmd := &cobra.Command{
		Use: "bitlox-cli",
		Run: func(cmd *cobra.Command, args []string) {
			walletList()
		},
		PersistentPreRun: appPreRun,
	}

	appCmd.PersistentFlags().StringVarP(&unit, "unit", "u", "btc", "Specify the unit for displaying values")
	appCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	appCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Show debug messages (very verbose)")

	walletCmd := &cobra.Command{
		Use:   "wallet <wallet number>",
		Short: "Show balance of the specified wallet",
		Long: `Show balance of the specified wallet

Verbose output will show all addresses for the recieve and change chains and their individual balances`,
		PersistentPreRun: walletPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			balance()
		},
	}

	balanceCmd := &cobra.Command{
		Use:   "balance",
		Short: "Show balance of the specified wallet",
		Long: `Show balance of the specified wallet

Verbose output will show all addresses for the recieve and change chains and their individual balances`,
		Run: func(cmd *cobra.Command, args []string) {
			balance()
		},
	}

	addressesCmd := &cobra.Command{
		Use:   "addresses",
		Short: "Show addresses of the specified wallet",
		Long: `Show addresses of the specified wallet

Verbose output will show the balance of each address.`,
		Run: func(cmd *cobra.Command, args []string) {
			if walletNumber < 0 {
				logger.Fatal("Wallet number is required for the addresses command")
			}
			addresses()
		},
	}

	signCmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign a message",
		Long: `Sign a message

Sign a message with the specified address or address chain index (receive addresses only).`,
		PreRun: func(cmd *cobra.Command, args []string) {
			logger.Log("Loading addresses")
			w.LoadBalance()
			if address != "" {
				logger.Log("Finding chain index for", address)
				for index, addr := range w.Addresses(wallet.CHAIN_INDEX_RECEIVE) {
					if chainIndex >= 0 {
						continue
					}
					pub, err := addr.Address()
					if err != nil {
						continue
					}
					if pub == address {
						chainIndex = index
					}
				}
				if chainIndex < 0 {
					logger.Fatalf("Address %s not found on device\n", address)
				}
				logger.Log("Found chain index", chainIndex)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if walletNumber < 0 {
				logger.Fatal("Wallet number is required for the sign command")
			}
			sign([]byte(args[1]))
		},
	}

	signCmd.Flags().IntVarP(&chainIndex, "chain-index", "i", -1, "Specify the address chain index")
	signCmd.Flags().StringVarP(&address, "address", "a", "", "Specify the address")

	walletCmd.AddCommand(balanceCmd, addressesCmd, signCmd)

	appCmd.AddCommand(walletCmd)
	appCmd.Execute()

}

func getDevice() {
	logger.Log("Getting device connection")
	var err error
	dev, err = bitlox.GetDevice()
	if err != nil {
		logger.Fatal(err)
	}
}

func walletList() {

	wallets, err := bitlox.GetWallets(dev)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Log("\nWALLETS")
	for _, w := range wallets {
		logger.Log(w)
	}

}

func balance() {

	logger.Log("Loading balance")
	w.LoadBalance()

	if verbose {
		logger.Log("\nRECEIVE CHAIN")
		for chainIndex, address := range w.Addresses(wallet.CHAIN_INDEX_RECEIVE) {
			logger.Logf("%-3d %-35s %16s %16s unconfirmed\n",
				chainIndex, address, address.Balance().Format(UNIT), address.UnconfirmedBalance().Format(UNIT))
		}
		logger.Log("\nCHANGE CHAIN")
		for chainIndex, address := range w.Addresses(wallet.CHAIN_INDEX_CHANGE) {
			logger.Logf("%-3d %-35s %16s %16s unconfirmed\n",
				chainIndex, address, address.Balance().Format(UNIT), address.UnconfirmedBalance().Format(UNIT))
		}
	}

	logger.Logf("\nBALANCE\n%s\n", w.Balance().Format(UNIT))

}

func addresses() {

	logger.Log("Loading addresses")
	w.LoadBalance()

	for chainIndex, address := range w.Addresses(wallet.CHAIN_INDEX_RECEIVE) {
		logger.Logf("%-3d %-35s", chainIndex, address)
		if verbose {
			logger.Logf("%16s %16s unconfirmed\n", address.Balance().Format(UNIT), address.UnconfirmedBalance().Format(UNIT))
		} else {
			logger.Logf("\n")
		}
	}

}

func sign(message []byte) {
	logger.Log("Signing. Check Device")
	addresses := w.Addresses(wallet.CHAIN_INDEX_RECEIVE)
	address := addresses[chainIndex]
	sig, err := bitlox.SignMessage(dev, address, message)
	if err != nil {
		logger.Fatal(err, string(sig))
	}

	logger.Log("\nSIGNATURE")
	logger.Logf("%s\n\n", sig)
	logger.Logf(`bitcoin-cli verifymessage %s "%s" "%s"`+"\n", address, sig, message)
}

func appPreRun(cmd *cobra.Command, args []string) {
	if debug {
		logger.EnableDebug()
	}
	switch unit {
	case "bitcoin", "btc", "BTC":
		UNIT = btcinfo.UnitBTC
	case "mbtc", "mBTC":
		UNIT = btcinfo.UnitMBTC
	case "ubtc", "bits":
		UNIT = btcinfo.UnitBits
	case "satoshi":
		UNIT = btcinfo.UnitSatoshi
	}
	getDevice()
}

func walletPreRun(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		logger.Fatal("Wallet number is required for the balance command")
	}
	walletNumber, err := strconv.Atoi(args[0])
	if err != nil {
		logger.Fatal("Invalid wallet number")
	}
	if walletNumber < 0 {
		logger.Fatal("Invalid wallet number")
	}
	if cmd.Use == "sign" {
		if chainIndex < 0 && address == "" {
			logger.Fatal("You must supply either --chain-index or --address to sign a message")
		}
		if chainIndex >= 0 && address != "" {
			logger.Fatal("You cannot supply both --chain-index or --address to sign a message")
		}
		if len(args) < 2 {
			logger.Fatal("Missing message to sign")
		}
	}
	appPreRun(cmd, args)

	// get the wallet in question
	logger.Log("Loading wallet info")
	err = bitlox.LoadWallet(dev, byte(walletNumber))
	if err != nil {
		logger.Fatal(err)
	}

	logger.Log("Getting public key")
	xpub, err := bitlox.ScanWallet(dev)
	if err != nil {
		logger.Fatal(err)
	}

	w = wallet.WalletFromXpub(xpub)

}

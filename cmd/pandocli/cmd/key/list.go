package key

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/pandoprojects/pando/cmd/pandocli/cmd/utils"
	"github.com/pandoprojects/pando/wallet"
	wtypes "github.com/pandoprojects/pando/wallet/types"
)

// listCmd lists all the stored keys
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all keys",
	Long:    `List all keys.`,
	Example: "pandocli key list",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := cmd.Flag("config").Value.String()
		wallet, err := wallet.OpenWallet(cfgPath, wtypes.WalletTypeSoft, true)
		if err != nil {
			utils.Error("Failed to open wallet: %v\n", err)
		}

		keyAddresses, err := wallet.List()
		if err != nil {
			utils.Error("Failed to list keys: %v\n", err)
		}

		for _, keyAddress := range keyAddresses {
			fmt.Printf("%s\n", keyAddress.Hex())
		}
	},
}

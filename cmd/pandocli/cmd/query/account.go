package query

import (
	"encoding/json"
	"fmt"

	"github.com/pandoprojects/pando/cmd/pandocli/cmd/utils"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/rpc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	rpcc "github.com/ybbus/jsonrpc"
)

// accountCmd represents the account command.
// Example:
//		pandocli query account --address=0x2E833968E5bB786Ae419c4d13189fB081Cc43bab
var accountCmd = &cobra.Command{
	Use:     "account",
	Short:   "Get account status",
	Long:    `Get account status.`,
	Example: `pandocli query account --address=0x2E833968E5bB786Ae419c4d13189fB081Cc43bab`,
	Run:     doAccountCmd,
}

func doAccountCmd(cmd *cobra.Command, args []string) {
	client := rpcc.NewRPCClient(viper.GetString(utils.CfgRemoteRPCEndpoint))

	res, err := client.Call("pando.GetAccount", rpc.GetAccountArgs{
		Address: addressFlag,
		Height:  common.JSONUint64(heightFlag),
		Preview: previewFlag})
	if err != nil {
		utils.Error("Failed to get account details: %v\n", err)
	}
	if res.Error != nil {
		utils.Error("Failed to get account details: %v\n", res.Error)
	}
	json, err := json.MarshalIndent(res.Result, "", "    ")
	if err != nil {
		utils.Error("Failed to parse server response: %v\n%v\n", err, string(json))
	}
	fmt.Println(string(json))
}

func init() {
	accountCmd.Flags().StringVar(&addressFlag, "address", "", "Address of the account")
	accountCmd.Flags().Uint64Var(&heightFlag, "height", uint64(0), "height of the block")
	accountCmd.Flags().BoolVar(&previewFlag, "preview", false, "Preview account balance from the screened view")
	accountCmd.MarkFlagRequired("address")
}

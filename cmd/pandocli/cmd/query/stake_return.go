package query

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/pandoprojects/pando/cmd/pandocli/cmd/utils"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/rpc"

	rpcc "github.com/ybbus/jsonrpc"
)

// stakeReturnsCmd represents the stake_return command.
// Example:
//		pandocli query stake_returns --height=10
var stakeReturnsCmd = &cobra.Command{
	Use:     "stake_returns",
	Short:   "Get stake returns",
	Example: `pandocli query stake_returns, pandocli query stake_returns --height=800`,
	Run:     doStakeReturnsCmd,
}

func doStakeReturnsCmd(cmd *cobra.Command, args []string) {
	client := rpcc.NewRPCClient(viper.GetString(utils.CfgRemoteRPCEndpoint))

	purpose := purposeFlag
	if purpose != 2 {
		fmt.Println("Only support querying stake return for rametronenterprise (purpose=2) for now")
		return
	}

	height := heightFlag
	var res *rpcc.RPCResponse
	var err error
	if height == 0 {
		res, err = client.Call("pando.GetAllPendingrametronenterpriseStakeReturns", rpc.GetAllPendingRametronenterpriseStakeReturnsArgs{})
	} else {
		res, err = client.Call("pando.GetrametronenterpriseStakeReturnsByHeight", rpc.GetRametronenterpriseStakeReturnsByHeightArgs{Height: common.JSONUint64(height)})
	}
	if err != nil {
		utils.Error("Failed to get stake returns: %v\n", err)
	}
	if res.Error != nil {
		utils.Error("Failed to get stake returns: %v\n", res.Error)
	}
	json, err := json.MarshalIndent(res.Result, "", "    ")
	if err != nil {
		utils.Error("Failed to parse server response: %v\n%s\n", err, string(json))
	}
	fmt.Println(string(json))
}

func init() {
	stakeReturnsCmd.Flags().Uint8Var(&purposeFlag, "purpose", uint8(2), "purpose of the stake return query, validator_node=0, guardian_node=1, rametronenterprise=2")
	stakeReturnsCmd.Flags().Uint64Var(&heightFlag, "height", uint64(0), "height of the block, if height=0 the command returns all the pending stake returns")
	//stakeReturnsCmd.MarkFlagRequired("height")
}

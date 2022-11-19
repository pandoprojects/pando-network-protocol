package query

import (
	"encoding/json"
	"fmt"

	"github.com/pandoprojects/pando/cmd/pandocli/cmd/utils"
	"github.com/pandoprojects/pando/rpc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	rpcc "github.com/ybbus/jsonrpc"
)

// metatronCmd retreves metatron related information from Pando server.
// Example:
//		pandocli query metatron
var metatronCmd = &cobra.Command{
	Use:     "metatron",
	Short:   "Get metatron info",
	Long:    `Get metatron status.`,
	Example: `pandocli query metatron`,
	Run:     doMetatronCmd,
}

type MetatronResult struct {
	Address   string
	BlsPubkey string
	BlsPop    string
	Signature string
	Summary   string
}

func doMetatronCmd(cmd *cobra.Command, args []string) {
	client := rpcc.NewRPCClient(viper.GetString(utils.CfgRemoteRPCEndpoint))

	res, err := client.Call("pando.GetGuardianInfo", rpc.GetGuardianInfoArgs{})
	if err != nil {
		utils.Error("Failed to get metatron info: %v\n", err)
	}
	if res.Error != nil {
		utils.Error("Failed to get metatron info: %v\n", res.Error)
	}
	result := res.Result.(map[string]interface{})
	address, ok := result["Address"].(string)
	if !ok {
		json, err := json.MarshalIndent(res.Result, "", "    ")
		utils.Error("Failed to parse server response: %v\n%v\n", err, string(json))
	}
	blsPubkey, ok := result["BLSPubkey"].(string)
	if !ok {
		json, err := json.MarshalIndent(res.Result, "", "    ")
		utils.Error("Failed to parse server response: %v\n%v\n", err, string(json))
	}
	blsPop, ok := result["BLSPop"].(string)
	if !ok {
		json, err := json.MarshalIndent(res.Result, "", "    ")
		utils.Error("Failed to parse server response: %v\n%v\n", err, string(json))
	}
	sig, ok := result["Signature"].(string)
	if !ok {
		json, err := json.MarshalIndent(res.Result, "", "    ")
		utils.Error("Failed to parse server response: %v\n%v\n", err, string(json))
	}
	output := &MetatronResult{
		Address:   address,
		BlsPubkey: blsPubkey,
		BlsPop:    blsPop,
		Signature: sig,
	}
	output.Summary = address + blsPubkey + blsPop + sig
	json, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		utils.Error("Failed to parse server response: %v\n%v\n", err, string(json))
	}
	fmt.Println(string(json))
}

func init() {}

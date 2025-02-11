//Package cmd provides all functions related to command line
package cmd

import (
	"math/big"
	"razor/core"
	"razor/core/types"
	"razor/logger"
	"razor/pkg/bindings"
	"razor/utils"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/spf13/pflag"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "transfer razors from your account to others' account",
	Long: `transfer command allows user to transfer their razors to another account if they've not staked those razors

Example:
  ./razor transfer --value 100 --to 0x91b1E6488307450f4c0442a1c35Bc314A505293e --from 0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c`,
	Run: initialiseTransfer,
}

//This function initialises the ExecuteTransfer function
func initialiseTransfer(cmd *cobra.Command, args []string) {
	cmdUtils.ExecuteTransfer(cmd.Flags())
}

//This function sets the flag appropriately and executes the Transfer function
func (*UtilsStruct) ExecuteTransfer(flagSet *pflag.FlagSet) {
	razorUtils.AssignLogFile(flagSet)
	fromAddress, err := flagSetUtils.GetStringFrom(flagSet)
	utils.CheckError("Error in getting fromAddress: ", err)
	logger.Address = fromAddress

	config, err := cmdUtils.GetConfigData()
	utils.CheckError("Error in getting config: ", err)
	password := razorUtils.AssignPassword()

	toAddress, err := flagSetUtils.GetStringTo(flagSet)
	utils.CheckError("Error in getting toAddress: ", err)

	client := razorUtils.ConnectToClient(config.Provider)

	balance, err := razorUtils.FetchBalance(client, fromAddress)
	utils.CheckError("Error in fetching razor balance: ", err)

	valueInWei, err := cmdUtils.AssignAmountInWei(flagSet)
	utils.CheckError("Error in getting amount: ", err)

	transferInput := types.TransferInput{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Password:    password,
		ValueInWei:  valueInWei,
		Balance:     balance,
	}

	txn, err := cmdUtils.Transfer(client, config, transferInput)
	utils.CheckError("Transfer error: ", err)
	log.Info("Transaction Hash: ", txn)
	err = razorUtils.WaitForBlockCompletion(client, txn.Hex())
	utils.CheckError("Error in WaitForBlockCompletion for transfer: ", err)
}

//This function transfers the razors from your account to others account
func (*UtilsStruct) Transfer(client *ethclient.Client, config types.Configurations, transferInput types.TransferInput) (common.Hash, error) {

	razorUtils.CheckAmountAndBalance(transferInput.ValueInWei, transferInput.Balance)

	txnOpts := razorUtils.GetTxnOpts(types.TransactionOptions{
		Client:          client,
		Password:        transferInput.Password,
		AccountAddress:  transferInput.FromAddress,
		ChainId:         big.NewInt(config.ChainId),
		Config:          config,
		ContractAddress: core.RAZORAddress,
		MethodName:      "transfer",
		Parameters:      []interface{}{common.HexToAddress(transferInput.ToAddress), transferInput.ValueInWei},
		ABI:             bindings.RAZORMetaData.ABI,
	})
	log.Infof("Transferring %g tokens from %s to %s", razorUtils.GetAmountInDecimal(transferInput.ValueInWei), transferInput.FromAddress, transferInput.ToAddress)

	txn, err := tokenManagerUtils.Transfer(client, txnOpts, common.HexToAddress(transferInput.ToAddress), transferInput.ValueInWei)
	if err != nil {
		log.Errorf("Error in transferring tokens ")
		return core.NilHash, err
	}

	return transactionUtils.Hash(txn), err
}

func init() {
	rootCmd.AddCommand(transferCmd)
	var (
		Amount   string
		From     string
		To       string
		WeiRazor bool
	)

	transferCmd.Flags().StringVarP(&Amount, "value", "v", "0", "value to transfer")
	transferCmd.Flags().StringVarP(&From, "from", "", "", "transfer from")
	transferCmd.Flags().StringVarP(&To, "to", "", "", "transfer to")
	transferCmd.Flags().BoolVarP(&WeiRazor, "weiRazor", "", false, "value can be passed in wei")

	amountErr := transferCmd.MarkFlagRequired("value")
	utils.CheckError("Value error: ", amountErr)
	fromErr := transferCmd.MarkFlagRequired("from")
	utils.CheckError("From address error: ", fromErr)
	toErr := transferCmd.MarkFlagRequired("to")
	utils.CheckError("To address error: ", toErr)

}

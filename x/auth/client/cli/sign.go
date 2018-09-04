package cli

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

const (
	flagOverwriteSigs = "overwrite"
	flagPrintSigs     = "print-sigs"
)

// GetSignCommand returns the sign command
func GetSignCommand(codec *amino.Codec, decoder auth.AccountDecoder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign <file>",
		Short: "Sign transactions",
		Long: `Sign transactions created with the --generate-only flag.
Read a transaction from <file>, sign it, and print its JSON encoding.`,
		RunE: makeSignCmd(codec, decoder),
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String(client.FlagName, "", "Name of private key with which to sign")
	cmd.Flags().Bool(flagOverwriteSigs, false, "Overwrite the signatures that are already attached to the transaction")
	cmd.Flags().Bool(flagPrintSigs, false, "Print the addresses that must sign the transaction and those who have already signed it, then exit")
	return cmd
}

func makeSignCmd(cdc *amino.Codec, decoder auth.AccountDecoder) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		stdTx, err := readAndUnmarshalStdTx(cdc, args[0])
		if err != nil {
			return
		}

		if viper.GetBool(flagPrintSigs) {
			printSignatures(stdTx)
			return nil
		}

		name := viper.GetString(client.FlagName)
		keybase, err := keys.GetKeyBase()
		if err != nil {
			return
		}
		info, err := keybase.Get(name)
		if err != nil {
			return
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(decoder)
		acc, err := cliCtx.GetAccount(sdk.AccAddress(info.GetPubKey().Address()))
		if err != nil {
			return err
		}

		passphrase, err := keys.GetPassphrase(name)
		if err != nil {
			return err
		}
		newTx, err := signStdTx(stdTx, name, passphrase, acc)
		if err != nil {
			return err
		}
		json, err := cdc.MarshalJSON(newTx)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", json)
		return
	}
}

func signStdTx(stdTx auth.StdTx, name, passphrase string, acc auth.Account) (signedStdTx auth.StdTx, err error) {
	stdSignature, err := authctx.MakeSignature(name, passphrase, auth.StdSignMsg{
		ChainID:       viper.GetString(client.FlagChainID),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	})
	if err != nil {
		return
	}

	signedStdTx = authctx.SignStdTx(stdTx, stdSignature, viper.GetBool(flagOverwriteSigs))
	return
}

func printSignatures(stdTx auth.StdTx) {
	fmt.Println("Signers:")
	for i, signer := range stdTx.GetSigners() {
		fmt.Printf(" %v: %v\n", i, signer.String())
	}
	fmt.Println("")
	fmt.Println("Signatures:")
	for i, sig := range stdTx.GetSignatures() {
		fmt.Printf(" %v: %v\n", i, sdk.AccAddress(sig.Address()).String())
	}
	return
}

func readAndUnmarshalStdTx(cdc *amino.Codec, filename string) (stdTx auth.StdTx, err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if err = cdc.UnmarshalJSON(bytes, &stdTx); err != nil {
		return
	}
	return
}

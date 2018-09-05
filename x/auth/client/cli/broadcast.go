package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

// GetSignCommand returns the sign command
func GetBroadcastCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadcast <file>",
		Short: "Broadcast transactions generated offline",
		Long: `Broadcast transactions created with the --generate-only flag and signed with the sign command.
Read a transaction from <file> and broadcast it to a node.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(codec).WithLogger(os.Stdout)
			stdTx, err := readAndUnmarshalStdTx(cliCtx.Codec, args[0])
			if err != nil {
				return
			}
			txBytes, err := cliCtx.Codec.MarshalBinary(stdTx)
			if err != nil {
				return
			}
			return cliCtx.EnsureBroadcastTx(txBytes)
		},
	}
	return cmd
}

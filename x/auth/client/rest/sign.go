package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
)

type signBody struct {
	Tx               auth.StdTx `json:"tx"`
	LocalAccountName string     `json:"name"`
	Password         string     `json:"password"`
	ChainID          string     `json:"chain_id"`
	AccountNumber    int64      `json:"account_number"`
	Sequence         int64      `json:"sequence"`
}

// sign tx REST handler
func SignTxRequestHandlerFn(cdc *wire.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m signBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		sig, err := authctx.MakeSignature(m.LocalAccountName, m.Password, auth.StdSignMsg{
			ChainID:       m.ChainID,
			AccountNumber: m.AccountNumber,
			Sequence:      m.Sequence,
			Fee:           m.Tx.Fee,
			Msgs:          m.Tx.GetMsgs(),
			Memo:          m.Tx.GetMemo(),
		})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		output, err := wire.MarshalJSONIndent(cdc, authctx.SignStdTx(m.Tx, sig, false))
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(output)
	}
}

package keeper

import (
	"encoding/json"
	"reflect"
	"strconv"

	sdk "github.com/line/lbm-sdk/types"
	sdkerrors "github.com/line/lbm-sdk/types/errors"
	abci "github.com/line/ostracon/abci/types"

	"github.com/line/wasmd/x/wasm/types"
)

const (
	QueryListContractByCode = "list-contracts-by-code"
	QueryGetContract        = "contract-info"
	QueryGetContractState   = "contract-state"
	QueryGetCode            = "code"
	QueryListCode           = "list-code"
	QueryContractHistory    = "contract-history"
	QueryInactiveContracts  = "inactive-contracts"
	QueryIsInactiveContract = "inactive-contract"
)

const (
	QueryMethodContractStateSmart = "smart"
	QueryMethodContractStateAll   = "all"
	QueryMethodContractStateRaw   = "raw"
)

// NewLegacyQuerier creates a new querier
func NewLegacyQuerier(keeper types.ViewKeeper, gasLimit sdk.Gas) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			rsp interface{}
			err error
		)
		switch path[0] {
		case QueryGetContract:
			addr, addrErr := sdk.AccAddressFromBech32(path[1])
			if addrErr != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, addrErr.Error())
			}
			//nolint:staticcheck
			rsp, err = queryContractInfo(ctx, addr, keeper)
		case QueryListContractByCode:
			codeID, parseErr := strconv.ParseUint(path[1], 10, 64)
			if parseErr != nil {
				return nil, sdkerrors.Wrapf(types.ErrInvalid, "code id: %s", parseErr.Error())
			}
			//nolint:staticcheck
			rsp = queryContractListByCode(ctx, codeID, keeper)
		case QueryGetContractState:
			if len(path) < 3 {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown data query endpoint")
			}
			return queryContractState(ctx, path[1], path[2], req.Data, gasLimit, keeper)
		case QueryGetCode:
			codeID, parseErr := strconv.ParseUint(path[1], 10, 64)
			if parseErr != nil {
				return nil, sdkerrors.Wrapf(types.ErrInvalid, "code id: %s", parseErr.Error())
			}
			//nolint:staticcheck
			rsp, err = queryCode(ctx, codeID, keeper)
		case QueryListCode:
			//nolint:staticcheck
			rsp, err = queryCodeList(ctx, keeper)
		case QueryContractHistory:
			contractAddr, addrErr := sdk.AccAddressFromBech32(path[1])
			if addrErr != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, addrErr.Error())
			}
			//nolint:staticcheck
			rsp, err = queryContractHistory(ctx, contractAddr, keeper)
		case QueryInactiveContracts:
			//nolint:staticcheck
			rsp = queryInactiveContracts(ctx, keeper)
		case QueryIsInactiveContract:
			contractAddr, addrErr := sdk.AccAddressFromBech32(path[1])
			if addrErr != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, addrErr.Error())
			}
			//nolint:staticcheck
			rsp = keeper.IsInactiveContract(ctx, contractAddr)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown data query endpoint")
		}
		if err != nil {
			return nil, err
		}
		if rsp == nil || reflect.ValueOf(rsp).IsNil() {
			return nil, nil
		}
		bz, err := json.MarshalIndent(rsp, "", "  ")
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil
	}
}

func queryContractState(ctx sdk.Context, bech, queryMethod string, data []byte, gasLimit sdk.Gas, keeper types.ViewKeeper) (json.RawMessage, error) {
	contractAddr, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, bech)
	}

	switch queryMethod {
	case QueryMethodContractStateAll:
		resultData := make([]types.Model, 0)
		// this returns a serialized json object (which internally encoded binary fields properly)
		keeper.IterateContractState(ctx, contractAddr, func(key, value []byte) bool {
			resultData = append(resultData, types.Model{Key: key, Value: value})
			return false
		})
		bz, err := json.Marshal(resultData)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil
	case QueryMethodContractStateRaw:
		// this returns the raw data from the state, base64-encoded
		return keeper.QueryRaw(ctx, contractAddr, data), nil
	case QueryMethodContractStateSmart:
		// we enforce a subjective gas limit on all queries to avoid infinite loops
		ctx = ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))
		msg := types.RawContractMessage(data)
		if err := msg.ValidateBasic(); err != nil {
			return nil, sdkerrors.Wrap(err, "json msg")
		}
		// this returns raw bytes (must be base64-encoded)
		bz, err := keeper.QuerySmart(ctx, contractAddr, msg)
		return bz, err
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, queryMethod)
	}
}

//nolint:unparam
func queryCodeList(ctx sdk.Context, keeper types.ViewKeeper) ([]types.CodeInfoResponse, error) {
	var info []types.CodeInfoResponse
	keeper.IterateCodeInfos(ctx, func(i uint64, res types.CodeInfo) bool {
		info = append(info, types.CodeInfoResponse{
			CodeID:                i,
			Creator:               res.Creator,
			DataHash:              res.CodeHash,
			InstantiatePermission: res.InstantiateConfig,
		})
		return false
	})
	return info, nil
}

//nolint:unparam
func queryContractHistory(ctx sdk.Context, contractAddr sdk.AccAddress, keeper types.ViewKeeper) ([]types.ContractCodeHistoryEntry, error) {
	history := keeper.GetContractHistory(ctx, contractAddr)
	// redact response
	for i := range history {
		history[i].Updated = nil
	}
	return history, nil
}

//nolint:unparam
func queryContractListByCode(ctx sdk.Context, codeID uint64, keeper types.ViewKeeper) []string {
	var contracts []string
	keeper.IterateContractsByCode(ctx, codeID, func(addr sdk.AccAddress) bool {
		contracts = append(contracts, addr.String())
		return false
	})
	return contracts
}

func queryInactiveContracts(ctx sdk.Context, keeper types.ViewKeeper) []string {
	var contracts []string
	keeper.IterateInactiveContracts(ctx, func(contractAddress sdk.AccAddress) bool {
		contracts = append(contracts, contractAddress.String())
		return false
	})
	return contracts
}

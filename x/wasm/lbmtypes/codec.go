package lbmtypes

import (
	"github.com/line/lbm-sdk/codec"
	"github.com/line/lbm-sdk/codec/legacy"
	"github.com/line/lbm-sdk/codec/types"
	cryptocodec "github.com/line/lbm-sdk/crypto/codec"
	sdk "github.com/line/lbm-sdk/types"
	"github.com/line/lbm-sdk/types/msgservice"
	authzcodec "github.com/line/lbm-sdk/x/authz/codec"
	govcodec "github.com/line/lbm-sdk/x/gov/codec"
	govtypes "github.com/line/lbm-sdk/x/gov/types"

	wasmTypes "github.com/line/wasmd/x/wasm/types"
)

// RegisterLegacyAminoCodec registers the account types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) { //nolint:staticcheck
	legacy.RegisterAminoMsg(cdc, &MsgStoreCodeAndInstantiateContract{}, "wasm/StoreCodeAndInstantiateContract")

	cdc.RegisterConcrete(&DeactivateContractProposal{}, "wasm/DeactivateContractProposal", nil)
	cdc.RegisterConcrete(&ActivateContractProposal{}, "wasm/ActivateContractProposal", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	wasmTypes.RegisterInterfaces(registry)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgStoreCodeAndInstantiateContract{},
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&DeactivateContractProposal{},
		&ActivateContractProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz  and gov Amino codec so that this can later be
	// used to properly serialize MsgGrant, MsgExec and MsgSubmitProposal instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
	RegisterLegacyAminoCodec(govcodec.Amino)
}

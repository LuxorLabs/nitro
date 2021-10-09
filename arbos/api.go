package arbos

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

func Initialize(stateDB vm.StateDB) ArbosAPI {
	return NewArbosAPIImpl(stateDB)
}

type ArbosAPI interface {
	SplitInboxMessage(bytes []byte) ([]MessageSegment, error)

	// Should return ErrIntrinsicGas if there isn't enough gas
	StartTxHook(msg core.Message) (uint64, error) // returns amount of gas to take as extra charge

	// extraGasCharged is any gas remaining subtracted during the ExtraGasChargingHook, which is also included in totalGasUsed
	EndTxHook(
		msg core.Message,
		totalGasUsed uint64,
		extraGasCharged uint64,
	) error

	// return an extra segment (that wasn't directly in the input) that is waiting to be executed,
	GetExtraSegmentToBeNextBlock() *MessageSegment

	// StateDB can be used to read or write storage slots, balances, etc.
	FinalizeBlock(header *types.Header, txs types.Transactions, receipts types.Receipts)
}

type MessageSegment interface {
	// StateDB can be used to read *but not write* arbitrary storage slots, balances, etc.
	CreateBlockContents(
		beforeState *state.StateDB,
		api     *ArbosAPIImpl,
	) (
		[]*types.Transaction, // transactions to (try to) put in the block
		*big.Int, // timestamp
		common.Address, // coinbase address
		uint64, // gas limit
		error,
	)
}

type ArbosPrecompile interface {
	GasToCharge(input []byte) uint64

	// Important fields: evm.StateDB and evm.Config.Tracer
	// NOTE: if precompileAddress != actingAsAddress, watch out! This is a delegatecall or callcode, so caller might be wrong. In that case, unless this precompile is pure, it should probably revert.
	Call(
		input []byte,
		precompileAddress common.Address,
		actingAsAddress common.Address,
		caller common.Address,
		value *big.Int,
		readOnly bool,
		evm *vm.EVM,
	) (output []byte, err error)
}

package service

import (
	"errors"
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	gsrpc_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/ethereum-optimism/optimism/op-avail/internal/config"
	"github.com/ethereum-optimism/optimism/op-avail/internal/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/vedhavyas/go-subkey"
)

// GetBlock: To fetch the extrinsic Data from block's extrinsic by hash
func GetBlockExtrinsicData(avail_blk_ref types.AvailBlockRef, l log.Logger) ([]byte, error) {

	// Load config
	var config config.Config
	err := config.GetConfig("../op-avail/config.json")
	if err != nil {
		l.Error("Unable to create config variable for op-avail")
		panic(fmt.Sprintf("cannot get config:%v", err))
	}

	//Intitializing variables
	ApiURL := config.ApiURL
	Hash := avail_blk_ref.BlockHash
	Address := avail_blk_ref.Sender
	Nonce := avail_blk_ref.Nonce

	//Creating new substrate api
	api, err := gsrpc.NewSubstrateAPI(ApiURL)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot create api:%w", err)
	}

	// Converting this string type into gsrpc_types.hash type
	blk_hash, err := gsrpc_types.NewHashFromHexString(Hash)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to convert string hash into types.hash, error:%v", err)
	}

	// Fetching block based on block hash
	avail_blk, err := api.RPC.Chain.GetBlock(blk_hash)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot get block for hash:%v and getting error:%v", Hash, err)
	}

	//Extracting the required extrinsic according to the reference
	for _, ext := range avail_blk.Block.Extrinsics {
		//Extracting sender address for extrinsic
		ext_Addr, err := subkey.SS58Address(ext.Signature.Signer.AsID.ToBytes(), 42)
		if err != nil {
			l.Error("unable to get sender address from extrinsic", "err", err)
		}
		if ext_Addr == Address && ext.Signature.Nonce.Int64() == Nonce {
			args := ext.Method.Args
			var data []byte
			err = codec.Decode(args, &data)
			if err != nil {
				return []byte{}, fmt.Errorf("Unable to decode the extrinsic data by address: %v with nonce: %v", Address, Nonce)
			}
			return data, nil
		}
	}

	return []byte{}, errors.New(fmt.Sprintf("Didn't found any extrinsic data for address:%v in block having hash:%v", Address, Hash))
}

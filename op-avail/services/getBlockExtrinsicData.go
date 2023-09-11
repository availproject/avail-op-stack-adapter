package service

import (
	"errors"
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/ethereum-optimism/optimism/op-avail/internal/config"
	avail_types "github.com/ethereum-optimism/optimism/op-avail/types"
	"github.com/vedhavyas/go-subkey"
)

// GetBlock: To fetch the extrinsic Data from block's extrinsic by hash
func GetBlockExtrinsicData(avail_blk_ref avail_types.AvailBlockRef) ([]byte, error) {

	var config config.Config
	err := config.GetConfig("../op-avail/config.json")
	if err != nil {
		panic(fmt.Sprintf("cannot get config:%v", err))
	}

	//Intitializing variables
	ApiURL := config.ApiURL
	Hash := avail_blk_ref.BlockHash
	Address := avail_blk_ref.Sender
	Nonce := avail_blk_ref.Nonce

	api, err := gsrpc.NewSubstrateAPI(ApiURL)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot create api:%w", err)
	}

	// Converting this string type hash into types.hash type
	blk_hash, err := types.NewHashFromHexString(Hash)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to convert string hash into types.hash, error:%v", err)
	}

	avail_blk, err := api.RPC.Chain.GetBlock(blk_hash)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot get block for hash:%v and getting error:%v", Hash, err)
	}

	for _, ext := range avail_blk.Block.Extrinsics {
		// these values below are specific indexes only for data submission, differs with each extrinsic
		ext_Addr, err := subkey.SS58Address(ext.Signature.Signer.AsID.ToBytes(), 42)
		if err != nil {
			fmt.Printf("error in creating address from accountId, error:%v", err)
		}
		fmt.Println("ext_Addr=", ext_Addr, " Address=", Address, " ext_Nonce=", ext.Signature.Nonce.Int64(), " Nonce=", Nonce)
		if ext_Addr == Address && ext.Signature.Nonce.Int64() == Nonce && ext.Signature.AppID.Int64() == 1 && ext.Method.CallIndex.SectionIndex == 29 && ext.Method.CallIndex.MethodIndex == 1 {
			args := ext.Method.Args
			var data []byte
			err = codec.Decode(args, &data)
			if err != nil {
				fmt.Printf("Error in decoding args :%v", err)
			}
			return data, nil
		}
	}

	return []byte{}, errors.New(fmt.Sprintf("Didn't found any extrinsic data for address:%v in block having hash:%v", Address, Hash))
}

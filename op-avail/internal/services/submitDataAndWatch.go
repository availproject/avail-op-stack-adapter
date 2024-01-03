package service

import (
	"errors"
	"fmt"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	gsrpc_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum-optimism/optimism/op-avail/internal/config"
	"github.com/ethereum-optimism/optimism/op-avail/internal/types"
	"github.com/ethereum-optimism/optimism/op-avail/internal/utils"
	"github.com/ethereum/go-ethereum/log"
)

// submitData creates a transaction and makes a Avail data submission
func SubmitDataAndWatch(api *gsrpc.SubstrateAPI, config config.DAConfig, data []byte) (types.AvailBlockRef, error) {

	Seed := config.Seed
	AppID := config.AppID

	// //Creating new substrate api
	// api, err := gsrpc.NewSubstrateAPI(ApiURL)
	// if err != nil {
	// 	fmt.Printf("cannot create api: error:%v", err)
	// 	return types.AvailBlockRef{}, err
	// }

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		fmt.Printf("cannot get metadata: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	appID := 0
	// if app id is greater than 0 then it must be created before submitting data
	if AppID != 0 {
		appID = AppID
	}

	c, err := gsrpc_types.NewCall(meta, "DataAvailability.submit_data", gsrpc_types.NewBytes([]byte(data)))
	if err != nil {
		fmt.Printf("cannot create new call: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	// Create the extrinsic
	ext := gsrpc_types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		fmt.Printf("cannot get block hash: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		fmt.Printf("cannot get runtime version: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	keyringPair, err := signature.KeyringPairFromSecret(Seed, 42)
	if err != nil {
		fmt.Printf("cannot create LeyPair: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	key, err := gsrpc_types.CreateStorageKey(meta, "System", "Account", keyringPair.PublicKey)
	if err != nil {
		fmt.Printf("cannot create storage key: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	var accountInfo gsrpc_types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		fmt.Printf("cannot get latest storage: error:%v", err)
		if !ok && err == nil {
			err = fmt.Errorf("cannot get latest stotage")
		}
		return types.AvailBlockRef{}, err
	}

	nonce := utils.GetAccountNonce(uint32(accountInfo.Nonce))
	//fmt.Println("Nonce from localDatabase:", nonce, "    ::::::::   from acountInfo:", accountInfo.Nonce)
	o := gsrpc_types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                gsrpc_types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              gsrpc_types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                gsrpc_types.NewUCompactFromUInt(0),
		AppID:              gsrpc_types.NewUCompactFromUInt(uint64(appID)),
		TransactionVersion: rv.TransactionVersion,
	}

	// Sign the transaction using Alice's default account
	err = ext.Sign(keyringPair, o)
	if err != nil {
		fmt.Printf("cannot sign: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	// Send the extrinsic
	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		fmt.Printf("cannot submit extrinsic: error:%v", err)
		return types.AvailBlockRef{}, err
	}

	log.Info("Tx batch is submitted to Avail", "length", len(data), "address", keyringPair.Address, "appID", appID)

	defer sub.Unsubscribe()
	timeout := time.After(500 * time.Second)
	for {
		select {
		case status := <-sub.Chan():
			if status.IsFinalized {
				return types.AvailBlockRef{BlockHash: string(status.AsFinalized.Hex()), Sender: keyringPair.Address, Nonce: o.Nonce.Int64()}, nil
			}
		case <-timeout:
			return types.AvailBlockRef{}, errors.New("Timitout before getting finalized status")
		}
	}
}

package service

import (
	"errors"
	"fmt"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum-optimism/optimism/op-avail/internal/config"
	avail_types "github.com/ethereum-optimism/optimism/op-avail/internal/types"
	"github.com/ethereum-optimism/optimism/op-avail/internal/utils"
)

// submitData creates a transaction and makes a Avail data submission
func SubmitDataAndWatch(data []byte) (avail_types.AvailBlockRef, error) {

	var config config.Config
	err := config.GetConfig("../op-avail/config.json")
	if err != nil {
		panic(fmt.Sprintf("cannot get config:%v", err))
	}

	//Intitializing variables
	ApiURL := config.ApiURL
	Seed := config.Seed
	AppID := config.AppID

	api, err := gsrpc.NewSubstrateAPI(ApiURL)
	if err != nil {
		fmt.Printf("cannot create api: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		fmt.Printf("cannot get metadata: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	appID := 0
	// if app id is greater than 0 then it must be created before submitting data
	if AppID != 0 {
		appID = AppID
	}

	c, err := types.NewCall(meta, "DataAvailability.submit_data", types.NewBytes([]byte(data)))
	if err != nil {
		fmt.Printf("cannot create new call: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		fmt.Printf("cannot get block hash: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		fmt.Printf("cannot get runtime version: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	keyringPair, err := signature.KeyringPairFromSecret(Seed, 42)
	if err != nil {
		fmt.Printf("cannot create LeyPair: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", keyringPair.PublicKey)
	if err != nil {
		fmt.Printf("cannot create storage key: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		fmt.Printf("cannot get latest storage: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	nonce := utils.GetAccountNonce(uint32(accountInfo.Nonce))
	//fmt.Println("Nonce from localDatabase:", nonce, "    ::::::::   from acountInfo:", accountInfo.Nonce)
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		AppID:              types.NewUCompactFromUInt(uint64(appID)),
		TransactionVersion: rv.TransactionVersion,
	}

	// Sign the transaction using Alice's default account
	err = ext.Sign(keyringPair, o)
	if err != nil {
		fmt.Printf("cannot sign: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	// Send the extrinsic
	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		fmt.Printf("cannot submit extrinsic: error:%v", err)
		return avail_types.AvailBlockRef{}, err
	}

	fmt.Println("Data of lenght :", len(data), "submitted by op-stack with address ", keyringPair.Address, " using appID ", appID)

	defer sub.Unsubscribe()
	timeout := time.After(50 * time.Second)
	for {
		select {
		case status := <-sub.Chan():
			if status.IsInBlock {
				fmt.Printf("Txn inside block %v\n", status.AsInBlock.Hex())
			}
			return avail_types.AvailBlockRef{BlockHash: status.AsInBlock.Hex(), Sender: keyringPair.Address, Nonce: o.Nonce}, nil
		case <-timeout:
			fmt.Printf("timeout of 100 seconds reached without getting finalized status for extrinsic")
			return avail_types.AvailBlockRef{}, errors.New("Timitout before getting finalized status")
		}
	}
}

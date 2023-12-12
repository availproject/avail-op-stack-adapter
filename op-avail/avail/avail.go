package avail

import (
	"fmt"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"

	"github.com/ethereum/go-ethereum/log"

	config "github.com/ethereum-optimism/optimism/op-avail/internal/config"
	service "github.com/ethereum-optimism/optimism/op-avail/internal/services"
	types "github.com/ethereum-optimism/optimism/op-avail/internal/types"
)

// AvailMessageHeaderFlag indicates that this data is a Blob Pointer
// which will be used to retrieve data from Avail
const AvailMessageHeaderFlag byte = 0x0a

func IsAvailMessageHeaderByte(header byte) bool {
	return (AvailMessageHeaderFlag & header) > 0
}

type AvailDA struct {
	cfg config.DAConfig
	api *gsrpc.SubstrateAPI
}

func NewAvailDA(path string) (*AvailDA, error) {

	// Load config
	var cfg config.DAConfig
	err := cfg.GetConfig(path)
	if err != nil {
		log.Error("Unable to create config variable for op-avail")
		panic(fmt.Sprintf("cannot get config:%v", err))
	}

	//Creating new substrate api
	api, err := gsrpc.NewSubstrateAPI(cfg.ApiURL)
	if err != nil {
		return nil, err
	}

	log.Info("Connected to Avail ✅")
	return &AvailDA{
		cfg: cfg,
		api: api,
	}, nil
}

func (a *AvailDA) SubmitTxDataAndGetRef(TxData []byte) ([]byte, error) {
	log.Info("Working on batch submission for avail")

	//Checking for the size of TxData
	if len(TxData) >= 512000 {
		return []byte{}, fmt.Errorf("size of TxData is more than 512KB, it is higher than a single data submit transaction supports on avail")
	}

	// Submitting data to Avail
	avail_Blk_Ref, err := service.SubmitDataAndWatch(a.api, a.cfg, TxData)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot submit data:%v", err)
	}

	log.Info("Avail Block Reference:", "Ref", avail_Blk_Ref)

	ref_bytes_Data, err := avail_Blk_Ref.MarshalToBinary()
	if err != nil {
		return []byte{}, fmt.Errorf("cannot get the binary form of avail block reference:%v", err)
	}

	return ref_bytes_Data, nil
}

func (a *AvailDA) GetTxDataByDARef(RefData []byte) ([]byte, error) {
	//Getting Avail block reference from callData
	avail_blk_ref := types.AvailBlockRef{}
	err := avail_blk_ref.UnmarshalFromBinary(RefData)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to unmarshal the ethereum trxn data to avail block refrence, error:%v", err)
	}
	log.Info("Avail Block Reference:", "Ref", avail_blk_ref)

	txData, err := service.GetBlockExtrinsicData(a.api, a.cfg, avail_blk_ref)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get block extrinsic data, error:%v", err)
	}

	return txData, nil
}

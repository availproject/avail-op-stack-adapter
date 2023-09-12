package avail

import (
	"fmt"

	service "github.com/ethereum-optimism/optimism/op-avail/internal/services"
	types "github.com/ethereum-optimism/optimism/op-avail/internal/types"
)

func GetTxDataByDARef(RefData []byte) ([]byte, error) {
	//Getting Avail block reference from callData
	avail_blk_ref := types.AvailBlockRef{}
	err := avail_blk_ref.UnmarshalFromBinary(RefData)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to unmarshal the ethereum trxn data to avail block refrence, error:%v", err)
	}
	fmt.Printf("Avail Block Reference: %+v", avail_blk_ref)

	txData, err := service.GetBlockExtrinsicData(avail_blk_ref)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to get block extrinsic data, error:%v", err)
	}

	return txData, nil
}

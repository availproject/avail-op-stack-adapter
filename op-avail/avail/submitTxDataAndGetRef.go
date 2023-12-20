package avail

import (
	"fmt"

	service "github.com/ethereum-optimism/optimism/op-avail/internal/services"
	"github.com/ethereum/go-ethereum/log"
)

func SubmitTxDataAndGetRef(TxData []byte, l log.Logger) ([]byte, error) {
	l.Info("Working on batch submission for avail")

	//Checking for the size of TxData
	if len(TxData) >= 512000 {
		return []byte{}, fmt.Errorf("size of TxData is more than 512KB, it is higher than a single data submit transaction supports on avail")
	}

	// Submitting data to Avail
	avail_Blk_Ref, err := service.SubmitDataAndWatch(TxData, l)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot submit data:%v", err)
	}

	l.Info("Avail Block Reference:", "Ref", avail_Blk_Ref)

	ref_bytes_Data, err := avail_Blk_Ref.MarshalToBinary()
	if err != nil {
		return []byte{}, fmt.Errorf("cannot get the binary form of avail block reference:%v", err)
	}

	return ref_bytes_Data, nil
}

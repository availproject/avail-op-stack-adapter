package avail

import (
	"fmt"

	service "github.com/ethereum-optimism/optimism/op-avail/internal/services"
)

func SubmitTxDataAndGetRef(TxData []byte) ([]byte, error) {
	fmt.Println("Working on batch submission for avail")

	// Submitting data to Avail
	avail_Blk_Ref, err := service.SubmitDataAndWatch(TxData)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot submit data:%v", err)
	}

	fmt.Printf("Avail Block Reference: %+v", avail_Blk_Ref)

	ref_bytes_Data, err := avail_Blk_Ref.MarshalToBinary()
	if err != nil {
		return []byte{}, fmt.Errorf("cannot get the binary form of avail block reference:%v", err)
	}

	return ref_bytes_Data, nil
}

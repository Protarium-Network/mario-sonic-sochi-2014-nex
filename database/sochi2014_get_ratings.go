package database

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	datastore "github.com/PretendoNetwork/nex-protocols-go/v2/datastore"
	datastore_types "github.com/PretendoNetwork/nex-protocols-go/v2/datastore/types"
)

// Sochi2014GetRatings reproduces the response layout used by the original
// Nintendo server. Sochi expects one rating-slot list and one result per data
// ID; a flat List<DataStoreRatingInfo> makes the client lose byte alignment.
func Sochi2014GetRatings(err error, packet nex.PacketInterface, callID uint32, dataIDs types.List[types.UInt64], accessPassword types.UInt64) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.DataStore.Unknown, "change_error")
	}

	endpoint := packet.Sender().Endpoint()
	ratings := types.NewList[types.List[datastore_types.DataStoreRatingInfoWithSlot]]()
	results := types.NewList[types.QResult]()

	for _, dataID := range dataIDs {
		slotCount := 8
		// The world-ranking objects in the Nintendo capture use 16 rating
		// slots, while the earlier 910xxx objects use 8.
		if uint64(dataID) >= 941238 && uint64(dataID) <= 941537 {
			slotCount = 16
		}

		slots := types.NewList[datastore_types.DataStoreRatingInfoWithSlot]()
		for slot := 0; slot < slotCount; slot++ {
			rating := datastore_types.NewDataStoreRatingInfoWithSlot()
			rating.Slot = types.NewInt8(int8(slot))
			slots = append(slots, rating)
		}

		ratings = append(ratings, slots)
		// 0x00690001 is the successful DataStore result returned for every
		// entry by the original server capture.
		results = append(results, types.NewQResultSuccess(nex.ResultCodes.DataStore.Unknown))
	}

	rmcResponseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	ratings.WriteTo(rmcResponseStream)
	results.WriteTo(rmcResponseStream)

	rmcResponse := nex.NewRMCSuccess(endpoint, rmcResponseStream.Bytes())
	rmcResponse.ProtocolID = datastore.ProtocolID
	rmcResponse.MethodID = datastore.MethodGetRatings
	rmcResponse.CallID = callID

	return rmcResponse, nil
}

package nex

import (
	"slices"

	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	commonglobals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	matchmakedatabase "github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension/database"
	matchmakingtypes "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	matchmakeextension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	"github.com/Protarium-Network/mario-sonic-sochi-2014-nex/globals"
)

// sochiAutoMatchmakeWithGatheringIDPostpone implements the friends-only
// matchmaking path used by Sochi. The common protocol currently has no
// implementation for method 0x21.
func sochiAutoMatchmakeWithGatheringIDPostpone(
	err error,
	packet nex.PacketInterface,
	callID uint32,
	gatheringIDs types.List[types.UInt32],
	gatheringHolder matchmakingtypes.GatheringHolder,
	message types.String,
) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}
	if len(message) > 256 || len(gatheringIDs) > 100 {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	connection, ok := packet.Sender().(*nex.PRUDPConnection)
	if !ok {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)
	manager := globals.MatchmakingManager

	sourceSession, ok := gatheringHolder.Object.(matchmakingtypes.MatchmakeSession)
	if !ok || !commonglobals.CheckValidMatchmakeSession(sourceSession) {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	matchmakedatabase.EndMatchmakeSessionsParticipation(manager, connection)

	friendPIDs := []uint32{}
	if manager.GetUserFriendPIDs != nil {
		friendPIDs = manager.GetUserFriendPIDs(uint32(connection.PID()))
	}

	var resultSession matchmakingtypes.MatchmakeSession
	found := false
	for _, gatheringID := range gatheringIDs {
		candidate, _, candidateError := matchmakedatabase.GetMatchmakeSessionByID(manager, endpoint, uint32(gatheringID))
		if candidateError != nil {
			if candidateError.ResultCode == nex.ResultCodes.RendezVous.SessionVoid {
				continue
			}
			return nil, candidateError
		}

		// Method 0x21 is the friends path. Do not let a client-provided GID
		// bypass the friendship and normal joinability checks.
		if !slices.Contains(friendPIDs, uint32(candidate.OwnerPID)) ||
			!bool(candidate.OpenParticipation) ||
			candidate.HostPID == 0 ||
			bool(candidate.UserPasswordEnabled) ||
			bool(candidate.SystemPasswordEnabled) ||
			(candidate.MaximumParticipants != 0 && candidate.ParticipationCount >= types.UInt32(candidate.MaximumParticipants)) {
			continue
		}

		resultSession = candidate
		found = true
		break
	}

	if !found {
		resultSession = sourceSession.Copy().(matchmakingtypes.MatchmakeSession)
		if createError := matchmakedatabase.CreateMatchmakeSession(manager, connection, &resultSession); createError != nil {
			return nil, createError
		}
	}

	participants, joinError := matchmakedatabase.JoinMatchmakeSession(manager, resultSession, connection, 1, string(message))
	if joinError != nil {
		return nil, joinError
	}
	resultSession.ParticipationCount = types.NewUInt32(participants)

	resultHolder := matchmakingtypes.NewGatheringHolder()
	resultHolder.Object = resultSession.Copy().(matchmakingtypes.GatheringInterface)
	responseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())
	resultHolder.WriteTo(responseStream)

	response := nex.NewRMCSuccess(endpoint, responseStream.Bytes())
	response.ProtocolID = matchmakeextension.ProtocolID
	response.MethodID = matchmakeextension.MethodAutoMatchmakeWithGatheringIDPostpone
	response.CallID = callID

	return response, nil
}

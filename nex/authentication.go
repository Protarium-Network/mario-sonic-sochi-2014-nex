// Package nex is the Mario & Sonic at the Sochi 2014 Olympic Winter Games
// NEX server (game_server_id 10106900). Same series/engine as Rio 2016, but
// captures confirm this title has real online matchmaking on top of ranking
// and DataStore, so the secure endpoint registers the match-making
// protocols too. Authentication and secure run as separate PRUDP endpoints.
package nex

import (
	"fmt"
	"os"
	"strconv"

	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_ticket_granting "github.com/PretendoNetwork/nex-protocols-common-go/v2/ticket-granting"
	ticket_granting "github.com/PretendoNetwork/nex-protocols-go/v2/ticket-granting"
	"github.com/Protarium-Network/mario-sonic-sochi-2014-nex/globals"
)

var AuthenticationServer *nex.PRUDPServer
var AuthenticationEndpoint *nex.PRUDPEndPoint

func StartAuthenticationServer() {
	AuthenticationServer = nex.NewPRUDPServer()

	AuthenticationEndpoint = nex.NewPRUDPEndPoint(1)
	AuthenticationEndpoint.ServerAccount = globals.AuthenticationServerAccount
	AuthenticationEndpoint.AccountDetailsByPID = globals.AccountDetailsByPID
	AuthenticationEndpoint.AccountDetailsByUsername = globals.AccountDetailsByUsername
	AuthenticationServer.BindPRUDPEndPoint(AuthenticationEndpoint)

	AuthenticationServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(globals.NEXMajor, globals.NEXMinor, globals.NEXPatch))
	AuthenticationServer.AccessKey = globals.AccessKey
	// Unlike Rio 2016, confirmed via a real LoginEx capture that Sochi 2014's
	// client does NOT write a version byte before the AuthenticationInfo.Data
	// structure - just a plain 4-byte content length, then [token string]
	// [ngsVersion u32][tokenType u8][serverVersion u32] directly. With
	// UseStructureHeader=true nex-go's parser reads a version byte first,
	// misaligning everything after it ("Structure content length longer than
	// data size"). Leave it at the library default (false).

	AuthenticationEndpoint.OnData(func(packet nex.PacketInterface) {
		request := packet.RMCMessage()
		if request == nil {
			return
		}
		pid := uint64(packet.Sender().PID())
		fmt.Printf("[Sochi2014 Auth] PID=%d protocol=0x%02X method=0x%02X\n", pid, request.ProtocolID, request.MethodID)
		// One-off wire-format diagnostic: dump the raw RMC
		// parameter bytes for LoginEx so the AuthenticationInfo layout can be
		// confirmed by hand against a real capture instead of guessing.
		if request.ProtocolID == 0x0A && request.MethodID == 0x02 {
			fmt.Printf("[Sochi2014 Auth] RAW LoginEx params (%d bytes): %x\n", len(request.Parameters), request.Parameters)
		}
	})

	registerAuthenticationServerProtocols()

	port, _ := strconv.Atoi(os.Getenv("PN_SOCHI2014_AUTH_PORT"))
	globals.Logger.Successf("[Sochi2014] Authentication server listening on UDP %d", port)
	AuthenticationServer.Listen(port)
}

func registerAuthenticationServerProtocols() {
	ticketGrantingProtocol := ticket_granting.NewProtocol()
	AuthenticationEndpoint.RegisterServiceProtocol(ticketGrantingProtocol)
	commonTicketGrantingProtocol := common_ticket_granting.NewCommonProtocol(ticketGrantingProtocol)

	securePort, _ := strconv.Atoi(os.Getenv("PN_SOCHI2014_SECURE_PORT"))

	// Must stay short (~15 chars): the retail binary truncates this field
	// into a small fixed-size buffer, so use a bare host, not a subdomain.
	secureHost := os.Getenv("PN_SOCHI2014_SECURE_HOST")
	if secureHost == "" {
		secureHost = "localhost"
	}

	secureStationURL := types.NewStationURL("")
	secureStationURL.SetURLType(constants.StationURLPRUDPS)
	secureStationURL.SetAddress(secureHost)
	secureStationURL.SetPortNumber(uint16(securePort))
	secureStationURL.SetConnectionID(1)
	secureStationURL.SetPrincipalID(types.NewPID(2))
	secureStationURL.SetStreamID(1)
	secureStationURL.SetStreamType(constants.StreamTypeRVSecure)
	secureStationURL.SetType(uint8(constants.StationURLFlagPublic))

	commonTicketGrantingProtocol.ValidateLoginData = globals.ValidateLoginData
	commonTicketGrantingProtocol.SecureStationURL = secureStationURL
	commonTicketGrantingProtocol.BuildName = types.NewString("")
	commonTicketGrantingProtocol.SecureServerAccount = globals.SecureServerAccount
}

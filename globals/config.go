package globals

// NEX configuration for "Mario & Sonic at the Sochi 2014 Olympic Winter
// Games" (game_server_id 10106900 / 269510912 decimal).
//
// AccessKey and game_server_id come from kinnay.github.io's public Wii U NEX
// game database (entry "MARIO & SONIC SOCHI 2014"); the game_server_id also
// matches exactly what the console requested from the account/nex_token
// endpoint.
const (
	GameServerID = "10106900"
	AccessKey    = "585214a5"

	// PRUDP library version reported by both endpoints. Sochi 2014 ships NEX
	// 3.4.7. Two libraries are overridden at runtime (see nex/secure.go):
	// MatchMaking is pinned to 3.3.0 (Sochi uses the pre-3.4 MatchmakeSession
	// / search-criteria layout) and Ranking is bumped to 3.6.0 so
	// RankingRankData.UpdateTime is serialized.
	NEXMajor = 3
	NEXMinor = 4
	NEXPatch = 7

	MatchMakingMajor = 3
	MatchMakingMinor = 3
	MatchMakingPatch = 0
)

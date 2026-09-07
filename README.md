# Mario & Sonic at the Sochi 2014 Olympic Winter Games — NEX server

A preservation-oriented NEX server for the Wii U title **Mario & Sonic at the
Sochi 2014 Olympic Winter Games** (`game_server_id` `10106900`). It speaks the
game's PRUDP authentication and secure protocols: leaderboards, DataStore
score attachments, and **online matchmaking** (which real 2014 captures of
this title confirm).

Built on the [Pretendo Network](https://github.com/PretendoNetwork) NEX
libraries (AGPL-3.0). `internal/nex-protocols-common-go-patch/` is a vendored
fork carrying the DataStore/S3 changes this flow needs.

## Recovered configuration

| Field            | Value |
|------------------|-------|
| Game server ID   | `10106900` |
| Access key       | `585214a5` (kinnay.github.io Wii U NEX database, cross-checked against the console's own `nex_token` request) |
| NEX SDK version  | `3.4.7` |
| MatchMaking lib  | pinned to `3.3.0` (Sochi uses the pre-3.4 MatchmakeSession / search-criteria layout) |
| Ranking lib      | bumped to `3.6.0` (so `RankingRankData.UpdateTime` serializes) |
| Byte stream      | no structure headers (`UseStructureHeader = false`, unlike Rio 2016) |

## Scope

- **Ticket Granting** — login / secure-server handoff
- **Secure Connection**, **Utility** — baseline secure-endpoint handshake
- **Ranking** — leaderboards, common data, score upload
- **DataStore** — score attachments (ghosts / photos): `GetRatings`,
  `PostMetaBinary`, `PrepareGetObject`, `PreparePostObject` → S3 → metadata
  updates, `CompletePostObject`
- **NAT Traversal + MatchMaking + MatchMakingExt + MatchmakeExtension** —
  "find an opponent" online play. Includes a Sochi-specific decoder patch for
  `AutoMatchmakeWithSearchCriteria_Postpone` (pre-3.4 wire layout) and a
  friends-only `AutoMatchmakeWithGatheringID_Postpone` handler.

## Database

One PostgreSQL database holds everything: the `sochi2014_*` ranking tables,
the `datastore` schema, and the `matchmaking` / `tracking` schemas. The
schema is created on first start.

The `matchmaking` / `tracking` schema is **hand-authored** — the NEX common
library ships the queries but not the DDL. See
[docs/matchmaking-schema.md](docs/matchmaking-schema.md).

## Running

### Local preservation mode

```bash
cp .env.example .env                     # PN_SOCHI2014_LOCAL_MODE=1 by default
cp settings.example.json settings.json   # add your console's PID + NEX password
docker compose up --build
```

In local mode there is no account server: player NEX passwords come from
`settings.json` and the login token is accepted unconditionally. Use it only
on an isolated network.

### Shared mode

Set `PN_SOCHI2014_LOCAL_MODE` to anything but `1` and provide
`PN_SOCHI2014_NEX_TOKEN_AES_KEY` (64 hex chars) and
`PN_SOCHI2014_NEX_PASSWORD_SECRET` (≥32 bytes hex), matching your account
server.

### Without Docker

```bash
go build -o sochi2014-nex .
./sochi2014-nex
```

## License

AGPL-3.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

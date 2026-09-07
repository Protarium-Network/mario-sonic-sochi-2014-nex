# Protocol coverage

What this server implements for Mario & Sonic at the Sochi 2014 Olympic
Winter Games, and how each was confirmed.

## Authentication endpoint

| Protocol         | Notes |
|------------------|-------|
| Ticket Granting  | Login / LoginEx / RequestTicket. `ValidateLoginData` accepts the account-server token (or anything, in local mode). |

Unlike Rio 2016, this title's client does **not** write structure-header
version bytes before `AuthenticationInfo.Data` — just a 4-byte content
length. `ByteStreamSettings.UseStructureHeader` is left `false` on both
endpoints; setting it `true` misaligns the token fields.

## Secure endpoint

| Protocol            | Coverage | Confirmed by |
|---------------------|----------|--------------|
| Secure Connection   | baseline handshake, insecure `Register` | required for any secure endpoint |
| Utility             | baseline | baseline calls |
| Ranking (lib 3.6.0) | `GetRankings`, `GetRankingsAndCount`, `UploadScore`, common data get/upload | leaderboard list + submission |
| DataStore           | `GetRatings` (Sochi's slot-list layout), `PostMetaBinary`, `PrepareGetObject`, `PreparePostObject` → S3 → `CompletePostObject`, `ChangeMeta`, `GetMetasMultipleParam`, period / meta-binary / data-type updates | real score-upload capture (`GetObjectOwnerByDataID not defined` before wiring the CompletePostObject callbacks) |
| NAT Traversal       | `ReportNATProperties` etc. | real capture |
| MatchMaking / MatchMakingExt | via the common manager | real capture |
| MatchmakeExtension (lib 3.3.0) | `AutoMatchmakeWithSearchCriteria_Postpone` (Sochi pre-3.4 decoder patch), `AutoMatchmakeWithGatheringID_Postpone` (friends-only handler), `OpenParticipation`, `EndParticipation` | real 2014 capture against Nintendo's servers |

### Sochi matchmaking specifics

- `MatchMaking` library pinned to `3.3.0`: `MatchmakeSession` has no
  `ProgressScore` byte and search criteria have no `VacantParticipants`
  field in this title's wire layout.
- `nex/automatch_patch.go` decodes method `0x0F` with the pre-3.4
  search-criteria / gathering layout and injects `VacantParticipants = 1`
  (the historical default) before handing off to the common handler.
- `nex/friends_automatch.go` implements method `0x21`
  (`AutoMatchmakeWithGatheringID_Postpone`), which the common protocol does
  not provide, with friendship + joinability checks.

## Not implemented

- Persistent gatherings / communities beyond the schema and the common
  library's own handlers — this title does not appear to use them.

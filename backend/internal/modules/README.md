# Backend Modules

Feature-oriented backend modules live here.

Current extracted modules:

- `auth`
- `user`
- `team`
- `rbac`

For this refactor pass, `domain`, `repository`, and `service` packages for these areas were moved under `internal/modules/*`.

HTTP handlers still live under `internal/handler` for now and can be moved later once the routing layer is split by module as well.

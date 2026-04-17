---
name: eco-knock-maintainer
description: Maintain this repository while preserving its current architecture, naming, and layering. Use when adding features, refactoring packages, reorganizing directories, wiring gRPC or config code, or checking whether a change fits this repository's established style.
---

# Eco Knock Maintainer

Keep changes aligned with this repository's existing style before introducing new patterns.

## Workflow

1. Inspect nearby code before changing structure.
2. Copy the repository's current pattern instead of introducing a cleaner-looking pattern from scratch.
3. Keep layers explicit and shallow.
4. Update tests and wiring in the same change.
5. Run `go test ./...` before finishing.

## Repository Rules

- Prefer explicit, domain-named files over generic helper buckets.
- Prefer shallow directories over heavily fragmented package trees.
- Keep transport adapters under `internal/grpc/server/...`.
- Keep domain and hardware logic under `internal/<domain>/...`.
- Keep top-level wiring under `cmd/server`.
- Keep protocol-specific code inside the domain package that owns it.
- Move code to `internal/common/...` only after it is clearly reused across domains.

## Naming Rules

- Prefer file names that expose the role directly, such as `sensor_grpc_server.go` or `xiaomi_airpurifier_service.go`.
- Avoid creating `helpers.go`, `types.go`, `commands.go`, or `utils.go` unless the repository already uses that pattern nearby.
- Keep package names short and concrete.
- Keep DTO names explicit when they cross boundaries.

## Language Rules

- Keep user-facing and developer-facing prose in Korean unless there is a clear technical reason not to.
- Write log messages, error messages, validation messages, comments, and README/document text in Korean.
- Keep code identifiers such as package names, file names, type names, function names, and protobuf field names aligned with the existing code style, even when the surrounding prose is Korean.
- When touching nearby code that mixes English and Korean prose, normalize the touched prose to Korean in the same change when practical.

## gRPC and Error Handling

- Keep protobuf contracts in `proto/...` and generated code in `internal/grpc/pb/...`.
- Register multiple gRPC services on one server when they share the same port.
- Split gRPC startup into `startGRPCServer(...)` plus service-specific helper functions when wiring grows.
- Return `apperror.AppError`-based gRPC errors through `apperror.ToGRPCError(...)`.
- Keep gRPC handlers focused on request/response mapping and service invocation.

## Configuration Rules

- Keep app-wide config in `internal/common/config/common_config.go`.
- Keep domain-specific parsing logic in the domain config package when possible.
- Use `application.yaml` with `${ENV_NAME}` placeholders.
- Avoid hidden default values in code unless the default is intentional and stable.

## Refactoring Rules

- Remove obsolete code paths when architecture changes instead of leaving parallel flows behind.
- Keep test helpers under `internal/<domain>/.../test/...` when reused across multiple tests.
- Refactor names and variable names while moving structure so the final code reads consistently.
- Favor repository precedent over generic clean-architecture advice when the two conflict.

## Checks

- Read the nearest related package before inventing a new layout.
- Keep new files and directories justified by an existing pattern or repeated need.
- Run `go test ./...`.

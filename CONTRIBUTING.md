# Contributing

Thanks for improving ai-credit-scrub.

## Development rules

- Keep matching conservative: a product name alone is never sufficient.
- Add a false-positive fixture whenever adding a new built-in signature.
- Add a provider fixture and a user-facing README entry for any new adapter.
- Run `go test ./...`, `go vet ./...`, and `gofmt -w` on changed Go files.

## Pull requests

Describe the new signature or behavior, include fixtures that demonstrate both
removal and preservation, and avoid collecting or uploading user content.

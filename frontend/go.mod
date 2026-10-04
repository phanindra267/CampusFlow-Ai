// This nested module marks frontend/ as outside the API module.
//
// npm packages occasionally ship Go sources of their own, and
// frontend/node_modules/flatted/golang is one such case. Without this boundary
// the parent module's ./... pattern picks those files up, so `go build ./...`,
// `go vet ./...` and `go test ./...` would compile third-party code from
// node_modules. A nested go.mod excludes the subtree from the parent module.
//
// There is intentionally no Go code in this directory.
module github.com/campuscare/api/frontend

go 1.26.5
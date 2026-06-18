// Package client provides go.uber.org/mock mocks for the controller-runtime
// client interfaces, replacing the mocks gardener dropped from
// third_party/mock/controller-runtime with v1.14x.
package client

//go:generate go run go.uber.org/mock/mockgen -package client -destination mocks.go sigs.k8s.io/controller-runtime/pkg/client Client,StatusWriter,SubResourceWriter

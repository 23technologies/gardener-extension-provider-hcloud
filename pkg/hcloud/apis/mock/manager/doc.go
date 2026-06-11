// Package manager provides go.uber.org/mock mocks for the controller-runtime
// manager interface, replacing the mocks gardener dropped from
// third_party/mock/controller-runtime with v1.14x.
package manager

//go:generate go run go.uber.org/mock/mockgen -package manager -destination mocks.go sigs.k8s.io/controller-runtime/pkg/manager Manager

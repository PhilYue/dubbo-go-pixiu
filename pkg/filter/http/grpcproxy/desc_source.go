package grpcproxy
//
//import (
//	"context"
//	"github.com/jhump/protoreflect/desc"
//	"github.com/jhump/protoreflect/grpcreflect"
//)
//
//type DescriptorSource interface {
//	// ListServices returns a list of fully-qualified service names. It will be all services in a set of
//	// descriptor files or the set of all services exposed by a gRPC server.
//	ListServices() ([]string, error)
//	// FindSymbol returns a descriptor for the given fully-qualified symbol name.
//	FindSymbol(fullyQualifiedName string) (desc.Descriptor, error)
//	// AllExtensionsForType returns all known extension fields that extend the given message type name.
//	AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error)
//}
//
//
//type serverSource struct {
//	client *grpcreflect.Client
//}
//
//func (s serverSource) ListServices() ([]string, error) {
//	panic("implement me")
//}
//
//func (s serverSource) FindSymbol(fullyQualifiedName string) (desc.Descriptor, error) {
//	file, err := s.client.FileContainingSymbol(fullyQualifiedName)
//	if err != nil {
//		return nil, err
//	}
//	d := file.FindSymbol(fullyQualifiedName)
//	if d == nil {
//		return nil, nil
//	}
//	return d, nil
//}
//
//func (s serverSource) AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error) {
//	panic("implement me")
//}
//
//func DescriptorSourceFromServer(_ context.Context, refClient *grpcreflect.Client) DescriptorSource {
//	return serverSource{client: refClient}
//}
//
//

package grpcproxy

import (
	"fmt"
	"github.com/apache/dubbo-go-pixiu/pkg/logger"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DescriptorSource interface {
	// ListServices returns a list of fully-qualified service names. It will be all services in a set of
	// descriptor files or the set of all services exposed by a gRPC server.
	ListServices() ([]string, error)
	// FindSymbol returns a descriptor for the given fully-qualified symbol name.
	FindSymbol(fullyQualifiedName string) (desc.Descriptor, error)
	// AllExtensionsForType returns all known extension fields that extend the given message type name.
	AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error)
}

var ErrReflectionNotSupported = errors.New("server does not support the reflection API")

var reflSource DescriptorSource

type serverSource struct {
	client *grpcreflect.Client
}

func GetReflSource(client *grpcreflect.Client) DescriptorSource {
	if reflSource == nil {
		reflSource = serverSource{client: client}
	}
	return reflSource
}

//func (s serverSource) setClient(client *grpcreflect.Client) *serverSource {
//	s.client = client
//	return &s
//}

func (s serverSource) ListServices() ([]string, error) {
	svcs, err := s.client.ListServices()
	return svcs, reflectionSupport(err)
}

func (s serverSource) FindSymbol(fullyQualifiedName string) (desc.Descriptor, error) {
	file, err := s.client.FileContainingSymbol(fullyQualifiedName)
	if err != nil {
		return nil, reflectionSupport(err)
	}
	d := file.FindSymbol(fullyQualifiedName)
	if d == nil {
		return nil, errors.New(fmt.Sprintf("%s not found: %s", "Symbol", fullyQualifiedName))
	}
	return d, nil
}

func (s serverSource) AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error) {
	var exts []*desc.FieldDescriptor
	nums, err := s.client.AllExtensionNumbersForType(typeName)
	if err != nil {
		return nil, reflectionSupport(err)
	}
	for _, fieldNum := range nums {
		ext, err := s.client.ResolveExtension(typeName, fieldNum)
		if err != nil {
			return nil, reflectionSupport(err)
		}
		exts = append(exts, ext)
	}
	return exts, nil
}

func reflectionSupport(err error) error {
	if err == nil {
		return nil
	}
	if stat, ok := status.FromError(err); ok && stat.Code() == codes.Unimplemented {
		return ErrReflectionNotSupported
	}
	return err
}

var descSource *compositeSource

// composite
type compositeSource struct {
	reflection DescriptorSource
	file       DescriptorSource
}

func GetCmpSource() *compositeSource {
	if descSource == nil {
		descSource = &compositeSource{}
	}
	return descSource
}

func (cs *compositeSource) WithFileDS(file DescriptorSource) *compositeSource {
	cs.file = file
	return cs
}

func (cs *compositeSource) WithReflectionDS(reflection DescriptorSource) *compositeSource {
	cs.reflection = reflection
	return cs
}

func (cs compositeSource) ListServices() ([]string, error) {
	return cs.reflection.ListServices()
}

func (cs compositeSource) FindSymbol(fullyQualifiedName string) (desc.Descriptor, error) {
	descriptor, err := cs.reflection.FindSymbol(fullyQualifiedName)
	if err == nil {
		logger.Debugf("%s find symbol by reflection : %v", loggerHeader, descriptor)
		return descriptor, nil
	}
	return cs.file.FindSymbol(fullyQualifiedName)
}

func (cs compositeSource) AllExtensionsForType(typeName string) ([]*desc.FieldDescriptor, error) {
	exts, err := cs.reflection.AllExtensionsForType(typeName)
	if err != nil {
		return cs.file.AllExtensionsForType(typeName)
	}
	tags := make(map[int32]bool)
	for _, ext := range exts{
		tags[ext.GetNumber()] = true
	}
	fileExts, err := cs.file.AllExtensionsForType(typeName)
	if err != nil {
		return exts, nil
	}
	for _, ext := range fileExts {
		if !tags[ext.GetNumber()] {
			exts = append(exts, ext)
		}
	}
	return exts, nil
}
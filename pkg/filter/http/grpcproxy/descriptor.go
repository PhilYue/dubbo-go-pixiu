package grpcproxy

import (
	"context"
	"fmt"
	"github.com/apache/dubbo-go-pixiu/pkg/logger"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Descriptor struct {

	sf bool

	descSource DescriptorSource

	//reflSource *serverSource
	fileSource *fileSource

}

func (dr *Descriptor) initialize (cfg *Config) *Descriptor {
	// once
	if dr.sf {
		return dr
	}

	dr.initDescriptorSource(cfg)

	dr.sf = true
	return dr
}

func (dr *Descriptor) GetDescriptor (cfg *Config, refCtx context.Context, cc *grpc.ClientConn) DescriptorSource {

	// none, local, remote, compose
	switch cfg.DescriptorSourceStrategy {
	case "local":
		dr.initFileDescriptorSource(cfg)
		return dr.descSource
	case "remote":
		 var refClient *grpcreflect.Client

		 // server descriptor
		 refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))

		 ss := &serverSource{client: refClient}
		 //ds.initServerDescriptorSource(cfg)
		 var ds DescriptorSource = ss

		 //return &ds
		return ds
	case "auto", " ", "compose":
		 //ds.initComposeDescriptorSource(cfg)
		cs := &compositeSource{}
		cs.reflection = *(dr.getServerDescriptorSource(refCtx, cc))
		cs.file = dr.fileSource
		var source DescriptorSource = cs
		dr.descSource = source
		return source
	default:
		//
		logger.Infof("%s grpc descriptor source not init by strategy : ", loggerHeader, cfg.DescriptorSourceStrategy)
	}

	return dr.descSource
}

func (dr *Descriptor) initDescriptorSource (cfg *Config) *Descriptor {

	// none, local, remote, compose
	switch cfg.DescriptorSourceStrategy {
	case "local":
		return dr.initFileDescriptorSource(cfg)
	case "remote":
		return dr.initServerDescriptorSource(cfg)
	case "auto", " ", "compose":
		return dr.initComposeDescriptorSource(cfg)
	default:
		//
		logger.Infof("%s grpc descriptor source not init by strategy : ", loggerHeader, cfg.DescriptorSourceStrategy)
	}

	return dr
}

func (dr *Descriptor) initComposeDescriptorSource (cfg *Config) *Descriptor {

	//descriptor, err := initFileSource(cfg)
	//
	//if err != nil {
	//	//panic()
	//	logger.Errorf("%s init gRPC descriptor error, ", loggerHeader, err)
	//	return ds
	//}
	//
	//var source DescriptorSource = descriptor

	//var source DescriptorSource = &compositeSource{*ds.descSource, ds.fileSource}
	//
	//ds.descSource = source

	return dr
}

func (dr *Descriptor) initServerDescriptorSource (cfg *Config) *Descriptor {

	//ss := &serverSource{}
	//var dss DescriptorSource = ss

	return dr
}

func (dr *Descriptor) getServerDescriptorSource (refCtx context.Context, cc *grpc.ClientConn) *DescriptorSource {

	var refClient *grpcreflect.Client

	// server descriptor
	refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))

	ss := &serverSource{client: refClient}
	//ds.initServerDescriptorSource(cfg)
	var dss DescriptorSource = ss
	return &dss
}

func (dr *Descriptor) initFileDescriptorSource (cfg *Config) *Descriptor {

	if dr.fileSource != nil {
		return dr
	}

	descriptor, err := loadFileSource(cfg)

	if err != nil {
		logger.Errorf("%s init gRPC descriptor error, ", loggerHeader, err)
		return dr
	}

	dr.fileSource = descriptor

	return dr
}

func loadFileSource(cfg *Config) (*fileSource, error) {

	gc := cfg

	cur := gc.Path
	if !filepath.IsAbs(cur) {
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}
		cur = filepath.Dir(ex) + string(os.PathSeparator) + gc.Path
	}

	logger.Infof("%s load proto files from %s", loggerHeader, cur)

	fileLists := make([]string, 0)
	items, err := ioutil.ReadDir(cur)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if !item.IsDir() {
			sp := strings.Split(item.Name(), ".")
			length := len(sp)
			if length >= 2 && sp[length-1] == "proto" {
				fileLists = append(fileLists, item.Name())
			}
		}
	}

	if err != nil {
		return nil, err
	}

	importPaths := []string{gc.Path}

	fileNames, err := protoparse.ResolveFilenames(importPaths, fileLists...)
	if err != nil {
		return nil, err
	}
	p := protoparse.Parser{
		ImportPaths:           importPaths,
		InferImportPaths:      len(importPaths) == 0,
		IncludeSourceCodeInfo: true,
	}
	fds, err := p.ParseFiles(fileNames...)
	if err != nil {
		return nil, fmt.Errorf("could not parse given files: %v", err)
	}

	fsrc.files = make(map[string]*desc.FileDescriptor)
	for _, fd := range fds {
		name := fd.GetName()
		fsrc.files[name] = fd
	}

	return &fsrc, nil
}
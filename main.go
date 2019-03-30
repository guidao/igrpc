package main

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/fullstorydev/grpcurl"
	descpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/juju/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("addr is need")
		return
	}
	rt := &Runtime{
		addr:    os.Args[1],
		ctx:     NewGContext(),
		factory: NewFactory(),
	}
	rt.dial()
	for {
		input := prompt.Input("igrpc> ", rt.complete, prompt.OptionHistory(nil))
		if input == "quit" || input == "exit" {
			break
		}
		cmd := strings.Split(input, " ")
		if len(cmd) <= 0 {
			continue
		}
		switch cmd[0] {
		case "list":
			rt.list(cmd[1:])
		case "set":
			rt.set(cmd[1:])
		case "desc":
			rt.desc(cmd[1:])
		case "call":
			rt.call(cmd[1:])
		}
	}
}

type Runtime struct {
	addr       string
	cc         *grpc.ClientConn
	refClient  *grpcreflect.Client
	descSource grpcurl.DescriptorSource
	ctx        *GContext
	factory    Factory
}

func (this *Runtime) complete(d prompt.Document) []prompt.Suggest {
	cmd := this.factory.Match(d.Text)
	if cmd == nil {
		return prompt.FilterHasPrefix(this.suggestCommand(), d.GetWordBeforeCursor(), true)
	}
	return prompt.FilterHasPrefix(cmd.Suggest(this.ctx), d.GetWordBeforeCursor(), true)
}

func (this *Runtime) dial() error {
	if this.cc != nil && this.cc.GetState() != connectivity.TransientFailure {
		return nil
	}
	if this.addr == "" {
		return errors.New("remote server addr is empty")
	}
	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	var creds credentials.TransportCredentials
	cc, err := grpcurl.BlockingDial(ctx, "tcp", this.addr, creds, opts...)
	if err != nil {
		return errors.Trace(err)
	}
	this.refClient = grpcreflect.NewClient(context.Background(),
		reflectpb.NewServerReflectionClient(cc))
	this.descSource = grpcurl.DescriptorSourceFromServer(context.Background(), this.refClient)
	this.cc = cc

	this.setMetadata()
	return nil
}

func (this *Runtime) suggestCommand() []prompt.Suggest {
	return []prompt.Suggest{
		{
			Text:        "list",
			Description: "列出所有方法",
		},
		{
			Text:        "call",
			Description: "调用方法(json)",
		},
		{
			Text:        "desc",
			Description: "获取详细信息",
		},
	}
}

func (this *Runtime) list(args []string) {
	if len(args) <= 0 {
		svcs, err := this.refClient.ListServices()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, svc := range svcs {
			fmt.Println(svc)
		}
	} else {
		name := args[0]
		methods, err := grpcurl.ListMethods(this.descSource, name)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, m := range methods {
			fmt.Println(m)
		}
	}

}

func (this *Runtime) set(args []string) {
	if len(args) <= 0 {
		return
	}
	if args[0] == "addr" {
		this.addr = args[1]
	}
}

func (this *Runtime) desc(args []string) {
	if len(args) <= 0 {
		return
	}
	symbol := args[0]
	dsc, err := this.descSource.FindSymbol(symbol)
	if err != nil {
		fmt.Println(err)
		return
	}
	fqn := dsc.GetFullyQualifiedName()
	switch d := dsc.(type) {
	case *desc.MessageDescriptor:
		parent, ok := d.GetParent().(*desc.MessageDescriptor)
		if ok {
			if d.IsMapEntry() {
				for _, f := range parent.GetFields() {
					if f.IsMap() && f.GetMessageType() == d {
						// found it: describe the map field instead
						dsc = f
						break
					}
				}
			} else {
				// see if it's a group
				for _, f := range parent.GetFields() {
					if f.GetType() == descpb.FieldDescriptorProto_TYPE_GROUP &&
						f.GetMessageType() == d {
						// found it: describe the map field instead
						dsc = f
						break
					}
				}
			}

		}
	}
	txt, err := grpcurl.GetDescriptorText(dsc, this.descSource)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fqn, txt)
}

func (this *Runtime) call(args []string) {
	if len(args) < 2 {
		fmt.Println("method and data is need")
		return
	}
	this.dial()
	rf, formatter, err := grpcurl.RequestParserAndFormatterFor(grpcurl.Format("json"),
		this.descSource, false, true, strings.NewReader(args[1]))
	if err != nil {
		fmt.Println(err)
		return
	}

	h := grpcurl.NewDefaultEventHandler(os.Stdout, this.descSource, formatter, true)
	err = grpcurl.InvokeRPC(context.Background(), this.descSource, this.cc, args[0],
		nil, h, rf.Next)
	if err != nil {
		fmt.Println(err)
		return
	}
}

//得到元数据
func (this *Runtime) setMetadata() {
	//1. 获取service
	svcs, err := this.refClient.ListServices()
	if err != nil {
		return
	}
	this.ctx.Add("service", svcs)
	//2. 获取methods
	for _, svc := range svcs {
		if strings.HasPrefix(svc, "grpc.reflection") {
			continue
		}
		methods, err := grpcurl.ListMethods(this.descSource, svc)
		if err != nil {
			continue
		}
		this.ctx.Add("method", methods)
		//TODO parse argument
	}
}

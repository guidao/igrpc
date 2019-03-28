package main

import (
	"github.com/c-bata/go-prompt"
	"strings"
)

type GContext struct {
	service []prompt.Suggest
	methods []prompt.Suggest
	args    []prompt.Suggest
}

func (this *GContext) Add(t string, vals []string) {
	switch t {
	case "service":
		this.service = append(this.service, this.suggest(vals)...)
	case "method":
		this.service = append(this.methods, this.suggest(vals)...)
	case "arg":
		this.service = append(this.args, this.suggest(vals)...)
	}
}

func (this *GContext) suggest(vals []string) []prompt.Suggest {
	ret := make([]prompt.Suggest, len(vals))
	for _, v := range vals {
		ret = append(ret, prompt.Suggest{
			Text: v,
		})
	}
	return ret
}

type Commander interface {
	Name() string
	Suggest(ctx *GContext) []prompt.Suggest
}

type Factory interface {
	Match(input string) Commander
}

type factoryCmd struct {
}

func (this *factoryCmd) Match(input string) Commander {
	cmd := newCommand(input)
	name := cmd.Name()
	if name == "list" || name == "call" || name == "desc" {
		return cmd
	}
	return nil
}

type baseCmd struct {
	raw  string
	name string
	args []string
}

func newCommand(input string) *baseCmd {
	args := strings.Split(input, " ")
	if len(args) <= 0 {
		return nil
	}
	return &baseCmd{
		raw:  input,
		name: args[0],
		args: args[1:],
	}
}

func (this *baseCmd) Name() string {
	return this.name
}

func (this *baseCmd) Suggest(ctx *GContext) []prompt.Suggest {
	if this.name == "list" {
		return ctx.service
	}

	if this.name == "desc" {
		out := make([]prompt.Suggest, 0)
		out = append(out, ctx.service...)
		out = append(out, ctx.methods...)
		out = append(out, ctx.args...)
		return out
	}
	if this.name == "call" {
		return ctx.methods
	}
	return nil
}

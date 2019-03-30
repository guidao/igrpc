package main

import (
	"bytes"
	"github.com/c-bata/go-prompt"
	"strings"
	"text/scanner"
)

type GContext struct {
	service []prompt.Suggest
	methods []prompt.Suggest
	args    []prompt.Suggest
}

func NewGContext() *GContext {
	return &GContext{}
}

func (this *GContext) Add(t string, vals []string) {
	switch t {
	case "service":
		this.service = this.reduceDup(append(this.service, this.suggest(vals)...))
	case "method":
		this.methods = this.reduceDup(append(this.methods, this.suggest(vals)...))
	case "arg":
		this.args = this.reduceDup(append(this.args, this.suggest(vals)...))
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

func (this *GContext) reduceDup(vals []prompt.Suggest) []prompt.Suggest {
	m := make(map[string]bool)
	ret := make([]prompt.Suggest, 0)
	for _, v := range vals {
		if m[v.Text] {
			continue
		}
		m[v.Text] = true
		if v.Text == "" {
			continue
		}
		ret = append(ret, v)
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

func NewFactory() Factory {
	return &factoryCmd{}
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
	scan scanner.Scanner
	raw  string
	name string
	args []string
}

func newCommand(input string) *baseCmd {
	args := strings.Split(input, " ")
	if len(args) <= 0 {
		return nil
	}
	cmd := &baseCmd{
		raw:  input,
		name: args[0],
		args: args[1:],
	}
	//cmd.init()
	return cmd
}

func (this *baseCmd) init() {
	this.scan.Init(strings.NewReader(this.raw))
	buf := bytes.NewBufferString("")
	for tok := this.scan.Next(); tok != scanner.EOF; tok = this.scan.Next() {
		if this.scan.Whitespace&(1<<uint(tok)) != 0 {
			if buf.String() == "" {
				continue
			}
		}
		buf.WriteRune(tok)
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

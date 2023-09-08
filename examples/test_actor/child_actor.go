package main

import (
	"fmt"

	cherryActor "github.com/cherry-game/cherry/net/actor"
)

type childActor struct {
	cherryActor.Base
}

func (p *childActor) OnInit() {
	fmt.Println("[childActor] Execute OnInit()")

	p.Remote().Register("hello", p.hello)
}

func (p *childActor) hello() {
	text := "[childActor] Call hello()"
	fmt.Println(text)
}

func (*childActor) OnStop() {
}

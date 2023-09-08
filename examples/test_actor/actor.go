package main

import (
	"fmt"

	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryActor "github.com/cherry-game/cherry/net/actor"
)

type actor struct {
	cherryActor.Base
}

func (*actor) AliasID() string {
	return "parentActor"
}

func (p *actor) OnInit() {
	fmt.Println("[actor] Execute OnInit()")

	childActorID := "1"
	p.Child().Create(childActorID, &childActor{})

	targetPath := cherryFacade.NewChildPath("", p.AliasID(), childActorID)
	targetFuncName := "hello"

	p.CallWait(targetPath, targetFuncName, nil, nil)
	//fmt.Println(reply)
}

func (*actor) OnStop() {
}

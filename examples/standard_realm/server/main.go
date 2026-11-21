package main

import (
	"hero-armory-sdk"
	"hero-armory-sdk/vanilla"
)

type standardRealmPlugin struct {
	sdk.BaseServerPlugin
}

func (standardRealmPlugin) Init() {
	vanilla.RegisterAll()
}

func (standardRealmPlugin) OnNPCInteract(ev sdk.NPCInteractEvent) (sdk.NPCDialogResult, bool) {
	return vanilla.HandleNPCInteract(ev)
}

func init() {
	sdk.RegisterServer(standardRealmPlugin{})
}

func main() {}

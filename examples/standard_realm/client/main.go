package main

import (
	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/tactical"
)

type standardClientPlugin struct {
	sdk.BaseClientPlugin
}

func (standardClientPlugin) OnPacket(channel uint16, data []byte) {
	tactical.HandleHUDPacket(channel, data)
}

func (standardClientPlugin) OnUIEvent(ev sdk.UIClickEvent) {
	tactical.HandleUIEvent(ev)
}

func init() {
	sdk.RegisterClient(standardClientPlugin{})
}

func main() {}

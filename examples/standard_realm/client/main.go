package main

import (
	"hero-armory-sdk"
)

type standardClientPlugin struct {
	sdk.BaseClientPlugin
}

func init() {
	sdk.RegisterClient(standardClientPlugin{})
}

func main() {}

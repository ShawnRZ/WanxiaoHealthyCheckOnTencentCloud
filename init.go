package main

import (
	"github.com/BurntSushi/toml"
	"github.com/FNDHSTD/logor"
)

var logger *logor.ConsoleLogger
var settings struct {
	Users []User `toml:"users"`
}

func init() {
	logger = logor.NewConsoleLogger("debug")
	toml.DecodeFile("./settings.toml", &settings)
}

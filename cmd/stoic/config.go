package main

import (
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("debug", false)
	viper.SetDefault("root", "")

	viper.SetEnvPrefix("stoic")
	viper.BindEnv("debug")
	viper.BindEnv("root")

	if viper.GetBool("debug") {
		jww.SetStdoutThreshold(jww.LevelDebug)
	}
}

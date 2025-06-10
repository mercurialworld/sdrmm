package main

import (
	"fmt"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/drm"
	"rustlang.pocha.moe/sdrmm/utils"
)

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	utils.HandleError(err)
}

func main() {
	readConfig()

	cmd, res, extra := drm.Parse() // returns json

	fmt.Printf("Command type is %s\n", cmd)
	fmt.Printf("%s\n", res)

	if extra != nil {
		fmt.Print(extra)
	}

	// do something with json (serialize, filter, etc)

	// add to database
}

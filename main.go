package main

import (
	"rustlang.pocha.moe/sdrmm/config"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/parser"
	"rustlang.pocha.moe/sdrmm/utils"
)

func main() {
	config.ReadConfig()
	config := config.GetConfig()

	db := database.InitializeDB()

	cmd, args, err := parser.Parse() // returns json
	utils.PanicOnError(err)

	RunCommands(cmd, args, config, db)

	database.CloseDB(db)
}

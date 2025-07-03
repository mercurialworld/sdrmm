package main

import (
	"rustlang.pocha.moe/sdrmm/config"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/parser"
	"rustlang.pocha.moe/sdrmm/utils"
)

func main() {
	config.ReadConfig()

	db := database.InitializeDB()

	cmd, args, err := parser.Parse() // returns json
	utils.PanicOnError(err)

	RunCommands(cmd, args, db)

	database.CloseDB(db)
}

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

	db := database.DRMDatabase{DB: database.InitializeDB()}

	runner := Runner{config: config, database: db}

	cmd, args, err := parser.Parse()
	utils.PanicOnError(err)

	runner.RunCommands(cmd, args)

	db.CloseDB()
}

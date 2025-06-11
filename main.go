package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/database"
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

	db := database.InitializeDB()

	cmd, res, extra, err := drm.Parse() // returns json
	utils.HandleError(err)

	fmt.Printf("Command type is %s\n", cmd)
	fmt.Printf("%s\n", res)

	if extra != nil {
		fmt.Print(extra)
	}

	switch cmd {
	// add to queue
	case "add":
		var mapToQueue drm.MapData
		json.Unmarshal(res, &mapToQueue)

		mapToQueue, err := FilterMap(mapToQueue, db)
		if err != nil {
			fmt.Printf("{\"message\": %s}", err)
		} else {
			addKeyArgs := mapToQueue.BsrKey

			if username, ok := extra.(string); ok && username != "" {
				addKeyArgs += "?user=" + username
			}

			fmt.Printf("%s", drm.RequestDRM("addKey", addKeyArgs))
		}

	// ban/unban map
	case "ban":
		var mapToBan drm.MapData
		json.Unmarshal(res, &mapToBan)
		database.BanMap(mapToBan.BsrKey, mapToBan.Hash, db)
	case "unban":
		var mapToUnban drm.MapData
		json.Unmarshal(res, &mapToUnban)
		database.UnbanMap(mapToUnban.BsrKey, db)

	// anything else
	default:
		fmt.Printf("%s", res)
	}

	database.CloseDB(db)
}

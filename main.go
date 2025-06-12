package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/database"
	"rustlang.pocha.moe/sdrmm/drm"
	"rustlang.pocha.moe/sdrmm/parser"
	"rustlang.pocha.moe/sdrmm/utils"
)

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, set defaults

			// DRM
			viper.SetDefault("drm.url", "http://localhost")
			viper.SetDefault("drm.port", 13337)

			// BSR limits that aren't NPS or NJS
			viper.SetDefault("bsr.request-limit", 0)
			viper.SetDefault("bsr.newer-than", "2018-05-08")
			viper.SetDefault("bsr.map-age", 0)
			viper.SetDefault("bsr.min-length", 0)
			viper.SetDefault("bsr.max-length", 0)

			// NPS limits
			viper.SetDefault("nps.min", 0)
			viper.SetDefault("nps.max", 0)

			// NJS limits
			viper.SetDefault("njs.min", 0)
			viper.SetDefault("njs.max", 0)

			// Write the defaults
			viper.SafeWriteConfig()
		} else {
			utils.PanicOnError(err)
		}
	}
}

func main() {
	readConfig()

	db := database.InitializeDB()

	cmd, res, extra, err := parser.Parse() // returns json
	utils.PanicOnError(err)

	fmt.Printf("Command type is %s\n", cmd)

	if extra != nil {
		fmt.Printf("%s\n", extra)
	}

	switch cmd {
	// add to queue
	case "add":
		var mapToQueue drm.MapData
		json.Unmarshal(res, &mapToQueue)

		if username, ok := extra.(string); ok && username != "" {

			mapToQueue, err := FilterMap(mapToQueue, username, db)
			if err != nil {
				fmt.Printf("{\"message\": \"%s\"}", err)
			} else {
				addKeyArgs := mapToQueue.BsrKey + "?user=" + username
				drm.RequestDRM("addKey", addKeyArgs)

				requestLimit := viper.GetInt("request-limit")

				// increment request counter
				if requestLimit != 0 {
					userNumRequests := database.GetUserRequests(username, db)
					userNumRequests++

					database.SetUserRequests(username, userNumRequests, db)
				}
			}

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

	// new session
	case "new":
		// remove all request limits
		database.ClearRequestLimits(db)
		// remove the queue (no caching)
		drm.RequestDRM("queue", "clear")
		fmt.Println("\"message\": \"New session created\"")

	// anything else
	default:
		fmt.Printf("%s", res)
	}

	database.CloseDB(db)
}

package config

import (
	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/utils"
)

func ReadConfig() {
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
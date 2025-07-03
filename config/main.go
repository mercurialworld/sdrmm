package config

import (
	"time"

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

func GetConfig() BSRConfig {
	minLength := viper.GetInt("min-length")
	maxLength := viper.GetInt("max-length")
	noteLimits := NoteLimits{
		MinNJS: viper.GetFloat64("njs.min"),
		MaxNJS: viper.GetFloat64("njs.max"),
		MinNPS: viper.GetFloat64("nps.min"),
		MaxNPS: viper.GetFloat64("nps.max"),
	}
	requestLimit := viper.GetInt("bsr.request-limit")
	newerThan, err := time.Parse("2006-01-02", viper.GetString("bsr.newer-than"))
	if err != nil {
		newerThan, _ = time.Parse("2006-01-02", "2000-01-01")
	}
	mapAge := viper.GetInt("bsr.map-age")

	return BSRConfig{
		MinLength:    minLength,
		MaxLength:    maxLength,
		NoteLimits:   noteLimits,
		RequestLimit: requestLimit,
		NewerThan:    newerThan,
		MapAge:       mapAge,
	}
}

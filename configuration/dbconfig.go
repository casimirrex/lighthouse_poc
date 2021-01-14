package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Configure() map[string]string {
	//product_config may be DEV, UAT, PROD, TEST, TESTENV etc.
	product_config := os.Getenv("LANTERN")

	if product_config == "" {
		product_config = "DEV"
	}

	if product_config == "DEV" {
		viper.SetConfigName("config.development")
		viper.SetConfigType("json")
	} else if product_config == "UAT" {
		viper.SetConfigName("config.uat")
		viper.SetConfigType("json")
	} else if product_config == "PROD" {
		viper.SetConfigName("config.production")
		viper.SetConfigType("json")
	} else if product_config == "TEST" {
		viper.SetConfigName("config.test")
		viper.SetConfigType("json")
	} else if product_config == "TESTENV" {
		viper.SetConfigName("config1.test")
		viper.SetConfigType("env")
	}
	viper.AddConfigPath("lantern/configuration/conf/")
	viper.AddConfigPath("configuration/conf/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	/* store key/value in a map*/
	config := make(map[string]string)
	config["hostname"] = viper.GetString("HOSTNAME")
	config["port"] = viper.GetString("PORT")
	config["username"] = viper.GetString("USERNAME")
	config["password"] = viper.GetString("PASSWORD")
	config["serverurl"] = viper.GetString("SERVERURL")
	config["dbname"] = viper.GetString("DBNAME")
	config["recordsize"] = viper.GetString("RECORDSIZE")
	return config
}

//https://dev.to/ilyakaznacheev/a-clean-way-to-pass-configs-in-a-go-application-1g64

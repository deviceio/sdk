package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/alecthomas/kingpin"
	_ "github.com/deviceio/agent/resources"
	_ "github.com/deviceio/agent/resources/filesystem"
	_ "github.com/deviceio/agent/resources/process"
	"github.com/deviceio/agent/transport"
	"github.com/deviceio/shared/logging"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cliID = kingpin.Flag("id", "Override the id used for agent identification").Action(func(pc *kingpin.ParseContext) error {
		viper.Set("id", lookupKingpinValue("id", pc))
		return nil
	}).String()

	cliInsecure = kingpin.Flag("insecure", "Allow self signed certificates on the hub transport").Default("false").Action(func(pc *kingpin.ParseContext) error {
		viper.Set("transport.allow_self_signed", lookupKingpinValue("insecure", pc))
		return nil
	}).Bool()

	cliTransportHost = kingpin.Flag("host", "The transport host to connect to").Action(func(pc *kingpin.ParseContext) error {
		viper.Set("transport.host", lookupKingpinValue("host", pc))
		return nil
	}).String()

	cliTransportPort = kingpin.Flag("port", "The transport host port to use").Action(func(pc *kingpin.ParseContext) error {
		viper.Set("transport.port", lookupKingpinValue("port", pc))
		return nil
	}).Int()
)

func main() {
	rand.Seed(time.Now().UnixNano()) //very important

	homedir, err := homedir.Dir()

	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(fmt.Sprintf("%v/.deviceio/agent/", homedir))
	viper.AddConfigPath("$HOME/.deviceio/agent/")
	viper.AddConfigPath("/etc/deviceio/agent/")
	viper.AddConfigPath("/opt/deviceio/agent/")
	viper.AddConfigPath("c:/PROGRA~1/deviceio/agent/")
	viper.AddConfigPath("c:/ProgramData/deviceio/agent/")
	viper.AddConfigPath(".")

	viper.SetDefault("tags", []string{})
	viper.SetDefault("transport.host", "127.0.0.1")
	viper.SetDefault("transport.port", 8975)
	viper.SetDefault("transport.allow_self_signed", false)

	viper.BindEnv("id", "DEVICEIO_AGENT_ID")
	viper.BindEnv("tags", "DEVICEIO_AGENT_TAGS")
	viper.BindEnv("transport.host", "DEVICEIO_AGENT_TRANSPORT_HOST")
	viper.BindEnv("transport.port", "DEVICEIO_AGENT_TRANSPORT_PORT")
	viper.BindEnv("transport.allow_self_signed", "DEVICEIO_AGENT_TRANSPORT_INSECURE")

	viper.ReadInConfig()
	kingpin.Parse() // cli MUST be parsed after viper, to overwrite any values

	log.Println("Using configuration file: ", viper.ConfigFileUsed())

	transport.NewConnection(&logging.DefaultLogger{}).Dial(&transport.ConnectionOpts{
		ID:   viper.GetString("id"),
		Tags: viper.GetStringSlice("tags"),
		TransportAllowSelfSigned: viper.GetBool("transport.allow_self_signed"),
		TransportHost:            viper.GetString("transport.host"),
		TransportPort:            viper.GetInt("transport.port"),
	})

	<-make(chan bool)
}

func lookupKingpinValue(name string, pc *kingpin.ParseContext) *string {
	for _, el := range pc.Elements {
		switch el.Clause.(type) {
		case *kingpin.ArgClause:
			if (el.Clause).(*kingpin.ArgClause).Model().Name == name {
				return el.Value
			}
		case *kingpin.FlagClause:
			if (el.Clause).(*kingpin.FlagClause).Model().Name == name {
				return el.Value
			}
		}
	}

	def := ""
	return &def
}

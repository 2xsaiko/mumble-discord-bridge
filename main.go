package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"layeh.com/gumble/gumble"
	_ "layeh.com/gumble/opus"
)

var YBConfig *YammerConfig

func main() {
	godotenv.Load()

	mumbleAddr := flag.String("mumble-address", lookupEnvOrString("MUMBLE_ADDRESS", ""), "MUMBLE_ADDRESS, mumble server address, example example.com")
	mumblePort := flag.Int("mumble-port", lookupEnvOrInt("MUMBLE_PORT", 64738), "MUMBLE_PORT mumble port")
	mumbleUsername := flag.String("mumble-username", lookupEnvOrString("MUMBLE_USERNAME", "discord-bridge"), "MUMBLE_USERNAME, mumble username")
	mumblePassword := flag.String("mumble-password", lookupEnvOrString("MUMBLE_PASSWORD", ""), "MUMBLE_PASSWORD, mumble password, optional")
	mumbleInsecure := flag.Bool("mumble-insecure", lookupEnvOrBool("MUMBLE_INSECURE", false), "mumble insecure,  env alt MUMBLE_INSECURE")

	discordToken := flag.String("discord-token", lookupEnvOrString("DISCORD_TOKEN", ""), "DISCORD_TOKEN, discord bot token")
	discordGID := flag.String("discord-gid", lookupEnvOrString("DISCORD_GID", ""), "DISCORD_GID, discord gid")
	discordCID := flag.String("discord-cid", lookupEnvOrString("DISCORD_CID", ""), "DISCORD_CID, discord cid")

	flag.Parse()
	log.Printf("app.config %v\n", getConfig(flag.CommandLine))

	if *mumbleAddr == "" {
		log.Fatalln("missing mumble address")
	}
	if *mumbleUsername == "" {
		log.Fatalln("missing mumble username")
	}

	if *discordToken == "" {
		log.Fatalln("missing discord bot token")
	}
	if *discordGID == "" {
		log.Fatalln("missing discord gid")
	}
	if *discordCID == "" {
		log.Fatalln("missing discord cid")
	}

	// DISCORD Setup

	discord, err := discordgo.New("Bot " + *discordToken)
	if err != nil {
		log.Println(err)
		return
	}

	// Open Websocket
	discord.LogLevel = 2
	discord.StateEnabled = true
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates)
	// register handlers
	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	discord.AddHandler(guildCreate)
	err = discord.Open()
	if err != nil {
		log.Println(err)
		return
	}
	defer discord.Close()

	log.Println("Discord Bot Connected")
	config := gumble.NewConfig()
	config.Username = *mumbleUsername
	config.Password = *mumblePassword
	config.AudioInterval = time.Millisecond * 10

	YBConfig = &YammerConfig{
		Config:         config,
		MumbleAddr:     *mumbleAddr + ":" + strconv.Itoa(*mumblePort),
		MumbleInsecure: *mumbleInsecure,
		ActiveConns:    make(map[string]chan bool),
	}

	//go startBridge(discord, *discordGID, *discordCID, config, *mumbleAddr+":"+strconv.Itoa(*mumblePort), *mumbleInsecure, die)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Println("Bot shutting down")
}

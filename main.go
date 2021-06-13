package main

import (
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"context"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	token string
	google_cloud_token string
	fileName string
	sliceData []string
	buffer = make([][]byte, 0)
	playing bool = false
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}


func main() {

	//create a new discord session with the provided token
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Error creating Discord session: ", err)
	}

	//read input file and split it and set the seed
	rand.Seed(time.Now().UnixNano())
	fileName = "fn.txt"
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal("Could not read the file: ", err)
	}

	//Data := strings.Split(strings.ReplaceAll(string(fileBytes), "\r\n", "\n"), "\n")
	sliceData = strings.Split(string(fileBytes), ";")

	//register ready func as a callback for the ready events
	bot.AddHandler(ready)

	//register guildCreate as a callback for the guildCreate events
	bot.AddHandler(guildCreate)

	//register the messageCreate func as a callback for MessageCreate events
	bot.AddHandler(messageCreate)

	// we need information about guilds (which includes their channels),
	// messages and voice states
	bot.Identify.Intents = discordgo.IntentsGuilds| discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	//Open a websocket connection and begin listening
	err = bot.Open()
	if err != nil {
		log.Fatalln("Error while opening Discord connection: ", err)
	}

	//Wait for CTRL-C or other term signal
	fmt.Println("Bot is running. CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	//close the discord session
	bot.Close()
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err_r error) {

	if playing {
		return nil
	}

	rand_num := rand.Int()%len(sliceData)
	msg_string := sliceData[rand_num]
	file_name := strconv.Itoa(rand_num)+".mp3"
	var err error
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS","INSERT YOUR GOOGLE CLOUD CREDENTIALS HERE")
	if Exists(file_name) {
		log.Println("file found: ", file_name)
		if err != nil {
			fmt.Println("Error opening mp3 file :", err)
			return err
		}
	} else {
		log.Println("file not found: ", file_name)
		ctx := context.Background()

		client, err := texttospeech.NewClient(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		// Perform the text-to-speech request on the text input with the selected
		// voice parameters and audio file type.
		req := texttospeechpb.SynthesizeSpeechRequest{
			// Set the text input to be synthesized.
			Input: &texttospeechpb.SynthesisInput{
				InputSource: &texttospeechpb.SynthesisInput_Text{Text: msg_string},
			},
			// Build the voice request, select the language code and the SSML
			// voice gender.
			Voice: &texttospeechpb.VoiceSelectionParams{
				LanguageCode: "ru-RU",
				SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
				Name: "ru-RU-Wavenet-D",
			},
			// Select the type of audio file you want returned.
			AudioConfig: &texttospeechpb.AudioConfig{
				AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			},
		}

		resp, err := client.SynthesizeSpeech(ctx, &req)
		if err != nil {
			log.Fatal(err)
		}

		// The resp's AudioContent is binary.
		file_name = strconv.Itoa(rand_num) + ".mp3"
		err = ioutil.WriteFile(file_name, resp.AudioContent, 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Audio content written to file: %v\n", file_name)

		log.Println(file_name)

	}

	//encode mp3 to dca
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 128
	options.Application = "lowdelay"

	encodeSession, err := dca.EncodeFile(file_name, options)
	if err != nil {
		log.Fatal(err)
	}
	defer encodeSession.Cleanup()


	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}
	playing = true

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	done := make(chan error)
	dca.NewStream(encodeSession, vc, done)
	err = <-done
	if err != nil && err != io.EOF {
		return
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specified amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	playing = false

	return nil

}

func stop(s *discordgo.Session, guildID, channelID string) (err error) {
	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)

	if err != nil {
		return
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specified amount of time before ending.
	time.Sleep(2*time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	playing = false

	return
}

func playurl(s *discordgo.Session, guildID, channelID, url string) (err error) {

	//Setting a new DCA session
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 128
	options.Application = "lowdelay"

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	//Connect to the server (a.k.a: do the magic trick)
	encodeSession, err := dca.EncodeFile(url, options)
	if err != nil {
		return
	}

	//And now, stream!
	done := make(chan error)
	dca.NewStream(encodeSession, vc, done)
	err = <-done
	if err != nil && err != io.EOF {
		return
	}

	return nil
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}


// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "TTS!")
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Hello! Type TTS!V for me to join your voice channel!")
			return
		}
	}
}

//this function is called every time a new message is created on any channel
//the bot has access to
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.

	if m.Author.ID == s.State.User.ID {
		return
	}
	msg_string := sliceData[rand.Int()%len(sliceData)]
	if strings.HasPrefix(m.Content, "TTS!") {
		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		//write a random paragraph from the text file to the text channel
		if m.Content == "TTS!P" {
			s.ChannelMessageSend(m.ChannelID, msg_string)
		}
		// join voice chat and tts a random paragraph from the text channel
		if strings.HasPrefix(m.Content, "TTS!V") {

			// Look for the message sender in that guild's current voice states.
			for _, vs := range g.VoiceStates {
				if vs.UserID == m.Author.ID {
					err = playSound(s, g.ID, vs.ChannelID)
					if err != nil {
						fmt.Println("Error while playing sound:", err)
					}

					return
				}
			}
		}

		//leave voice channel
		if strings.HasPrefix(m.Content, "TTS!L") {
			for _, vs := range g.VoiceStates {
				if vs.UserID == m.Author.ID {
					err = stop(s, g.ID, vs.ChannelID)
					if err != nil {
						fmt.Println("Error stopping sound:", err)
					}

					return
				}
			}
		}
	}

}

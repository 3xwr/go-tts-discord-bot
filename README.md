# Go TTS Discord Bot
A simple discord TTS bot using the [DiscordGo](https://github.com/bwmarrin/discordgo) library.

## Requirements
- go compiler
- gcc compiler
- ffmpeg.exe
- ffprobe.exe

## Build

This assumes you already have a working Go environment setup and that
DiscordGo is correctly installed on your system.

From within the bot folder, run the below command to compile the
bot.

```sh
go build
```

## Usage

```
Usage of ./bot:
  -t string
        Bot Token
```
You should also specify the location of your Google Cloud credentials in the code.

## Features
This bot uses [Google Cloud Text-to-Speech](https://cloud.google.com/text-to-speech) to convert a random paragraph from included text file into natural-sounding speech.

## Commands
| Command           | Description                                                   |
|-------------------|---------------------------------------------------------------|
| TTS!P             | write a random paragraph from the  file to the text channel   |
| TTS!V             | joins your current voice channel and reads a random paragraph |
| TTS!L             | leaves the voice channel                                      |

## TODO
- Rewrite code to improve readability
- Sharding support

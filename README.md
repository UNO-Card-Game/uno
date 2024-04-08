# UNO Multiplayer 

![UNO_Logo](https://github.com/mahimdashora/UNO-game/assets/60029463/4198de68-0a20-44ac-81a1-3cd0a459d29a)
# Uno Game Setup Guide

This guide provides step-by-step instructions to set up and run the Uno game.

# Installation
## Install [Task](https://taskfile.dev/installation/)
### Mac OS
```bash
brew install go-task
```

## Step 1: Define Player Names

Open the `main.go` file and define player names in the `players` slice.

```go
// Example:
// players := []string{"Player1", "Player2", "Player3"}
```
## Step 2: Build and run the server 
```bash
task build:server
```
```bash
task run:server
```
## Step 3: Connect to the Server
Install wscat package </br>
### Windows: 
- Install Node.js </br>
- Open Command Prompt as an administrator. 
- ```npm install -g wscat```
### Mac: 
- ```brew install node ```
- ```npm install -g wscat```
### Ubuntu:
- ```sudo apt-get update```
- ```sudo apt-get install nodejs npm```


To connect to the server port, use `wscat` with the following command:
```bash
wscat -c ws://localhost:8080/ws
```

- Replace `8080` with your specific port number if needed.
- Enter the player name correctly when prompted.
- Wait for your turn to play.
### Use commands 
1. ```playcard <cardIndex> <NewColor>``` to play and <NewColor> arg is valid only for WILD and DRAW_4 </br>
2. ```showcards``` to see cards in hand and their index </br>
3. ```topcard``` to see top card which was last played in the game deck</br>

# UNO Multiplayer 

![UNO_Logo](https://github.com/UNO-Card-Game/uno/blob/main/assets/UNO_Logo.png?raw=true)

# Uno Game Setup Guide

This guide provides step-by-step instructions to set up and run the Uno game.

# Description

This is a backend service for a UNO card game multiplayer written in Go and based on WebSockets.

# Installation

## Install [Task](https://taskfile.dev/installation/)
### Mac OS
```bash
brew install go-task
```

### pip
```bash
pip install go-task-bin

```
### npm
```bash
npm install -g @go-task/cli
```

### RHEL/Fedora
```bash
dnf install go-task
```

## Installation of the Server

### Native

1. Clone the repository:
```bash
git clone https://github.com/mahimdashora/UNO-game.git
cd UNO-game
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build and run the server:
```bash
task build:binary
task run:server
```

### Docker (Container)

1. Clone the repository:
```bash
git clone https://github.com/UNO-Card-Game/uno.git
cd UNO-game
```

2. Build the Docker image:
```bash
build:image
```

3. Run the Docker container:
```bash
docker run -p 8080:8080 uno-server:latest
```

## Step 2: Build and run the server
```bash
task build:server
```
```bash
task run:server
```

## Environment Variables

`PORT`: Set this environment variable to run the server on a specific port. Default is `8080`.

Example:
```bash
export PORT=8080
task run:server
```

Replace `8080` with your specific port number if needed.

## Test WebSockets with Postman

1. Open Postman and create a new WebSocket request.
2. Enter the WebSocket URL: `ws://localhost:8080/ws`
3. To create a Game Room lobby, use the following URL:
```plaintext
ws://localhost:8080/create?player_name=[NAME]&max_players=[MAX_PLAYER_COUNT]
```
Example:
```plaintext
ws://localhost:8080/create?player_name=Alice&max_players=2
```
4. To join a Game Room lobby, use the following URL:
```plaintext
ws://localhost:8080/join?player_name=Bob&room_id=1234
```
```plaintext
ws://localhost:8080/join?player_name=Bob&room_id=1234
```

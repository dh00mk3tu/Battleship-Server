# Battleship-Server

Welcome to the Battleship Server! This project is a multiplayer implementation of the classic Battleship game, built using Go. The server handles player connections, matchmaking, and game logic.

## Features

- **Matchmaking**: Players can find matches or create/join private rooms. [Working]
- **Ship Placement**: Players can place their ships on a 10x10 grid. [Yea this ain't wokring yet]
- **Battle Mode**: Once all ships are placed, the battle begins (game logic to be implemented). [Haven't even started yet]
- **Dynamic Rooms**: Create private rooms with unique IDs for private matches. [Working]

## How It Works

### Server Setup
The server listens on `localhost:8080` for incoming connections. Players connect via TCP and interact with the server using predefined commands.

### Commands
- `FIND_MATCH`: Find a random opponent.
- `CREATE_PRIVATE`: Create a private room.
- `JOIN_PRIVATE:<RoomID>`: Join a private room using the room ID.
- `PLACE_SHIP:<ShipName>:<StartCoord>:<Direction>`: Place a ship on the board.

### Game Flow
1. Players connect to the server and provide their name.
2. Players can choose to find a match or create/join a private room.
3. Once matched, players place their ships on the board.
4. When all ships are placed, the battle begins (to be implemented).

## Code Highlights

### Board Struct
The `Board` struct represents the 10x10 grid. Ships are placed on the grid, and the board is serialized for communication with the client.

### Player Struct
The `Player` struct holds player-specific data, including their connection, name, board, and opponent.

### Matchmaking
The server supports both random matchmaking and private rooms. Private rooms are identified by unique IDs generated using a random alphanumeric string.

### Ship Placement
Ships are placed on the board using the `PLACE_SHIP` command. The server validates the placement to ensure ships do not overlap or go out of bounds.

## Known Issues
- The board serialization method (`Serialize`) is not working as expected. The client may not parse the board correctly.
- The main game logic for the battle phase is yet to be implemented.

## Future Improvements
- Implement the battle logic.
- Enhance error handling and logging.
- Allow customization of ship types and counts. Somethig like a blitz mode.
- Improve the client-side experience.

## Getting Started

### Prerequisites
- Go 1.16 or later

### Running the Server
1. Clone the repository.
2. Navigate to the project directory.
3. Run the server:
   ```bash
   go run main.go
   ```
4. The server will start on `localhost:8080`.

### Connecting to the Server
Use the Battleship client.

## Acknowledgments
This project was inspired by the classic Battleship game and built with a focus on learning Go. Special thanks to random YouTube videos and my divine intuition for guidance lmao!

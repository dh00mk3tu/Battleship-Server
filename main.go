package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
* I honestly hadn't any idea how to proceed with deafults. Ideally in JS/TS I would've created a JSON object, or an interface.
* I went in blind with GO and assumed there will be classes but to my surprise, there are no classes so I had to resovle to structs.
* Not sure if this is the best way to do it, but I think it works hehe.
* I also don't like classes much and would often incline to functional programming, but whatever works.
 */

/*
* board struct is the 10x10 grid of map.
* x axist - A-J
* y axis - 1-10
* no diagnonl.
* I think that the board isn't rendering properly because I initally deffined the board after the player struct and the player
* struct is using the board struct but go is compiled lang soi doesnt matter
 */
type Board struct {
	Grid  [10][10]string
	Ships map[string]int
}

/*
* Player struct holds the foolowing data. Initially it just held the connection info. I added other things as I progressed.
* - conn: the connection to the player. Essentially the socket connection I guess.
* - name: the name of the player (user input on client)
* - board: the player's board. Your map of ships. The board rendering loop is not working idk why.
* - placedShips: a boolean to check if the player has placed all ships or not.
* - opponent: the opponent player with whom the player is matched with or is playing against
 */

type Player struct {
	conn        net.Conn
	name        string
	board       *Board
	placedShips bool
	opponent    *Player
}

// room struct holds players who are part of the room.
type Room struct {
	players []*Player
}

/*
* struct variable to gold ship names and their lenght.
* I should shorten the names to one letter because on client side entering the whole ship name is a annoying and
* I dont expect the end user to know what all ship types are availble in the first place plus spelling mistake.
* I can probably use a map for this but idk I am working on divine intuintion rn.
 */

var ShipsToPlace = []struct {
	Name   string
	Length int
}{
	{"Carrier", 5},
	{"Battleship", 4},
	{"Cruiser", 3},
	{"Submarine", 3},
	{"Destroyer", 2},
}

var (
	waitingPlayers []*Player
	privateRooms   = make(map[string]*Room)
	mu             sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server started on localhost:8080")

	// found this shi in a random go video on youtube
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("What is your name commander: "))

	/* so originally there was a problem here. I was using bufio to read the connection and I was using the same bufio to read the response string from rthe client
	* reader := bufio.NewReader(conn)
	* serverReader := bufio.NewReader(conn)
	* I was doing this I figured that this was causing a deadlock since both reader and serverReader were trying to read from the same conn
	 */

	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	// prefixing fleet commander to the name because cool and shi
	name = "Fleet Commander " + strings.TrimSpace(name)

	player := &Player{
		conn:  conn,
		name:  name,
		board: NewBoard(),
	}

	// this essentially is the game loop.
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(player.name, "disconnected")
			removePlayer(player)
			if player.opponent != nil {
				player.opponent.conn.Write([]byte("OPPONENT_DISCONNECTED\n"))
			}
			return
		}

		/* reading msg from the client. I have defined the following enums basically that have their own meanings
		* FIND_MATCH: client requests to find a match
		* CREATE_PRIVATE: client requests to create a private room/lobby
		* JOIN_PRIVATE: client requests to join a private room/lobby
		* PLACE_SHIP: client requests to place a ship on the board
		 */
		message = strings.TrimSpace(message)

		if message == "FIND_MATCH" {
			handleFindMatch(player)
		} else if message == "CREATE_PRIVATE" {
			handleCreatePrivate(player)
		} else if strings.HasPrefix(message, "JOIN_PRIVATE:") {
			roomID := strings.TrimPrefix(message, "JOIN_PRIVATE:")
			handleJoinPrivate(player, roomID)
		} else if strings.HasPrefix(message, "PLACE_SHIP:") {
			handlePlaceShip(player, message)
		}
	}
}

func handleFindMatch(player *Player) {
	mu.Lock()
	defer mu.Unlock()

	waitingPlayers = append(waitingPlayers, player)

	if len(waitingPlayers) >= 2 {
		player1 := waitingPlayers[0]
		player2 := waitingPlayers[1]
		waitingPlayers = waitingPlayers[2:]

		startMatch(player1, player2)
	}
}

func handleCreatePrivate(player *Player) {
	mu.Lock()
	defer mu.Unlock()

	roomID := generateRoomID()
	room := &Room{players: []*Player{player}}
	privateRooms[roomID] = room

	player.conn.Write([]byte("ROOM_CREATED:" + roomID + "\n"))
}

func handleJoinPrivate(player *Player, roomID string) {
	mu.Lock()
	defer mu.Unlock()

	room, exists := privateRooms[roomID]
	if !exists {
		player.conn.Write([]byte("INVALID_ROOM\n"))
		return
	}

	if len(room.players) >= 2 {
		player.conn.Write([]byte("INVALID_ROOM\n"))
		return
	}

	room.players = append(room.players, player)
	delete(privateRooms, roomID)

	player.conn.Write([]byte("JOIN_SUCCESS:" + roomID + "\n"))

	startMatch(room.players[0], room.players[1])
}

func startMatch(player1, player2 *Player) {
	player1.opponent = player2
	player2.opponent = player1

	player1.board = NewBoard()
	player2.board = NewBoard()

	player1.conn.Write([]byte(fmt.Sprintf("MATCH_FOUND: You are matched with %s\n", player2.name)))
	player2.conn.Write([]byte(fmt.Sprintf("MATCH_FOUND: You are matched with %s\n", player1.name)))

	time.Sleep(1 * time.Second)

	player1.conn.Write([]byte("PLACE_SHIPS\n"))
	player2.conn.Write([]byte("PLACE_SHIPS\n"))

	sendBoard(player1)
	sendBoard(player2)
}

func sendBoard(player *Player) {
	player.conn.Write([]byte("BOARD\n"))
	player.conn.Write([]byte(player.board.Serialize()))
}

func handlePlaceShip(player *Player, message string) {
	// message format: PLACE_SHIP:ShipName:StartCoord:Direction
	parts := strings.Split(message, ":")
	if len(parts) != 4 {
		player.conn.Write([]byte("INVALID_COMMAND\n"))
		return
	}

	shipName := parts[1]
	startCoord := parts[2]
	direction := strings.ToUpper(parts[3]) // horizontal/vertical

	err := player.board.PlaceShip(shipName, startCoord, direction)
	if err != nil {
		player.conn.Write([]byte("PLACE_ERROR: " + err.Error() + "\n"))
		return
	}

	sendBoard(player)

	// Check if all ships placed on bvaord
	if player.board.AllShipsPlaced() {
		player.placedShips = true
		player.conn.Write([]byte("ALL_SHIPS_PLACED\n"))

		if player.opponent != nil && player.opponent.placedShips {
			startBattle(player, player.opponent)
		} else {
			player.conn.Write([]byte("WAITING_FOR_OPPONENT\n"))
		}
	}
}

func startBattle(player1, player2 *Player) {
	player1.conn.Write([]byte("BATTLE_START\n"))
	player2.conn.Write([]byte("BATTLE_START\n"))

	// TODO: game logic or main battleship game engine will go here
}

func generateRoomID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	// rand.seed is deprecated but I found this online so using this at the moment because this depracation doestn matter much. It wokrs.
	rand.Seed(time.Now().UnixNano())

	id := make([]byte, 6)
	for i := range id {
		id[i] = letters[rand.Intn(len(letters))]
	}
	return string(id)
}

// remove player logic. I added this later on becuase player would disconnect and the other client wouldnt know so added this to handle that.
func removePlayer(player *Player) {
	mu.Lock()
	defer mu.Unlock()

	// Remove from waiting players if they were in queue
	for i, p := range waitingPlayers {
		if p == player {
			waitingPlayers = append(waitingPlayers[:i], waitingPlayers[i+1:]...)
			break
		}
	}

	// Remove from private rooms if in any
	for id, room := range privateRooms {
		for _, p := range room.players {
			if p == player {
				delete(privateRooms, id)
				break
			}
		}
	}
}

// NewBoard initializes an empty board
// This is essentially a constructor for the board struct
func NewBoard() *Board {
	board := &Board{
		Grid: [10][10]string{},
		// maybe make this ships struct dynamic of sorts so that people can have custom ships and game with just lets say 2 ship of same class
		Ships: map[string]int{
			"Carrier":    5,
			"Battleship": 4,
			"Cruiser":    3,
			"Submarine":  3,
			"Destroyer":  2,
		},
	}
	return board
}

// place ship on the board
func (b *Board) PlaceShip(name string, start string, direction string) error {
	row, col, err := parseCoordinates(start)
	if err != nil {
		return err
	}

	length, exists := b.Ships[name]
	if !exists {
		return fmt.Errorf("unknown ship")
	}

	// check bounds
	if direction == "H" && col+length > 10 {
		return fmt.Errorf("ship out of bounds horizontally")
	}
	if direction == "V" && row+length > 10 {
		return fmt.Errorf("ship out of bounds vertically")
	}

	// check if place is free
	for i := 0; i < length; i++ {
		r, c := row, col
		if direction == "H" {
			c += i
		} else {
			r += i
		}
		if b.Grid[r][c] != "" {
			return fmt.Errorf("space already occupied")
		}
	}

	// place the ship
	for i := 0; i < length; i++ {
		r, c := row, col
		if direction == "H" {
			c += i
		} else {
			r += i
		}
		b.Grid[r][c] = "S"
	}

	// remove ship from `Ships` map after placing
	delete(b.Ships, name)

	return nil
}

// helper function to parse something like "A5" into (row, col)
func parseCoordinates(input string) (int, int, error) {
	if len(input) < 2 {
		return 0, 0, fmt.Errorf("invalid coordinate")
	}
	rowLetter := strings.ToUpper(string(input[0]))
	colStr := input[1:]

	row := int(rowLetter[0] - 'A')
	col, err := strconv.Atoi(colStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid column number")
	}
	col--

	if row < 0 || row >= 10 || col < 0 || col >= 10 {
		return 0, 0, fmt.Errorf("coordinates out of bounds")
	}

	return row, col, nil
}

// serialize turns the board into a printable string but this method is not working properly and I have no idea why.
// probably the client isn't able to parse this properly.
func (b *Board) Serialize() string {
	var builder strings.Builder
	for _, row := range b.Grid {
		for _, cell := range row {
			// cell name is inprired from dragon ball because cell always had that crazy fit on
			if cell == "" {
				builder.WriteString("W")
			} else {
				builder.WriteString(cell)
			}
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

// returns true if all ships are placed
func (b *Board) AllShipsPlaced() bool {
	return len(b.Ships) == 0
}

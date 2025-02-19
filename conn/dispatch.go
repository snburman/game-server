package conn

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/snburman/game-server/db"
)

const (
	Authenticate        FunctionName = "authenticate"
	LoadOnlinePlayers   FunctionName = "load_online_players"
	LoadNewOnlinePlayer FunctionName = "load_new_online_player"
	RemoveOnlinePlayer  FunctionName = "remove_online_player"
	UpdatePlayer        FunctionName = "update_player"
	Chat                FunctionName = "chat"
)

const (
	Up Direction = iota
	Down
	Left
	Right
)

type (
	Direction int
	Position  struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	}
	FunctionName    string
	Dispatch[T any] struct {
		ID       string       `json:"id"`
		conn     *Conn        `json:"-"`
		Function FunctionName `json:"function"`
		Data     T            `json:"data"`
	}
	PlayerUpdate Player
	ChatMessage  struct {
		UserID   string `json:"user_id"`
		UserName string `json:"username"`
		Message  string `json:"message"`
	}
	Authentication map[string][]string
)

func NewDispatch[T any](id string, conn *Conn, function FunctionName, data T) Dispatch[T] {
	return Dispatch[T]{
		ID:       id,
		conn:     conn,
		Function: function,
		Data:     data,
	}
}

func (d Dispatch[T]) Marshal() Dispatch[[]byte] {
	databytes, err := json.Marshal(d.Data)
	if err != nil {
		log.Println("dispatch data not json encodable", "error", err)
		return Dispatch[[]byte]{}
	}
	return Dispatch[[]byte]{
		ID:       d.ID,
		conn:     d.conn,
		Function: d.Function,
		Data:     databytes,
	}
}

func (d Dispatch[T]) Publish() {
	if d.conn == nil {
		log.Println("nil connection, message not sent")
		return
	}
	dispatchBytes, err := json.Marshal(d)
	if err != nil {
		log.Println("dispatch struct not json encodable", "error", err)
		return
	}
	d.conn.Publish(dispatchBytes)
}

func ParseDispatch[T any](d Dispatch[[]byte]) Dispatch[T] {
	var dis Dispatch[T]
	err := json.Unmarshal(d.Data, &dis.Data)
	if err != nil {
		log.Println("error unmarshalling dispatch data", "error", err)
	}
	dis.ID = d.ID
	dis.conn = d.conn
	dis.Function = d.Function
	return dis
}

func RouteDispatch(d Dispatch[[]byte]) {
	if d.conn == nil {
		panic("nil connection, dispatch not sent")
	}

	if !d.conn.authenticated {
		log.Println("unauthenticated connection")
		d.conn.Close()
		return
	}

	switch d.Function {
	case UpdatePlayer:
		dispatch := ParseDispatch[PlayerUpdate](d)
		player := Player(dispatch.Data)

		// if switching maps
		if d.conn.MapID != "" && d.conn.MapID != player.MapID {
			// remove player from old map
			// create new dispatch
			removalDispatch := NewDispatch(uuid.NewString(), d.conn, RemoveOnlinePlayer, player.UserID)
			// marshal data and call RemoveOnlinePlayer dispatch
			RouteDispatch(removalDispatch.Marshal())

			// update conn with new map id
			d.conn.MapID = player.MapID

			// load player in new map
			// create new dispatch
			loadDispatch := NewDispatch(uuid.NewString(), d.conn, LoadNewOnlinePlayer, player)
			// marshal data and call LoadNewOnlinePlayer dispatch
			RouteDispatch(loadDispatch.Marshal())
		} else if d.conn.MapID != "" && d.conn.MapID == player.MapID {
			// same map
			// update player in player pool
			playerPool.Set(player)
			// get all players in map and update
			for _, p := range playerPool.GetAllByMapID(player.MapID) {
				// player sending the update should not receive the update
				if p.UserID == player.UserID {
					continue
				}
				// get p conn
				conn, ok := wasmConnPool.Get(p.UserID)
				if !ok {
					continue
				}
				// create new dispatch
				newDispatch := NewDispatch(uuid.NewString(), conn, UpdatePlayer, PlayerUpdate(player))
				// update conn with new player data
				newDispatch.Marshal().Publish()
			}
		} else {
			// player is new
			// create new dispatch
			loadDispatch := NewDispatch(uuid.NewString(), d.conn, LoadNewOnlinePlayer, player)
			// marshal data and call LoadNewOnlinePlayer dispatch
			RouteDispatch(loadDispatch.Marshal())
		}
	case RemoveOnlinePlayer:
		// parse user id from dispatch and delete from player pool
		dispatch := ParseDispatch[string](d)
		oldUserID := dispatch.Data
		playerPool.Delete(oldUserID)

		// update all conns in old map
		for _, player := range playerPool.GetAllByMapID(dispatch.conn.MapID) {
			// get individual conns
			conn, ok := wasmConnPool.Get(player.UserID)
			if !ok {
				continue
			}
			// create new dispatch
			newDispatch := NewDispatch(uuid.NewString(), conn, RemoveOnlinePlayer, oldUserID)
			// update conn with player id to remove
			newDispatch.Marshal().Publish()
		}
	case LoadNewOnlinePlayer:
		// parse player from dispatch
		dispatch := ParseDispatch[Player](d)
		player := Player(dispatch.Data)

		// get new player characters
		newPlayerCharacters, err := db.GetPlayerCharactersByUserIDs(db.MongoDB, []string{player.UserID})
		if err != nil {
			log.Println("error getting player characters: ", err)
			return
		}

		// get all player ids in new map
		ids := []string{}
		for _, p := range playerPool.GetAllByMapID(player.MapID) {
			ids = append(ids, p.UserID)
			// get individual conns
			conn, ok := wasmConnPool.Get(p.UserID)
			if !ok {
				continue
			}
			// create new dispatch
			newDispatch := NewDispatch(uuid.NewString(), conn, LoadNewOnlinePlayer, newPlayerCharacters)
			// update conn with new player data
			newDispatch.Marshal().Publish()
		}

		// add new player to pool
		playerPool.Set(player)
		// update conn with new map ID
		d.conn.MapID = player.MapID

		// get all player characters in new map
		allCharacters := []db.PlayerAsset[db.PixelData]{}
		if len(ids) > 0 {
			allCharacters, err = db.GetPlayerCharactersByUserIDs(db.MongoDB, ids)
			if err != nil {
				log.Println("error getting player characters: ", err)
				return
			}
		}

		// update player positions
		for key, p := range allCharacters {
			// get individual conns
			conn, ok := wasmConnPool.Get(p.UserID)
			if !ok {
				continue
			}
			// get player from player pool
			player, ok := playerPool.Get(conn.MapID, p.UserID)
			if !ok {
				continue
			}
			// update player position
			allCharacters[key].X = player.Pos.X
			allCharacters[key].Y = player.Pos.Y
		}
		// create new dispatch
		characterDispatch := NewDispatch(uuid.NewString(), d.conn, LoadOnlinePlayers, allCharacters)
		// update conn with all player characters in new map
		characterDispatch.Marshal().Publish()

	case Chat:
		// parse chat message from dispatch
		dispatch := ParseDispatch[ChatMessage](d)
		chatMessage := dispatch.Data.Message

		// limit message length
		if len(chatMessage) > 50 {
			chatMessage = chatMessage[:50] + "..."
		}

		// get all players in same map as sender
		players, ok := playerPool.GetPlayersInMapByUserID(dispatch.Data.UserID)
		if !ok {
			return
		}
		// send chat message to all players in map
		for _, player := range players {
			// get individual chat conns to display in chat box
			conn, ok := chatConnPool.Get(player.UserID)
			if !ok {
				continue
			}
			// create new dispatch with chat conn
			newDispatch := NewDispatch(uuid.NewString(), conn, Chat, ChatMessage{
				UserID:   dispatch.Data.UserID,
				UserName: dispatch.Data.UserName,
				Message:  chatMessage,
			})
			// update conn with chat message
			newDispatch.Marshal().Publish()

			// get individual wasm conns to display above player
			wasmConn, ok := wasmConnPool.Get(player.UserID)
			if !ok {
				continue
			}
			// update dispatch with wasm conn
			newDispatch.conn = wasmConn
			// update conn with chat message
			newDispatch.Marshal().Publish()
		}
	}
}

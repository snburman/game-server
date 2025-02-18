package conn

import "sync"

var playerPool = NewPlayerPool()

type Player struct {
	UserID string    `json:"user_id"`
	MapID  string    `json:"map_id"`
	Dir    Direction `json:"dir"`
	Pos    Position  `json:"pos"`
}

type PlayerPool struct {
	mu sync.Mutex
	// mapID -> userID -> Player
	pool map[string]map[string]Player
}

func NewPlayerPool() *PlayerPool {
	return &PlayerPool{
		pool: make(map[string]map[string]Player),
	}
}

func (p *PlayerPool) GetAllByMapID(mapID string) map[string]Player {
	p.mu.Lock()
	defer p.mu.Unlock()
	players, ok := p.pool[mapID]
	if !ok {
		return make(map[string]Player)
	}
	return players
}

func (p *PlayerPool) Get(mapID string, userID string) (Player, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	players, ok := p.pool[mapID]
	if !ok {
		return Player{}, false
	}
	player, ok := players[userID]
	return player, ok
}

// GetPlayersInMapByUserID returns all players in a map by one present userID
func (p *PlayerPool) GetPlayersInMapByUserID(userID string) (map[string]Player, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// get player by userID
	for _, players := range p.pool {
		if _, ok := players[userID]; ok {
			return players, true
		}
	}
	return nil, false
}

func (p *PlayerPool) Set(player Player) {
	p.mu.Lock()
	defer p.mu.Unlock()
	players, ok := p.pool[player.MapID]
	if !ok {
		players = make(map[string]Player)
	}
	players[player.UserID] = player
	p.pool[player.MapID] = players
}

func (p *PlayerPool) Delete(userID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for mapID, players := range p.pool {
		if _, ok := players[userID]; ok {
			delete(players, userID)
			p.pool[mapID] = players
		}
	}
}

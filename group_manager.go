package grouplay

import (
	"fmt"
)

const (
	MaxGroupNum = 10
)

var groups map[string]*GameGroup

func init() {
	groups = make(map[string]*GameGroup)
}

type GameGroup struct {
	ID             string        `json:"host"`
	Host           *GamePlayer   `json:"host"`
	MaxPlayer      int           `json:"limit"`
	Players        []*GamePlayer `json:"players"`
	Spectators     []*GamePlayer `json:"spectators"`
	AllowSpectator bool          `json:"allowSpectator"`
	Playing        bool          `json:"playing"`
	Game           *Game         `json:"game"`
}

type SimpleGroup struct {
	ID             string          `json:"id"`
	Host           *SimplePlayer   `json:"host"`
	MaxPlayer      int             `json:"limit"`
	Players        []*SimplePlayer `json:"players"`
	Spectators     []*SimplePlayer `json:"spectators"`
	AllowSpectator bool            `json:"allowSpectator"`
	Playing        bool            `json:"playing"`
	Game           *Game           `json:"game"`
}

type SimplePlayer struct {
	Name string `json:"name"`
	ID   string `json:""`
}

type MyInfo struct {
	Session string `json:"session"`
	InGame  bool   `json:"ingame"`
	Index   int    `json:"index"`
}

// Join : Let a player join into a group if the group doesn't reach the max size.
func (g *GameGroup) Join(p *GamePlayer) error {
	if g.Playing {
		return NewError("The game has been already started, try spectating!")
	}

	if length := len(g.Players); length == g.MaxPlayer {
		return NewError("Player number has reached the max size.")
	}
	if g.Exist(p) {
		return NewError("Already added into the group.")
	}
	g.Players = append(g.Players, p)
	return nil
}

// Spectate : Let a player spectate a group game if it is playing.
func (g *GameGroup) Spectate(p *GamePlayer) error {
	fmt.Println("GameGroup1==>", g)
	if g.Playing {
		if length := len(g.Players); length == g.MaxPlayer {
			g.Spectators = append(g.Spectators, p)
		} else {
			return NewError("The player have left!")
		}
	} else {
		return NewError("The game not start, try to join it!")
	}
	fmt.Println("GameGroup2==>", g)
	return nil
}

func (g *GameGroup) Exit(p *GamePlayer) error {
	if g.Playing {
		return NewError("The game has been already started, can't exit!")
	}
	for i, player := range g.Players {
		if player == p {
			// Remove relation
			player.GroupJoined = nil
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			// If is host of a group
			if player.GroupHosted != nil {
				player.GroupHosted.Host = nil
				player.GroupHosted = nil
				delete(groups, g.ID)
			}
			// Set new mapping
			if len(g.Players) > 0 {
				if g.Host == nil {
					newHost := g.Players[0]
					g.Host = newHost
					g.ID = newHost.ID
					newHost.GroupHosted = g
					groups[g.ID] = g
				}
			}
			return nil
		}
	}
	return NewError("You are not in the specified group.")
}

func (g *GameGroup) Exist(p *GamePlayer) bool {
	for _, player := range g.Players {
		if player == p {
			return true
		}
	}
	return false
}

func FindGroup(id string) (group *GameGroup, ok bool) {
	fmt.Println("Try to find a group with id", id)
	if id == "" {
		return group, false
	}
	group, ok = groups[id]
	return group, ok
}

// CreateGroup : Create a group and name it with the player's name
func CreateGroup(game *Game, player *GamePlayer, max int, allowSpectate bool) (group *GameGroup) {
	group = &GameGroup{
		ID:             player.ID,
		Host:           player,
		MaxPlayer:      max,
		Players:        make([]*GamePlayer, 0),
		Spectators:     make([]*GamePlayer, 0),
		AllowSpectator: allowSpectate,
		Playing:        false,
		Game:           game,
	}
	// Add to group
	groups[group.ID] = group
	return group
}

func generateSimplePlayer(p *GamePlayer) *SimplePlayer {
	simple := SimplePlayer{
		ID:   p.ID,
		Name: p.Name,
	}
	return &simple
}

func generateSimplePlayerArray(list []*GamePlayer) []*SimplePlayer {
	simple := make([]*SimplePlayer, len(list))
	for i, p := range list {
		simple[i] = generateSimplePlayer(p)
	}
	return simple
}

func generateSimpleGroup(g *GameGroup) *SimpleGroup {
	simple := SimpleGroup{
		ID:             g.ID,
		Host:           generateSimplePlayer(g.Host),
		MaxPlayer:      g.MaxPlayer,
		Players:        generateSimplePlayerArray(g.Players),
		Spectators:     generateSimplePlayerArray(g.Spectators),
		AllowSpectator: g.AllowSpectator,
		Playing:        g.Playing,
		Game:           g.Game,
	}
	return &simple
}

func BuildGroupList() GroupListMessage {
	waiting := make([]*SimpleGroup, 0)
	playing := make([]*SimpleGroup, 0)
	for _, group := range groups {
		simpleGroup := generateSimpleGroup(group)
		if group.Playing {
			fmt.Println("Add playing group")
			playing = append(playing, simpleGroup)
		} else {
			fmt.Println("Add waiting group")
			waiting = append(waiting, simpleGroup)
		}
	}
	return GroupListMessage{nil, nil, waiting, playing}
}

func NotifyGroupListToSpectator(p *GamePlayer) {
	groupList := BuildGroupList()
	if p.GroupSpectating != nil {
		groupList.Joined = generateSimpleGroup(p.GroupSpectating)
	} else {
		groupList.Joined = nil
	}
	// Create my info
	groupList.Info = &MyInfo{p.ID, p.InGame, p.Index}
	SendStructMessage(*p.Session, CmdGroupUpdate, groupList, true)
}

func NotifyGroupListToOne(p *GamePlayer) {
	groupList := BuildGroupList()
	if p.GroupJoined != nil {
		groupList.Joined = generateSimpleGroup(p.GroupJoined)
	} else {
		groupList.Joined = nil
	}
	// Create my info
	groupList.Info = &MyInfo{p.ID, p.InGame, 0}
	SendStructMessage(*p.Session, CmdGroupUpdate, groupList, true)
}
func NotifyGroupListToAll() {
	groupList := BuildGroupList()
	for _, p := range players {
		fmt.Println("player-->", p)
		if p.GroupJoined != nil {
			groupList.Joined = generateSimpleGroup(p.GroupJoined)
		} else {
			groupList.Joined = nil
		}
		// Create my info
		groupList.Info = &MyInfo{p.ID, p.InGame, 0}
		SendStructMessage(*p.Session, CmdGroupUpdate, groupList, true)
	}
}

func (g *GameGroup) NotifyPlayerExcept(cmd string, msg interface{}, player *GamePlayer) {
	for _, p := range g.Players {
		if p != player {
			SendStructMessage(*p.Session, cmd, msg, true)
		}
	}
}

func (g *GameGroup) NotifyPlayer(cmd string, msg interface{}) {
	g.NotifyPlayerExcept(cmd, msg, nil)
}

func (g *GameGroup) NotifySpectatorExcept(cmd string, msg interface{}, player *GamePlayer) {
	for _, p := range g.Spectators {
		if p != player {
			SendStructMessage(*p.Session, cmd, msg, true)
		}
	}
}

func (g *GameGroup) NotifySpectator(cmd string, msg interface{}) {
	g.NotifySpectatorExcept(cmd, msg, nil)
}

func (g *GameGroup) NotifyAll(cmd string, msg interface{}) {
	g.NotifyPlayerExcept(cmd, msg, nil)
	g.NotifySpectatorExcept(cmd, msg, nil)
}
func (g *GameGroup) NotifyAllExcept(cmd string, msg interface{}, player *GamePlayer) {
	g.NotifyPlayerExcept(cmd, msg, player)
	g.NotifySpectatorExcept(cmd, msg, player)
}

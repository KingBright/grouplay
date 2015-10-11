package grouplay

import (
	"container/list"
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
	ID             string
	Host           *GamePlayer
	MaxPlayer      int
	Players        *list.List
	Spectators     *list.List
	AllowSpectator bool
	Playing        bool
}

type GroupInfo struct {
	Host           string `json:"host"`
	ID             string `json:"id"`
	Limit          int    `json:"limit"`
	Players        int    `json:"current"`
	Spectators     int    `json:"spectators"`
	AllowSpectator bool   `json:"allowSpectator"`
	Playing        bool   `json:"playing"`
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
	if length := g.Players.Len(); length == g.MaxPlayer {
		return NewError("Player number has reached the max size.")
	}
	if g.Exist(p) {
		return NewError("Already added into the group.")
	}
	g.Players.PushBack(p)
	return nil
}

func (g *GameGroup) Exit(p *GamePlayer) error {
	if g.Playing {
		return NewError("The game has been already started, can't exit!")
	}
	for e := g.Players.Front(); e != nil; e = e.Next() {
		player := e.Value.(*GamePlayer)
		if player == p {
			// Remove relation
			player.GroupJoined = nil
			g.Players.Remove(e)
			// If is host of a group
			if player.GroupHosted != nil {
				player.GroupHosted.Host = nil
				player.GroupHosted = nil
				delete(groups, g.ID)
			}
			// Set new mapping
			if g.Players.Len() > 0 {
				if g.Host == nil {
					newHost := g.Players.Front().Value.(*GamePlayer)
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
	for e := g.Players.Front(); e != nil; e = e.Next() {
		player := e.Value.(*GamePlayer)
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
func CreateGroup(player *GamePlayer, max int, allowSpectate bool) (group *GameGroup) {
	group = &GameGroup{
		ID:             player.ID,
		Host:           player,
		MaxPlayer:      max,
		Players:        list.New(),
		Spectators:     list.New(),
		AllowSpectator: allowSpectate,
		Playing:        false,
	}
	// Add to group
	groups[group.ID] = group
	return group
}

func BuildGroupList() GroupListMessage {
	waiting := make([]GroupInfo, 0)
	playing := make([]GroupInfo, 0)
	for _, group := range groups {
		info := GroupInfo{
			Host:           group.Host.Name,
			ID:             group.ID,
			Limit:          group.MaxPlayer,
			Players:        group.Players.Len(),
			Spectators:     group.Spectators.Len(),
			AllowSpectator: group.AllowSpectator,
			Playing:        group.Playing,
		}
		if group.Playing {
			fmt.Println("Add playing group")
			playing = append(playing, info)
		} else {
			fmt.Println("Add waiting group")
			waiting = append(waiting, info)
		}
	}
	return GroupListMessage{nil, nil, waiting, playing}
}

func NotifyGroupListToAll() {
	groupList := BuildGroupList()
	var index = 0
	for _, p := range players {
		index++
		if p.GroupJoined != nil {
			groupList.Joined = &GroupInfo{
				Host:           p.GroupJoined.Host.Name,
				ID:             p.GroupJoined.ID,
				Limit:          p.GroupJoined.MaxPlayer,
				Players:        p.GroupJoined.Players.Len(),
				Spectators:     p.GroupJoined.Spectators.Len(),
				AllowSpectator: p.GroupJoined.AllowSpectator,
				Playing:        p.GroupJoined.Playing,
			}
		} else {
			groupList.Joined = nil
		}
		// Create my info
		groupList.Info = &MyInfo{p.ID, p.InGame, index}
		SendStructMessage(*p.Session, CmdGroupUpdate, groupList, true)
	}
}

func (g *GameGroup) NotifyPlayerString(cmd, msg string) {
	for e := g.Players.Front(); e != nil; e = e.Next() {
		p := e.Value.(*GamePlayer)
		Send(*p.Session, cmd, msg, true)
	}
}

func (g *GameGroup) NotifySpectatorString(cmd, msg string) {
	for e := g.Spectators.Front(); e != nil; e = e.Next() {
		p := e.Value.(*GamePlayer)
		Send(*p.Session, cmd, msg, true)
	}
}
func (g *GameGroup) NotifyPlayerStruct(cmd string, msg interface{}) {
	for e := g.Players.Front(); e != nil; e = e.Next() {
		p := e.Value.(*GamePlayer)
		SendStructMessage(*p.Session, cmd, msg, true)
	}
}

func (g *GameGroup) NotifySpectatorStruct(cmd string, msg interface{}) {
	for e := g.Spectators.Front(); e != nil; e = e.Next() {
		p := e.Value.(*GamePlayer)
		SendStructMessage(*p.Session, cmd, msg, true)
	}
}

func (g *GameGroup) NotifyAllStruct(cmd string, msg interface{}) {
	g.NotifyPlayerStruct(cmd, msg)
	g.NotifySpectatorStruct(cmd, msg)
}
func (g *GameGroup) NotifyAllString(cmd, msg string) {
	g.NotifyPlayerString(cmd, msg)
	g.NotifySpectatorString(cmd, msg)
}

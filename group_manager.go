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
}

// Join : Let a player join into a group if the group doesn't reach the max size.
func (g *GameGroup) Join(p *GamePlayer) error {
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
			}
			// Remove old mapping
			delete(groups, g.ID)
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
	for _, v := range groups {
		info := GroupInfo{
			Host:           v.Host.Name,
			ID:             v.ID,
			Limit:          v.MaxPlayer,
			Players:        v.Players.Len(),
			Spectators:     v.Spectators.Len(),
			AllowSpectator: v.AllowSpectator,
		}
		if v.Playing {
			fmt.Println("Add playing group")
			playing = append(playing, info)
		} else {
			fmt.Println("Add waiting group")
			waiting = append(waiting, info)
		}
	}
	return GroupListMessage{nil, waiting, playing}
}

func NotifyGroupList() {
	groupList := BuildGroupList()
	for _, p := range players {
		if p.GroupJoined != nil {
			groupList.Joined = &GroupInfo{
				Host:           p.GroupJoined.Host.Name,
				ID:             p.GroupJoined.ID,
				Limit:          p.GroupJoined.MaxPlayer,
				Players:        p.GroupJoined.Players.Len(),
				Spectators:     p.GroupJoined.Spectators.Len(),
				AllowSpectator: p.GroupJoined.AllowSpectator,
			}
		} else {
			groupList.Joined = nil
		}
		SendStructMessage(*p.Session, CmdGroupUpdate, groupList, true)
	}
}

func (g *GameGroup) NotifyPlayer(msg string) {
	for e := g.Players.Front(); e != nil; e = e.Next() {
		p := e.Value.(*GamePlayer)
		SendJsonMessage(*p.Session, msg)
	}
}

func (g *GameGroup) NotifySpectator(msg string) {
	for e := g.Spectators.Front(); e != nil; e = e.Next() {
		p := e.Value.(*GamePlayer)
		SendJsonMessage(*p.Session, msg)
	}
}

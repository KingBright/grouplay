package grouplay

import (
	"fmt"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

// Id as key
var players map[string]*GamePlayer

func init() {
	players = make(map[string]*GamePlayer)
}

type GamePlayer struct {
	ID              string
	Name            string
	Index           int
	Session         *sockjs.Session
	GroupHosted     *GameGroup
	GroupJoined     *GameGroup
	GroupSpectating *GameGroup
}

func (p *GamePlayer) Update(session sockjs.Session, id string) {
	oldId := p.ID
	fmt.Println("old id is", oldId)
	p.ID = id
	p.Session = &session
	delete(players, oldId)
	fmt.Println("old id removed", oldId)
	players[p.ID] = p
	fmt.Println("new id added", p.ID)
}

func Register(session sockjs.Session, oldId string, name string) {
	if player, ok := FindPlayer(oldId); ok {
		player.Update(session, session.ID())
		fmt.Println("Find an existed player & update it")
	} else {
		id := session.ID()
		players[id] = &GamePlayer{id, name, 0, &session, nil, nil, nil}
		fmt.Println("Register as new")
	}
}

func FindPlayer(id string) (player *GamePlayer, ok bool) {
	fmt.Println("Try to find a player with id", id)
	if id == "" {
		return player, false
	}
	player, ok = players[id]
	return player, ok
}

func (p *GamePlayer) CreateGroup(max int, allowSpectate bool) (bool, error) {
	if p.GroupHosted != nil {
		fmt.Println("You already hosted a group")
		return false, NewError("You already hosted a group.")
	}
	fmt.Println("group hosted", p.GroupHosted)
	if p.GroupJoined != nil {
		fmt.Println("You already joined a group")
		return false, NewError("You already joined a group.")
	}
	fmt.Println("group joined", p.GroupJoined)
	group := CreateGroup(p, max, allowSpectate)
	p.GroupHosted = group
	fmt.Println("A group created by player", p.ID)
	if err := group.Join(p); err == nil {
		p.GroupJoined = group
		return true, nil
	} else {
		return true, err
	}
}

func (p *GamePlayer) JoinGroup(id string) (bool, error) {
	if p.GroupJoined != nil {
		fmt.Println("Already joined a group.")
		return false, NewError("You Already joined a group.")
	}
	if group, ok := FindGroup(id); ok {
		if err := group.Join(p); err == nil {
			p.GroupJoined = group
			return true, nil
		} else {
			return false, err
		}
	}
	fmt.Println("Target group not found.")
	return false, NewError("Target group not found.")
}

func (p *GamePlayer) ExitGroup(id string) (bool, error) {
	if p.GroupJoined != nil {
		if err := p.GroupJoined.Exit(p); err == nil {
			return true, err
		} else {
			return false, err
		}
	}
	return false, NewError("You haven't joined any group")
}

func (p *GamePlayer) StartGame() {

}

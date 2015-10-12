package grouplay

var controllers map[*GameGroup]GameController

func init() {
	controllers = make(map[*GameGroup]GameController)
}

func StartGame(group *GameGroup, groupId string) error {
	if group.ID != groupId {
		return NewError("You are not the host of this group.")
	}
	if group.Playing {
		return NewError("Game is already started.")
	}

	creator := GetControllerCreator(group.Game.Name)

	if creator != nil {
		if group.Players != nil && group.Players.Len() >= 2 {
			for e := group.Players.Front(); e != nil; e = e.Next() {
				p := e.Value.(*GamePlayer)
				if p.InGame {
					return NewError("Someone is still in game." + p.Name)
				}
			}
			for e := group.Players.Front(); e != nil; e = e.Next() {
				p := e.Value.(*GamePlayer)
				p.InGame = true
			}
			group.Playing = true
		} else {
			return NewError("Players not enough!")
		}
		createIndex(group)

		controller := creator()
		controllers[group] = controller
		controller.InitData(group)
		return nil
	}
	return NewError("Controller creator is not registered.")
}

func createIndex(g *GameGroup) {
	index := 0
	for e := g.Players.Front(); e != nil; e = e.Next() {
		p := e.Value.(*GamePlayer)
		p.Index = index
		index++
	}
}

func ExitGame(player *GamePlayer) error {
	if player.InGame {
		player.InGame = false
		NotifyGroupListToAll()
		return nil
	} else {
		return NewError("Player is not in game.")
	}
}

func UpdateData(player *GamePlayer, group *GameGroup, action, data string) error {
	if group.Playing {
		if controller, ok := controllers[group]; ok {
			if err := controller.UpdateData(player.Index, action, data); err == nil {
				notifyDataUpdate(group)
				if controller.IsFinished() {
					group.Playing = false
					group.NotifyAll(CmdGameFinish, "")
				}
				return nil
			} else {
				return err
			}
		}
	}
	return NewError("The game not started yet.")
}

func GetDataForPlayer(player *GamePlayer) error {
	if player.GroupJoined == nil {
		return NewError("You are not in a group.")
	}
	if !player.GroupJoined.Playing {
		return NewError("Game is not started!")
	}
	controller := controllers[player.GroupJoined]
	if controller == nil {
		return NewError("Controller is empty!")
	}
	notifyPlayer(player, CmdUpdateData, controller.GetData(player, player.GroupJoined))
	return nil
}

func CheckPlayingGame(oldId, newId string) error {
	if player, ok := players[newId]; ok {
		if player.GroupJoined == nil {
			return NewError("You are not in a group.")
		}
		if !player.GroupJoined.Playing {
			// If refresh by user, set the InGame status to false
			if player.InGame {
				player.InGame = false
			}
			return NewError("Game is not started!")
		}
		controller := controllers[player.GroupJoined]
		if controller == nil {
			return NewError("Controller is empty!")
		}
		controller.OnSessionUpdate(oldId, newId)
		notifyPlayer(player, CmdPlaying, "")
		return nil
	} else {
		return NewError("Player is not exist.")
	}
}

func notifyPlayer(p *GamePlayer, cmd, data string) {
	Send(*p.Session, cmd, data, true)
}

func notifyDataUpdate(g *GameGroup) {
	if controller, ok := controllers[g]; ok {
		for e := g.Players.Front(); e != nil; e = e.Next() {
			p := e.Value.(*GamePlayer)
			notifyPlayer(p, CmdUpdateData, controller.GetData(p, g))
		}
		for e := g.Spectators.Front(); e != nil; e = e.Next() {
			p := e.Value.(*GamePlayer)
			notifyPlayer(p, CmdUpdateData, controller.GetData(nil, g))
		}
	}
}

type GameController interface {
	GetData(player *GamePlayer, group *GameGroup) string
	UpdateData(index int, action, data string) error
	InitData(group *GameGroup)
	IsFinished() bool
	OnSessionUpdate(oldSession, newSession string)
}

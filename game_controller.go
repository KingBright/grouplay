package grouplay

var controllers map[*GameGroup]*Controller
var creator ControllerCreator

func init() {
	controllers = make(map[*GameGroup]*Controller)
}

func RegisterControllerCreator(c ControllerCreator) {
	creator = c
}

func StartGame(group *GameGroup, groupId string) error {
	if group.ID != groupId {
		return NewError("You are not the host of this group.")
	}
	if group.Playing {
		return NewError("Game is already started.")
	}
	if creator != nil {
		controller := creator.CreateController()
		group.Playing = true
		controllers[group] = &controller
		controller.InitData()

		notifyDataUpdate(group)
		return nil
	}
	return NewError("Controller creator is not registered.")
}

func UpdateData(group *GameGroup, action, data string) error {
	if group.Playing {
		if controller, ok := controllers[group]; ok {
			ctrl := *controller
			if err := ctrl.UpdateData(action, data); err == nil {
				notifyDataUpdate(group)
				return nil
			} else {
				return err
			}
		}
	}
	return NewError("The game not started yet.")
}

func CheckPlayingGame(id string) {
	if player, ok := players[id]; ok {
		if player.GroupJoined != nil && player.GroupJoined.Playing {
			controller := *controllers[player.GroupJoined]
			notifyGamePlaying(player, controller.GetData())
		}
	}
}

func notifyGamePlaying(p *GamePlayer, data string) {
	Send(*p.Session, CmdUpdateData, data, true)
}

func notifyDataUpdate(group *GameGroup) {
	controller := *controllers[group]
	group.NotifyPlayer(CmdUpdateData, controller.GetData())
	group.NotifySpectator(CmdUpdateData, controller.GetData())
}

type Controller interface {
	GetData() string
	UpdateData(action, data string) error
	InitData()
}

type ControllerCreator interface {
	CreateController() Controller
}

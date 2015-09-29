package grouplay

var controllers map[*GameGroup]*Controller
var creator ControllerCreator

func init() {
	controllers = make(map[*GameGroup]*Controller)
}

func RegisterControllerCreator(c ControllerCreator) {
	creator = c
}

func StartGame(player *GamePlayer, controller *Controller) (string, error) {
	if player.GroupJoined != nil {
		if player.GroupJoined.Playing {
			return "", NewError("Game is already started.")
		}
		player.GroupJoined.Playing = true
		controllers[player.GroupJoined] = controller
		return (*controller).GetInitialData(), nil
	}
	return "", NewError("You are not in any group.")
}

func Update(player *GamePlayer, data string) error {
	if player.GroupJoined != nil {
		if player.GroupJoined.Playing {
			if controller, ok := controllers[player.GroupJoined]; ok {
				ctrl := *controller
				if ctrl.IsFinished() {
					return NewError("Game is already finished.")
				}
				if err := ctrl.UpdateData(data); err == nil {
					// If game is finished.
					if ctrl.IsFinished() {
						player.GroupJoined.NotifyPlayer(ctrl.GetFinishMsg())
						player.GroupJoined.NotifySpectator(ctrl.GetFinishMsg())
					} else {
						player.GroupJoined.NotifyPlayer(ctrl.GetData())
						player.GroupJoined.NotifySpectator(ctrl.GetData())
					}
					return nil
				} else {
					return err
				}
			}
			return NewError("The game not started yet.")
		}
		return NewError("The game not started yet.")
	}
	return NewError("You are not in a game group yet.")
}

type Controller interface {
	GetData() string
	UpdateData(data string) error
	IsFinished() bool
	GetFinishMsg() string
	GetInitialData() string
}

type ControllerCreator interface {
	createController() *Controller
}

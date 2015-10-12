package grouplay

var gameList []Game
var gameCreator map[string]func() GameController

func init() {
	gameList = make([]Game, 0)
	gameCreator = make(map[string]func() GameController)
}

type Game struct {
	Name          string `json:"name"`          //game name
	Url           string `json:"url"`           //game page
	Rule          string `json:"rule"`          //rule page
	SupportPlayer []int  `json:"supportPlayer"` // support player number
}

// More game private settings could be added

func RegisterGame(game Game, creator func() GameController) {
	gameList = append(gameList, game)
	gameCreator[game.Name] = creator
}

func GetGameList() []Game {
	return gameList
}

func GetGame(name string) *Game {
	for _, g := range gameList {
		if g.Name == name {
			return &g
		}
	}
	return nil
}

func GetControllerCreator(name string) func() GameController {
	return gameCreator[name]
}

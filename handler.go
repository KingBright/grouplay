package grouplay

import (
	"encoding/json"
	"fmt"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"net/http"
	"strings"
)

func NewHandler(prefix string) http.Handler {
	return sockjs.NewHandler(prefix, sockjs.DefaultOptions, quoridorHandler)
}

func quoridorHandler(session sockjs.Session) {
	for {
		if msg, err := session.Recv(); err == nil {
			handleMsg(session, msg)
			continue
		}
		break
	}
}

func handleMsg(session sockjs.Session, msg string) {
	fmt.Println("Message :", msg)
	decoder := json.NewDecoder(strings.NewReader(msg))
	message := new(Message)
	decoder.Decode(message)

	if message.Confirm {
		fmt.Println("Received confirm message : " + msg + " , from " + session.ID())
		return
	}

	switch message.Cmd {
	case CmdCreateGroup:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			createInfo := new(CreateGroupMesssage)
			decoder.Decode(createInfo)

			if ok, err := player.CreateGroup(createInfo.Max, createInfo.AllowSpectator); ok {
				SendStructMessage(session, message.Cmd, struct {
					ID     string `json:"groupId"`
					Hoster bool   `json:"isHoster"`
					OK     bool   `json:"ok"`
				}{ID: player.GroupHosted.ID, Hoster: true, OK: true}, true)
				NotifyGroupListToAll()
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdJoinGroup:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			joinInfo := new(JoinGroupMessage)
			decoder.Decode(joinInfo)

			if ok, err := player.JoinGroup(joinInfo.GroupId); ok {
				SendStructMessage(session, message.Cmd, struct {
					ID string `json:"groupId"`
					OK bool   `json:"ok"`
				}{ID: player.GroupJoined.ID, OK: true}, true)
				NotifyGroupListToAll()
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdExitGroup:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			exitInfo := new(ExitGroupMessage)
			decoder.Decode(exitInfo)

			if ok, err := player.ExitGroup(exitInfo.GroupId); ok {
				SendStructMessage(session, message.Cmd, struct {
					OK bool `json:"ok"`
				}{OK: true}, true)
				NotifyGroupListToAll()
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdRegister:
		decoder := json.NewDecoder(strings.NewReader(message.Msg))
		registerInfo := new(RegisterMessage)
		decoder.Decode(registerInfo)
		if err := Register(session, registerInfo.ID, registerInfo.Name); err == nil {
			SendStructMessage(session, message.Cmd, struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				OK   bool   `json:"ok"`
			}{session.ID(), registerInfo.Name, true}, true)
			NotifyGroupListToAll()
		} else {
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
		CheckPlayingGame(registerInfo.ID, session.ID())
	case CmdStartGame:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			startGameMessage := new(StartGameMessage)
			decoder.Decode(startGameMessage)

			group := player.GroupHosted
			if group == nil {
				err := NewError("You haven't hosted a group.")
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
			if err := StartGame(group, startGameMessage.GroupId); err == nil {
				//Notify all player to action
				group.NotifyPlayer(message.Cmd, struct {
					OK bool `json:"ok"`
				}{true})
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdGetData:
		if player, ok := FindPlayer(session.ID()); ok {
			if err := GetDataForPlayer(player); err != nil {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdExitGame:
		if player, ok := FindPlayer(session.ID()); ok {
			group := player.GroupJoined
			if group == nil {
				err := NewError("You haven't jonied a group.")
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			} else if group.Playing {
				err := NewError("The game is playing, you can't leave.")
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			} else {
				if err := ExitGame(player); err == nil {
					//Notify first player to action
					SendStructMessage(session, message.Cmd, struct {
						OK bool `json:"ok"`
					}{true}, true)
				} else {
					SendErrorMessage(session, message.Cmd, err.Error(), false, true)
				}
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdPlayerAction:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			dataUpdateMessage := new(DataUpdateMessage)
			decoder.Decode(dataUpdateMessage)

			if player.GroupJoined != nil {
				if err := UpdateData(player, player.GroupJoined, dataUpdateMessage.Action, dataUpdateMessage.Data); err == nil {
					SendStructMessage(session, message.Cmd, struct {
						OK bool `json:"ok"`
					}{true}, true)
				} else {
					SendErrorMessage(session, message.Cmd, err.Error(), false, true)
				}
			} else {
				err := NewError("You are not in a group")
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdQuitGame:
		if player, ok := FindPlayer(session.ID()); ok {
			if player.GroupHosted != nil {
				if player.GroupHosted.Playing {
					player.GroupHosted.Playing = false
					player.InGame = false
					player.GroupHosted.NotifyAllExcept(CmdHostStop, "", player)
				}
			} else {
				player.InGame = false
			}
			NotifyGroupListToOne(player)
		}
	}
}

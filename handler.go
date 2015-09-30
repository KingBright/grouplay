package grouplay

import (
	"encoding/json"
	"fmt"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"log"
	"net/http"
	"strings"
)

func StartServe() {
	handler := sockjs.NewHandler("/quoridor", sockjs.DefaultOptions, quoridorHandler)
	log.Fatal(http.ListenAndServe(":8081", handler))
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

func RegisterController() {

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
				NotifyGroupList()
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
				NotifyGroupList()
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
				NotifyGroupList()
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

			NotifyGroupList()
		} else {
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}

		CheckPlayingGame(session.ID())
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

				//Notify first player to action
				SendStructMessage(session, message.Cmd, struct {
					OK bool `json:"ok"`
				}{true}, true)
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}
	case CmdUpdateData:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			dataUpdateMessage := new(DataUpdateMessage)
			decoder.Decode(dataUpdateMessage)

			if player.GroupJoined != nil {
				if err := UpdateData(player.GroupJoined, dataUpdateMessage.Action, dataUpdateMessage.Data); err == nil {
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
	}
}

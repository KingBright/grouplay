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
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}

		NotifyGroupList()
	case CmdJoinGroup:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			joinInfo := new(JoinGroupMessage)
			decoder.Decode(joinInfo)

			if ok, err := player.JoinGroup(joinInfo.Id); ok {
				SendStructMessage(session, message.Cmd, struct {
					ID     string `json:"groupId"`
					Hoster bool   `json:isHoster`
					OK     bool   `json:"ok"`
				}{ID: player.GroupHosted.ID, Hoster: false, OK: true}, true)
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}

		NotifyGroupList()
	case CmdExitGroup:
		if player, ok := FindPlayer(session.ID()); ok {
			decoder := json.NewDecoder(strings.NewReader(message.Msg))
			exitInfo := new(ExitGroupMessage)
			decoder.Decode(exitInfo)

			if ok, err := player.ExitGroup(exitInfo.Id); ok {
				SendStructMessage(session, message.Cmd, struct {
					OK bool `json:"ok"`
				}{OK: true}, true)
			} else {
				SendErrorMessage(session, message.Cmd, err.Error(), false, true)
			}
		} else {
			err := NewError("No user found with id " + session.ID())
			SendErrorMessage(session, message.Cmd, err.Error(), false, true)
		}

		NotifyGroupList()
	case CmdRegister:
		decoder := json.NewDecoder(strings.NewReader(message.Msg))
		registerInfo := new(RegisterMessage)
		decoder.Decode(registerInfo)
		Register(session, registerInfo.ID, registerInfo.Name)
		SendStructMessage(session, message.Cmd, struct {
			ID string `json:"id"`
			OK bool   `json:ok`
		}{ID: session.ID(), true}, true)

		NotifyGroupList()
	}
}

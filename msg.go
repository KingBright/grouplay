package grouplay

import (
	"encoding/json"
	"fmt"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

// Communicate protocal.
const (
	CmdRegister     = "register"
	CmdCreateGroup  = "create_group"
	CmdExitGroup    = "exit_group"
	CmdJoinGroup    = "join_group"
	CmdGroupUpdate  = "group_update"
	CmdUpdateData   = "update_data"
	CmdStartGame    = "start_game"
	CmdExitGame     = "exit_game"
	CmdGetData      = "get_data"
	CmdPlaying      = "playing"
	CmdPlayerAction = "player_action"
	CmdQuitGame     = "quit_game"
	CmdGameFinish   = "game_finish"
	CmdHostStop     = "host_stop"
)

// Message format for transaction between server & client.
type Message struct {
	Cmd     string `json:"cmd"`     // Message cmd
	Msg     string `json:"msg"`     // Message body
	Confirm bool   `json:"confirm"` // Indicates if this message is a confirm message
}

type RegisterMessage struct {
	ID   string `json:"id"` //Old session id
	Name string `json:"name"`
}

type CreateGroupMesssage struct {
	Max            int  `json:"max"`
	AllowSpectator bool `json:"allowSpectator"`
}

type JoinGroupMessage struct {
	GroupId string `json:"groupId"`
}

type ExitGroupMessage struct {
	GroupId string `json:"groupId"`
}

type ErrorMessage struct {
	Msg string `json:"msg"`
	OK  bool   `json:"ok"`
}

type GroupListMessage struct {
	Info    *MyInfo     `json:"myInfo"`
	Joined  *GroupInfo  `json:"joined"`
	Waiting []GroupInfo `json:"waiting"`
	Playing []GroupInfo `json:"playing"`
}

type StartGameMessage struct {
	GroupId string `json:"groupId"`
}

type DataUpdateMessage struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

func (m Message) String() string {
	return fmt.Sprintf("%v %v %v", m.Cmd, m.Msg, m.Confirm)
}

// Send message back to client.
func Send(session sockjs.Session, cmd string, msg string, confirm bool) {
	message := new(Message)
	message.Cmd = cmd
	message.Msg = msg
	message.Confirm = confirm
	SendMessage(session, message)
}

func SendMessage(session sockjs.Session, message *Message) {
	bytes, err := json.Marshal(message)
	if err != nil {
		fmt.Println("send message error:", err)
		return
	}
	json := string(bytes)
	fmt.Println("Send message :", json)
	SendJsonMessage(session, json)
}

func SendJsonMessage(session sockjs.Session, json string) {
	session.Send(json)
}

func SendStructMessage(session sockjs.Session, cmd string, msg interface{}, confirm bool) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("send message error:", err)
		return
	}
	body := string(bytes)
	Send(session, cmd, body, confirm)
}

func SendErrorMessage(session sockjs.Session, cmd string, msg string, ok bool, confirm bool) {
	SendStructMessage(session, cmd, ErrorMessage{Msg: msg, OK: ok}, confirm)
}

func ToJson(msg interface{}) string {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return ""
	}
	return string(bytes)
}

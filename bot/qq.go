package bot

import (
	"github.com/mbeoliero/bot/chat"
	"github.com/mbeoliero/bot/utils"
	"github.com/tidwall/gjson"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/sirupsen/logrus"
)

const (
	DevicePath = "./conf/device.json"

	PrivateMsg MsgType = "private"
	GroupMsg   MsgType = "group"
)

type MsgType string

type QQBot struct {
	UID       int64
	Client    *client.QQClient
	GroupList []int64
	UserList  []int64
}

func NewQQBot(uid int64, pass string, groupList, userList []int64) *QQBot {
	return &QQBot{
		Client:    client.NewClient(uid, pass),
		GroupList: groupList,
		UserList:  userList,
	}
}

func (q *QQBot) Start() {
	q.SetDevice()
	q.Login()

	q.Client.GroupMessageEvent.Subscribe(q.ReplyMentionMessage)
	q.Client.PrivateMessageEvent.Subscribe(q.ReplyPrivateMessage)
}

func (q *QQBot) SetDevice() {
	var device *client.DeviceInfo
	if utils.PathExists(DevicePath) {
		device = new(client.DeviceInfo)
		err := device.ReadJson([]byte(utils.ReadFile(DevicePath)))
		if err != nil {
			logrus.Errorf("failed to load device info, err %s", err)
			return
		}
	} else {
		device = client.GenRandomDevice()
		utils.WriteFile(DevicePath, device.ToJson())
	}

	q.Client.UseDevice(device)

	currentName := device.Protocol.Version().SortVersionName
	remoteVersion, err := getRemoteLatestProtocolVersion(int(device.Protocol.Version().Protocol))
	if err != nil {
		logrus.Errorf("failed to get remote protocol version, err %s", err)
		return
	}

	remoteName := gjson.GetBytes(remoteVersion, "sort_version_name").String()
	if remoteName != currentName {
		err = device.Protocol.Version().UpdateFromJson(remoteVersion)
		if err != nil {
			logrus.Warnf("failed to update protocol version, err %s", err)
			return
		}

		logrus.Debugf("updated protocol version to %s", remoteName)
	}
}

func (q *QQBot) Login() {
	var err error
	if q.Client.Uin != 0 {
		err = commonLogin(q.Client)
	} else {
		err = qrcodeLogin(q.Client)
	}

	if err != nil {
		logrus.Errorf("failed to login, err %s", err)
		return
	}

	q.UID = q.Client.Uin
}

func (q *QQBot) needReply(targetID int64, msgType MsgType) bool {
	switch msgType {
	case PrivateMsg:
		for _, id := range q.UserList {
			if id == targetID {
				return true
			}
		}
	case GroupMsg:
		for _, id := range q.GroupList {
			if id == targetID {
				return true
			}
		}
	}

	return false
}

func (q *QQBot) ReplyMentionMessage(qqClient *client.QQClient, event *message.GroupMessage) {
	if !q.needReply(event.GroupCode, GroupMsg) {
		return
	}

	content, target := GetMessageContent(event.Elements)
	if target != q.UID {
		return
	}

	logrus.Debugf("received uid %d message %s", target, content)
	resp, err := chat.GetResponseFromGpt(content)
	if err != nil {
		logrus.Errorf("failed to get response from gpt, err %s", err)
		return
	}

	q.SendMessage(event.GroupCode, GroupMsg, NewReplyMessage(event.Id, event.Time, event.Sender.Uin, resp))
}

func (q *QQBot) ReplyPrivateMessage(qqClient *client.QQClient, event *message.PrivateMessage) {
	if !q.needReply(event.Sender.Uin, GroupMsg) {
		return
	}

	content, target := GetMessageContent(event.Elements)
	logrus.Debugf("received uid %d message %s", target, content)

	resp, err := chat.GetResponseFromGpt(content)
	if err != nil {
		logrus.Errorf("failed to get response from gpt, err %s", err)
		return
	}

	q.SendMessage(event.Sender.Uin, PrivateMsg, NewTextMessage(resp))
}

// SendMessage targetID: userID or groupID
func (q *QQBot) SendMessage(targetID int64, msgType MsgType, msg ...Element) {
	elemList := make([]message.IMessageElement, 0)
	for _, m := range msg {
		elemList = append(elemList, m.ToElem())
	}

	send := &message.SendingMessage{Elements: elemList}

	switch msgType {
	case PrivateMsg:
		q.SendPrivateMessage(targetID, send)
	case GroupMsg:
		q.SendGroupMessage(targetID, send)
	default:
		return
	}
}

func (q *QQBot) SendPrivateMessage(userID int64, msg *message.SendingMessage) {
	q.Client.SendPrivateMessage(userID, msg)
}

func (q *QQBot) SendGroupMessage(groupID int64, msg *message.SendingMessage) {
	q.Client.SendGroupMessage(groupID, msg)
}

func (q *QQBot) Stop() {
	q.Client.Disconnect()
	q.Client.Release()
}

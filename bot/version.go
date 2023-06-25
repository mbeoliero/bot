package bot

import (
	"errors"

	"github.com/mbeoliero/bot/utils"
)

var remoteVersions = map[int]string{
	1: "https://raw.githubusercontent.com/RomiChan/protocol-versions/master/android_phone.json",
	6: "https://raw.githubusercontent.com/RomiChan/protocol-versions/master/android_pad.json",
}

func getRemoteLatestProtocolVersion(protocolType int) ([]byte, error) {
	url, ok := remoteVersions[protocolType]
	if !ok {
		return nil, errors.New("remote version unavailable")
	}

	handler := &utils.HTTPHandler{URL: url}
	resp, err := handler.Bytes()
	if err != nil {
		handler.URL = "https://ghproxy.com/" + url
		return handler.Bytes()
	}

	return resp, nil
}

package bot

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/mattn/go-colorable"
	"github.com/mbeoliero/bot/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrSMSRequestError SMS request error
var ErrSMSRequestError = errors.New("sms request error")

func commonLogin(cli *client.QQClient) error {
	res, err := cli.Login()
	if err != nil {
		return err
	}
	return loginResponseProcessor(cli, res)
}

func printQRCode(imgData []byte) {
	const (
		black = "\033[48;5;0m  \033[0m"
		white = "\033[48;5;7m  \033[0m"
	)
	img, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		log.Panic(err)
	}
	data := img.(*image.Gray).Pix
	bound := img.Bounds().Max.X
	buf := make([]byte, 0, (bound*4+1)*(bound))
	i := 0
	for y := 0; y < bound; y++ {
		i = y * bound
		for x := 0; x < bound; x++ {
			if data[i] != 255 {
				buf = append(buf, white...)
			} else {
				buf = append(buf, black...)
			}
			i++
		}
		buf = append(buf, '\n')
	}
	_, _ = colorable.NewColorableStdout().Write(buf)
}

func qrcodeLogin(cli *client.QQClient) error {
	resp, err := cli.FetchQRCodeCustomSize(1, 2, 1)
	if err != nil {
		return err
	}
	_ = os.WriteFile("qrcode.png", resp.ImageData, 0o644)
	defer func() { _ = os.Remove("qrcode.png") }()

	logrus.Infof("pls use qq app scan code (qrcode.png) : ")
	time.Sleep(time.Second)
	printQRCode(resp.ImageData)
	s, err := cli.QueryQRCodeStatus(resp.Sig)
	if err != nil {
		return err
	}

	prevState := s.State
	for {
		time.Sleep(time.Second)
		s, _ = cli.QueryQRCodeStatus(resp.Sig)
		if s == nil {
			continue
		}
		if prevState == s.State {
			continue
		}

		prevState = s.State
		switch s.State {
		case client.QRCodeCanceled:
			logrus.Fatalf("scan cancel...")
		case client.QRCodeTimeout:
			logrus.Fatalf("scan code timeout...")
		case client.QRCodeWaitingForConfirm:
			logrus.Infof("scan code success, pls confirm on your phone")
		case client.QRCodeConfirmed:
			res, err := cli.QRCodeLogin(s.LoginInfo)
			if err != nil {
				return err
			}
			return loginResponseProcessor(cli, res)
		case client.QRCodeImageFetch, client.QRCodeWaitingForScan:
			// ignore
		}
	}
}

func loginResponseProcessor(cli *client.QQClient, res *client.LoginResponse) error {
	var err error
	for {
		if err != nil {
			return err
		}
		if res.Success {
			return nil
		}
		var text string
		switch res.Error {
		case client.SliderNeededError:
			logrus.Warnf("need slider")
			ticket := getTicket(res.VerifyUrl)
			if ticket == "" {
				logrus.Infof("press enter to continue....")
				utils.ReadLine()
				os.Exit(0)
			}
			res, err = cli.SubmitTicket(ticket)
			continue
		case client.NeedCaptcha:
			logrus.Warnf("need captcha")
			_ = os.WriteFile("captcha.jpg", res.CaptchaImage, 0o644)
			utils.DelFile("captcha.jpg")

			logrus.Warnf("pls enter (captcha.jpg): ")
			text = utils.ReadLine()
			res, err = cli.SubmitCaptcha(text, res.CaptchaSign)
			utils.DelFile("captcha.jpg")
			continue
		case client.SMSNeededError:
			logrus.Warnf("sms needed, input enter send %v sms", res.SMSPhone)
			utils.ReadLine()
			if !cli.RequestSMS() {
				logrus.Warnf("send sms failed, maybe too frequently")
				return errors.WithStack(ErrSMSRequestError)
			}

			logrus.Warn("pls enter sms code: ")
			text = utils.ReadLine()
			res, err = cli.SubmitSMS(text)
			continue
		case client.SMSOrVerifyNeededError:
			logrus.Warnf("sms or verify needed, input enter send %v sms", res.SMSPhone)
			logrus.Warnf("1. send %v sms code", res.SMSPhone)
			logrus.Warnf("2. use qq scan code login")
			logrus.Warn("pls enter (1 - 2): ")
			text = utils.ReadWithDefault("2")

			if strings.Contains(text, "1") {
				if !cli.RequestSMS() {
					logrus.Warnf("send sms failed, maybe too frequently")
					return errors.WithStack(ErrSMSRequestError)
				}

				logrus.Warn("pls enter sms code: ")
				text = utils.ReadLine()
				res, err = cli.SubmitSMS(text)
				continue
			}
			fallthrough
		case client.UnsafeDeviceError:
			logrus.Warnf("account open device lock, pls go to %v verify", res.VerifyUrl)
			logrus.Infof("wait or press enter to continue")

			utils.ReadLineTimeout(time.Second * 5)
			os.Exit(0)
		case client.OtherLoginError, client.UnknownLoginError, client.TooManySMSRequestError:
			msg := res.ErrorMessage
			logrus.Warnf("login failed: %v Code: %v", msg, res.Code)

			switch res.Code {
			case 235:
				logrus.Warnf("device info error, pls try again")
			case 237:
				logrus.Warnf("login too frequently, pls try again later")
			case 45:
				logrus.Warnf("pls verify in app")
			}

			logrus.Infof("press enter continue....")
			utils.ReadLine()
			os.Exit(0)
		}
	}
}

func getTicket(u string) string {
	logrus.Warnf("pls select slider ticker submission: ")
	logrus.Warnf("1. auto submit (default)")
	logrus.Warnf("2. manual submit")
	logrus.Warn("pls enter (1 - 2): ")

	text := utils.ReadLine()
	id := utils.RandomString(8)
	auto := !strings.Contains(text, "2")
	if auto {
		u = strings.ReplaceAll(u, "https://ssl.captcha.qq.com/template/wireless_mqq_captcha.html?", fmt.Sprintf("https://captcha.go-cqhttp.org/captcha?id=%v&", id))
	}

	logrus.Warnf("pls go to this address verify -> %v ", u)
	if !auto {
		logrus.Warn("pls enter ticket: ")
		return utils.ReadLine()
	}

	for count := 120; count > 0; count-- {
		str := fetchCaptcha(id)
		if str != "" {
			return str
		}
		time.Sleep(time.Second)
	}

	logrus.Warnf("verify timeout, pls try again")
	return ""
}

func fetchCaptcha(id string) string {
	handler := utils.HTTPHandler{URL: "https://captcha.go-cqhttp.org/captcha/ticket?id=" + id}
	g, err := handler.Json()
	if err != nil {
		logrus.Debugf("fetch captcha error: %v", err)
		return ""
	}

	if g.Get("ticket").Exists() {
		return g.Get("ticket").String()
	}
	return ""
}

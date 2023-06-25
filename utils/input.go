package utils

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.ilharper.com/x/isatty"
)

var console = bufio.NewReader(os.Stdin)

func ReadLine() (str string) {
	str, _ = console.ReadString('\n')
	str = strings.TrimSpace(str)
	return
}

func ReadLineTimeout(t time.Duration) {
	r := make(chan string)
	go func() {
		select {
		case r <- ReadLine():
		case <-time.After(t):
		}
	}()
	select {
	case <-r:
	case <-time.After(t):
	}
}

func ReadWithDefault(de string) (str string) {
	if isatty.Isatty(os.Stdin.Fd()) {
		return ReadLine()
	}
	logrus.Warnf("no input in terminal, return default: %s.", de)
	return de
}

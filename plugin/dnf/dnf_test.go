package dnf

import (
	"MiniBot/plugin/dnf/service"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestYxdr(t *testing.T) {
	service.Screenshot("跨1", "youxibi")
}

func TestColgNew(t *testing.T) {
	res, _ := service.ColgNews()
	logrus.Infoln(res)
}

func TestCreate(t *testing.T) {
	res, _ := service.GetColgUser()
	logrus.Infoln(res)
}

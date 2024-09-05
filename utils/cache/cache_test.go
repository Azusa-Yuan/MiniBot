package cache

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInit(t *testing.T) {
	logrus.Infoln("ok")
}

func TestGetAvatar(t *testing.T) {
	GetAvatar(1043728417)
}

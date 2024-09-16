package dnf

import (
	"MiniBot/plugin/dnf/service"
	"fmt"
	"testing"
)

func TestYxdr(t *testing.T) {
	service.Screenshot("跨1", "youxibi")
	fmt.Println("why?2354")
	service.Screenshot("跨2", "youxibi")
}

func TestColgNew(t *testing.T) {
	res, _ := service.ColgNews()
	fmt.Println(res)
}

func TestCreate(t *testing.T) {
	res, _ := service.GetColgUser()
	fmt.Println(res)
}

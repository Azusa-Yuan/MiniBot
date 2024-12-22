package bilibili

import (
	"fmt"
	"testing"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
)

func TestBilibil(t *testing.T) {
	searchRes, err := bz.SearchUser(cfg, "3546787012938107")
	fmt.Println(err)
	fmt.Println(searchRes)
}

package ymgal

import (
	"fmt"
	"testing"
	"time"
)

func TestPage(t *testing.T) {
	maxCgPageNumber, _, _ := initPageNumber()
	for i := 1; i <= maxCgPageNumber; i++ {
		err := getPicID(i, cgType)
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println(cgIDList)
	fmt.Println(len(cgIDList))
}

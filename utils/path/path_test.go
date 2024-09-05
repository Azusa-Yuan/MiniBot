package path

import (
	"fmt"
	"testing"
)

func TestPath(t *testing.T) {
	fmt.Println(DataPath)
}

func TestGetPluginDataPath(t *testing.T) {
	fmt.Println(GetDataPath())
}

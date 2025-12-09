package main

import (
	"fmt"

	"github.com/mortalglitch/gator/internal/config"
)

func main() {
	configData, err := config.Read()
	if err != nil {
		fmt.Errorf("Error: ", err)
	}
	fmt.Println(configData.DBURL)

	fmt.Println("Test 2 update and pull: ")
	config.SetUser(configData, "mortalglitch")
	configData2, err2 := config.Read()
	if err2 != nil {
		fmt.Errorf("Test 2 error: ", err2)
	}
	fmt.Println(configData2.DBURL)
	fmt.Println(configData2.CurrentUserNname)
}

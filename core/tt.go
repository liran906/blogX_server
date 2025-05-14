package core

import "fmt"

func Tt() {
	InitIPDB()
	fmt.Println(GetAddress("22.16.58.3"))
	fmt.Println(GetAddress("112.16.58.3"))
	fmt.Println(GetAddress("77.16.58.3"))
	fmt.Println(GetAddress("127.255.0.3"))
	fmt.Println(GetAddress("10.16.58.3"))
}

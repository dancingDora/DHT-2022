package main

import (
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	myself dhtNode
	myIP   string
)

func init() {
	var f *os.File
	f, _ = os.Create("log.txt")
	log.SetOutput(f)
	fmt.Printf("Please input your IP to start :")
	fmt.Scanln(&myIP)
	fmt.Println("IP is set successfully")
	myself = NewNode(myIP)
	myself.Run()
}

func main() {
	RED := color.New(color.FgRed)
	CYAN := color.New(color.FgCyan)

	for {
		var para1, para2, para3, para4 string = "", "", "", ""
		fmt.Scanln(&para1, &para2, &para3, &para4)
		if para1 == "[JOIN]" {
			ok := myself.Join(para2)
			if ok {
				CYAN.Println("[JOIN] Join ", para2, " Successfully !")
			} else {
				RED.Println("[JOIN] Fail to Join ", para2, " !")
			}
			continue
		}
		if para1 == "[CREATE]" {
			myself.Create()
			CYAN.Println("[CREATE] Create Network Successfully in ", myIP)
			continue
		}
		if para1 == "[UPLOAD]" {
			err := Upload(para2, para3, &myself)
			if err != nil {
				RED.Println("[UPLOAD] Fail to Upload ", para2)
			}
			continue
		}
		if para1 == "[DOWNLOAD]" {
			err := download(para2, para3, &myself)
			if err != nil {
				RED.Println("[DOWNLOAD] Fail to Download ", err)
			}
			continue
		}
		if para1 == "[QUIT]" {
			myself.Quit()
			CYAN.Println("[NODE] ", myIP, " Node Quit Successfully !")
			continue
		}
		if para1 == "[RUN]" {
			myself.Run()
			CYAN.Println("[RUN] ", myIP, " is Running ...")
			continue
		} else {
			RED.Println("[ERROR]Failed to Interpret ", para1)
		}
	}
}

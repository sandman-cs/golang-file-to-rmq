package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		sendMessage("Exiting Program, recieved: ")
		sendMessage(fmt.Sprintln(sig))

		//Add Code here to send message to kill all input threads.

		sendMessage("Closing input threads...")
		time.Sleep(time.Second * 5)
		os.Exit(0)
	}()

	for {
		time.Sleep(time.Second * 1)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		fmt.Println(text)
	}

}

func workLoop(workDir string, route string, ext string) {
	defer func() {
		if r := recover(); r != nil {
			sendMessage(fmt.Sprintf("Recovered from error: %s ", r))
		}
	}()

	pRoute := route + conf.ServerName + ".logs"
	for {
		files, err := ioutil.ReadDir(workDir)
		checkError(err)
		time.Sleep(5 * time.Second)
		for _, file := range files {
			sFileName := fmt.Sprintf("%s\\%s", workDir, file.Name())
			if info, err := os.Stat(sFileName); err == nil && !info.IsDir() {
				dat, err := ioutil.ReadFile(sFileName)
				checkError(err)
				if err == nil && strings.Contains(sFileName, ext) {
					//Uncomment for Debug
					//fmt.Println(string(dat))
					pubToRabbit(string(dat), pRoute)
				}
				checkError(os.Remove(sFileName))
			}
		}
	}
}

func stringToLines(s string) []string {
	var lines []string

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		sendMessage(fmt.Sprintln("Error reading standard input:", err))
	}
	return lines
}

// stringContains checkes the srcString for any matches in the
// list, which is an array of strings.
func stringContains(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(a, b) {
			return true
		}
	}
	return false
}

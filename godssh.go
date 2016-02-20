package main

import (
	"flag"
	"fmt"
	"github.com/becks/easyssh"
	"io/ioutil"
	"regexp"
	"runtime"
	"strings"
)

func readfile(file string) []byte {
	dat, err := ioutil.ReadFile(file)
	check(err)
	return dat
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func hashhost(file *string) map[string]string {
	h := make(map[string]string)
	re := regexp.MustCompile("^#")
	fmt.Println("Opening: ", *file)
	content := readfile(*file)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if re.FindString(line) != "" {
			continue
		}
		if len(line) == 0 {
			continue
		}
		kv := strings.Fields(line)
		h[kv[0]] = kv[1]
	}
	fmt.Printf("found %d hosts \n\t%s\n\n", len(h), h)
	return h
}

func init() {
}

func prun(cmd string, k string, v string) {
	ssh := &easyssh.MakeConfig{
		User:   "becks",
		Server: k,
		Key:    "/.ssh/id_rsa",
		Port:   "22",
	}
	fmt.Printf("[%s:%s -> %s]", k, v, cmd)
	response, err := ssh.Run(cmd)
	if err != nil {
		fmt.Println("Can't run remote command: " , err.Error())
	} else {
		fmt.Printf("\n\t%s", response)
	}
}

func pscp(file string, destdir string, k string, v string) {
	ssh := &easyssh.MakeConfig{
		User:   "becks",
		Server: k,
		Key:    "/.ssh/id_rsa",
		Port:   "22",
	}
	fmt.Printf("[%s:%s -> put %s]", k, v, file)
	err := ssh.Scp(file, "/tmp/")
	if err != nil {
		fmt.Println("Can't upload" + err.Error())
	} else {
		fmt.Println("success")
	}
}

func main() {

	file := flag.String("file", "hosts.txt", "file hosts")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	type Result struct {
		response string
		err      error
	}

	h := make(map[string]string)
	h = hashhost(file)
	for k, v := range h {

		prun("uptime", k, v)
		prun("who", k, v)
		pscp("/home/becks/hosts_linux.txt", "/tmp/", k, v)
		prun("ls -la /tmp/hosts_linux.txt", k, v)

	}

}

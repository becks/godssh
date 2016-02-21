package main

import (
	"flag"
	"fmt"
	"github.com/becks/easyssh"
	"io/ioutil"
	"path/filepath"
	"regexp"
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

func prun(ssh_conf *easyssh.MakeConfig, cmd string, k string, v string) {
	fmt.Printf("[%s:%s -> %s]", k, v, cmd)
	response, err := ssh_conf.Run(cmd)
	if err != nil {
		fmt.Println("Can't run remote command: ", err.Error())
	} else {
		fmt.Printf("\n%s", response)
	}
}

func pscp(ssh_conf *easyssh.MakeConfig, file string, destdir string, k string, v string) {
	fmt.Printf("[%s:%s -> put %s destdir: %s]", k, v, file, destdir)
	ssh_conf.Tty = false
	err := ssh_conf.Scp(file, destdir)
	if err != nil {
		fmt.Println(" Can't upload" + err.Error())
	} else {
		fmt.Println(" success")
		prun(ssh_conf, fmt.Sprintf("ls -la %s/%s", destdir, filepath.Base(file)), k, v)
	}
}

func prun_su(ssh_conf *easyssh.MakeConfig, cmd string, k string, v string) {
	fmt.Printf("[%s:%s -> su -c %s]", k, v, cmd)
	response, err := ssh_conf.Run(fmt.Sprintf("su -c \"%s\"", cmd))
	if err != nil {
		fmt.Println("Can't run remote command: ", err.Error())
	} else {
		fmt.Printf("\n%s", response)
	}
}

func main() {

	file := flag.String("file", "", "file hosts")
	flag.Parse()

	type Result struct {
		response string
		err      error
	}

	h := hashhost(file)
	for k, v := range h {
		ssh_conf := &easyssh.MakeConfig{
			User:   "becks",
			Server: k,
			//Key:    "/.ssh/id_rsa",
			Password: "XXXXXX",
			Port:     "22",
			Tty:      false,
			//Shell: "/bin/bash",
			Supassword: "YYYYYY",
		}
		prun(ssh_conf, "who", k, v)
		prun_su(ssh_conf, "who", k, v)
		pscp(ssh_conf, "/home/becks/hosts_linux.txt", "/tmp/", k, v)
	}

}

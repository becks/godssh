package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
  "runtime"
  "time"
  "flag"
  "becks/sshutil"
)

//type hosts struct {
//  ip string
//  host string
//}



func readfile(file string) []byte {
	dat, err := ioutil.ReadFile(file)
	check(err)
	//fmt.Print(string(dat))
	return dat
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}



func hashhost(file *string) map[string]string {
	// rex for comments in file
	h := make(map[string]string)
	//keys := make([]int, len(mymap))
	re := regexp.MustCompile("^#")
	fmt.Println("Opening: ", *file)
	content := readfile(*file)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		// skip comments
		if re.FindString(line) != "" {
			continue
		}
		if len(line) == 0 {
			continue
		}
		kv := strings.Fields(line)
		h[kv[0]] = kv[1]
	}
	fmt.Printf("found %d hosts \n\t%s\n", len(h), h)
	return h
}


func init() {
}

func main() {

  file := flag.String("file", "hosts.txt", "file hosts")
  paral := flag.Bool("p", false, "parallel execution, default false")
  flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
  results := make(chan string)
  timeout := time.After(5 * time.Second) // in 5 seconds the message will come to timeout channel


	sshConfig := &ssh.ClientConfig{
		User: "becks",
		Auth: []ssh.AuthMethod{
			sshutil.PublicKeyFile("/home/becks/.ssh/id_rsa"),
		},
	}

	cmd := &sshutil.SSHCommand{
			Path:   "uptime",
			Env:    []string{"LC_DIR=/"},
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

  //fmt.Println(sshConfig)


	h := make(map[string]string)
	h = hashhost(file)
	//fmt.Println(h)
  for k, v := range(h) {
		client := &sshutil.SSHClient{
			Config: sshConfig,
			Host:   k,
			Port:   22,
		}
    //run(k, v, cmd, client)
    if *paral {
			go func(k string ) {
						results <- fmt.Sprintf("[%s:%s] ", k, v) + sshutil.Run(k, v, cmd, client)
			}(k)
    } else {
						fmt.Println(fmt.Sprintf("[%s:%q] ", k, v) + sshutil.Run(k, v, cmd, client))
    }


  }

  fmt.Println("waiting for results...")
	for i := 0; i < len(h); i++ {
			select {
				case res := <-results:
						fmt.Println(res)
				case <-timeout:
						fmt.Println("Timed out!")
						return
				}
	}


}

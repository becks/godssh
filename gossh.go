package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
  "runtime"
  "time"
)

//type hosts struct {
//  ip string
//  host string
//}

var ifile = os.Args[1]

type SSHClient struct {
	Config *ssh.ClientConfig
	Host   string
	Port   int
}

type SSHCommand struct {
	Path   string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

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

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (client *SSHClient) RunCommand(cmd *SSHCommand) error {
	var (
		session *ssh.Session
		err     error
	)

	if session, err = client.newSession(); err != nil {
		return err
	}
	defer session.Close()
	
	if err = client.prepareCommand(session, cmd); err != nil {
		return err
	}

	err = session.Run(cmd.Path)
	return err
}


func (client *SSHClient) prepareCommand(session *ssh.Session, cmd *SSHCommand) error {
	for _, env := range cmd.Env {
		variable := strings.Split(env, "=")
		if len(variable) != 2 {
			continue
		}

		if err := session.Setenv(variable[0], variable[1]); err != nil {
			return err
		}
	}

	if cmd.Stdin != nil {
		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdin for session: %v", err)
		}
		go io.Copy(stdin, cmd.Stdin)
	}

	if cmd.Stdout != nil {
		stdout, err := session.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		go io.Copy(cmd.Stdout, stdout)
	}

	if cmd.Stderr != nil {
		stderr, err := session.StderrPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stderr for session: %v", err)
		}
		go io.Copy(cmd.Stderr, stderr)
	}

	return nil
}

func (client *SSHClient) newSession() (*ssh.Session, error) {
	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Host, client.Port), client.Config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %s", err)
	}

	modes := ssh.TerminalModes{
		// ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	return session, nil
}


func hashhost(ifile string) map[string]string {
	// rex for comments in file
	h := make(map[string]string)
	//keys := make([]int, len(mymap))
	re := regexp.MustCompile("^#")
	fmt.Println("Opening: ", ifile)
	content := readfile(ifile)
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
		//fmt.Println("-> ", kv)
		h[kv[0]] = kv[1]
	}
	fmt.Printf("found %d hosts \n", len(h))
	return h
}

func run(k string, v string, cmd *SSHCommand, client *SSHClient) (string) {
    fmt.Printf("[%s:%s] Running command: %s\n", v, k ,cmd.Path)
		if err := client.RunCommand(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "command run error: %s\n", err)
			os.Exit(1)
		}
    return "ok"
 
}


func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
  results := make(chan string)
  timeout := time.After(5 * time.Second) // in 5 seconds the message will come to timeout channel


	sshConfig := &ssh.ClientConfig{
		User: "becks",
		Auth: []ssh.AuthMethod{
			PublicKeyFile("/home/becks/.ssh/id_rsa"),
		},
	}

	cmd := &SSHCommand{
			Path:   "uptime ",
			Env:    []string{"LC_DIR=/"},
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

  //fmt.Println(sshConfig)


	h := make(map[string]string)
	h = hashhost(ifile)
	//fmt.Println(h)
  for k, v := range(h) {
		client := &SSHClient{
			Config: sshConfig,
			Host:   k,
			Port:   22,
		}
    //run(k, v, cmd, client)
	  go func(k string ) {
          results <- run(k, v, cmd, client)
		}(k)


  }

    for i := 0; i < len(h); i++ {
        select {
        case res := <-results:
            fmt.Print(res)
        case <-timeout:
            fmt.Println("Timed out!")
            return
        }
    }


}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"os"
	"sync"
	"time"
)

type maild struct {
	from    string
	to      string
	subject string
	body    []byte
}

type client struct {
	serverName string
	localName  string
	conn       net.Conn
	txt        *textproto.Conn
	maild      maild
}

type smtpserver struct {
	host string
	port string
}

var (
	Info  *log.Logger
	Error *log.Logger
)

func (c *client) Command(code int, cmd string) error {
	id, err := c.txt.Cmd(cmd)
	if err != nil {
		return err
	}

	c.txt.StartResponse(id)
	defer c.txt.EndResponse(id)

	if _, _, err := c.txt.ReadCodeLine(code); err != nil {
		return err
	}

	return nil
}

func Dial(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *client) Ehlo() error {

	if err := c.Command(220, "HELO "+c.localName); err != nil {
		return err
	}

	return nil
}

func (c *client) Mail() error {
	_, err := c.txt.Cmd("Mail from:xxx@gmail.com")
	if err != nil {
		return err
	}

	if _, err := c.txt.ReadLineBytes(); err != nil {
		return err
	}

	return nil
}

func (c *client) Rcpt() error {
	_, err := c.txt.Cmd("Rcpt to:jheyu@xxx")
	if err != nil {
		return err
	}

	if _, err := c.txt.ReadLineBytes(); err != nil {
		return err
	}

	return nil
}

func (c *client) Data() error {
	_, err := c.txt.Cmd("DATA")
	if err != nil {
		return err
	}

	if _, err = c.txt.ReadLineBytes(); err != nil {
		return err
	}

	_, err = c.txt.Cmd(string(c.maild.body))
	if err != nil {
		return err
	}

	_, err = c.txt.Cmd(".")
	if err != nil {
		return err
	}

	if _, err := c.txt.ReadLineBytes(); err != nil {
		return err
	}

	return nil
}

func (c *client) Rset() error {
	if _, err := c.txt.Cmd("RSET"); err != nil {
		return err
	}

	return nil
}

func (c *client) Quit() error {
	if _, err := c.txt.Cmd("Quit"); err != nil {
		return err
	}

	return nil
}

func SendMail(addr string, mailt maild) error {
	conn, err := Dial(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	text := textproto.NewConn(conn)
	c := &client{txt: text, conn: conn, serverName: addr, localName: "jheyu.localhost", maild: mailt}

	if err = c.Ehlo(); err != nil {
		return err
	}

	if err = c.Mail(); err != nil {
		return err
	}

	if err = c.Rcpt(); err != nil {
		return err
	}

	if err = c.Data(); err != nil {
		return err
	}

	if err = c.Quit(); err != nil {
		return err
	}

	return nil
}

func main() {

	// Initial log file
	errf, _ := os.OpenFile("stress_error_log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	info, _ := os.OpenFile("stress_info_log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	Error = log.New(errf, "[ERROR] ", log.Ltime|log.Lshortfile)
	Info = log.New(info, "[INFO] ", log.Ltime)

	log.SetOutput(errf)
	log.SetOutput(info)
	defer errf.Close()
	defer info.Close()

	_, _, _, mail_dir := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	// Read sample mail file in directory
	files, _ := ioutil.ReadDir(mail_dir)

	// Initial tcp connect
	addr := "xxx.xxx.xxx.xxx:25"
	//conn, err := Dial(addr)
	//if err != nil {
	//    Error.Println("connection failed")
	//}
	//defer conn.Close()

	//text := textproto.NewConn(conn)
	//var r *io.Reader
	//c := &client{txt: text, r: r, conn: conn, serverName: addr, localName: "jheyu.localhost"}

	// start compute program exec time
	start := time.Now()

	// use goroutines
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}

	// prof program performance
	//pf, err := os.Create("cpu-profile.prof")
	//if err != nil {
	//    Error.Println("create cpu-profile.prof failed")
	//}
	//pprof.StartCPUProfile(pf)
	for _, file := range files {
		wg.Add(1)
		go func(file os.FileInfo) {
			content, err := ioutil.ReadFile(mail_dir + "/" + file.Name())
			if err != nil {
				//Error.Println(file.Name() + " not exists")
				//continue
			}

			mailtest := maild{}
			mailtest.body = content

			//Info.Println(mailtest.from)
			mutex.Lock()
			SendMail(addr, mailtest)
			//if err != nil {
			//    continue
			//}
			//Info.Println("Send mail scuccessed!")
			mutex.Unlock()

			wg.Done()
		}(file)
	}

	wg.Wait()
	//c.Quit()
	//pprof.StopCPUProfile()

	elapsed := time.Since(start)

	fmt.Println(elapsed)
}

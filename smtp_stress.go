package main

import (
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"os"
	"time"
	//"sync"
	"strings"
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
	pipe       *textproto.Pipeline
	contents   []byte
}

var (
	Info  *log.Logger
	Error *log.Logger
	T     int
)

func (c *client) Command(code int, cmd string) error {
	id, err := c.txt.Cmd(cmd)
	if err != nil {
		return err
	}
	c.txt.StartResponse(id)
	defer c.txt.EndResponse(id)

	if _, _, err = c.txt.ReadCodeLine(code); err != nil {
		return err
	}

	return nil
}

func Dial(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		//Info.Println("Dial connection error")
	}

	return conn, nil
}

func (c *client) Ehlo() error {
	if err := c.Command(220, "HELO jheyu.local"); err != nil {
		return err
	}

	if str, err := c.txt.ReadLineBytes(); err != nil {
		Info.Println(string(str))
	}

	return nil
}

func (c *client) Mail() error {
	_, err := c.txt.Cmd("Mail from:xxx@gmail.com")
	if err != nil {
		return err
	}

	//if _, err = c.txt.ReadLineBytes(); err != nil {
	//      Error.Println("Mail: response message")
	//      return err
	//}

	if str, err := c.txt.ReadLineBytes(); err != nil {
		Info.Println(string(str))
	}

	return nil
}

func (c *client) Rcpt() error {
	_, err := c.txt.Cmd("Rcpt to:jheyu@xxx")
	if err != nil {
		return err
	}

	//if _, err = c.txt.ReadLineBytes(); err != nil {
	//      Error.Println("Rcpt: response message")
	//      return err
	//}

	if str, err := c.txt.ReadLineBytes(); err != nil {
		Info.Println(string(str))
	}

	return nil
}

func (c *client) Data() error {
	_, err := c.txt.Cmd("Data")
	if err != nil {
		return err
	}

	if str, err := c.txt.ReadLineBytes(); err != nil {
		Info.Println(string(str))
	}

	_, err = c.txt.Cmd(string(c.contents))
	if err != nil {
		return err
	}

	_, err = c.txt.Cmd(".")
	if err != nil {
		return err
	}

	if str, err := c.txt.ReadLineBytes(); err != nil {
		Info.Println(string(str))
	}

	return nil
}

func (c *client) Rset(index int) error {
	_, err := c.txt.Cmd("Rset")
	if err != nil {
		//return err
	}

	str, _ := c.txt.ReadLineBytes()
	if strings.Contains(string(str), "500") {
		c.Quit()
		T = 1
	}

	return nil
}

func (c *client) Quit() error {
	_, err := c.txt.Cmd("Quit")
	if err != nil {
		return err
	}

	str, _ := c.txt.ReadLineBytes()
	Info.Println(string(str))

	return nil
}

func SendMail(c *client, content []byte, index int) error {

	text := textproto.NewConn(c.conn)
	c.txt = text
	c.contents = content

	if index == 0 || T == 1 {
		if T == 1 {
			T = 0
		}

		if err := c.Ehlo(); err != nil {
			return err
		}
	}
	if err := c.Mail(); err != nil {
		return err
	}

	if err := c.Rcpt(); err != nil {
		return err
	}

	if err := c.Data(); err != nil {
		return err
	}

	if err := c.Rset(index); err != nil {
		return err
	}

	return nil
}

func main() {

	// Initial log file
	info, err := os.OpenFile("info_log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	Info = log.New(info, "[INFO] ", log.Ltime|log.Lshortfile)
	log.SetOutput(info)
	defer info.Close()

	_, _, _, mail_dir := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	files, _ := ioutil.ReadDir(mail_dir)

	addr := "xxx.xxx.xxx.xxx:25"

	conn, err := Dial(addr)
	if err != nil {
		//Info.Println("connection failed")
	}
	defer conn.Close()

	//mailt := maild{}
	//mailt.body = content

	text := textproto.NewConn(conn)
	c := &client{conn: conn, txt: text, serverName: addr, localName: "jheyu.localhost"}
	i := 0
	start := time.Now()

	//var wg sync.WaitGroup
	//var mutex = &sync.Mutex{}

	var content []byte

	for _, file := range files {

		//wg.Add(1)

		//go func(file os.FileInfo) {
		//fmt.Println(file.Name())

		content, _ = ioutil.ReadFile(mail_dir + "/" + file.Name())
		//c.contents = content
		//Info.Println("i:",i)
		//fmt.Println("i:", i)
		//Info.Println(string(c.contents))

		if T == 1 {
			//Info.Println("TTT")
			conn, err = Dial(addr)
			if err != nil {
				//Info.Println("connection failed")
			}
			defer conn.Close()

			c = &client{conn: conn, txt: text, serverName: addr, localName: "jheyu.localhost"}
		}

		//c.contents = content
		//Info.Println(string(c.maild.body))
		//Info.Println("index:", i)

		//mutex.Lock()

		SendMail(c, content, i)

		//mutex.Unlock()

		//wg.Done()

		i++
		//}(file)
	}

	//wg.Wait()

	elapsed := time.Since(start)

	fmt.Println(elapsed)
}

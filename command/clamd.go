package command

import (
	"errors"
	"io"
	"net/url"

	"github.com/asnur/clamav_gofiber/domain"
)

type Clamd struct {
	Address string
}

func (c *Clamd) newConnection() (conn *CLAMDConn, err error) {

	var u *url.URL

	if u, err = url.Parse(c.Address); err != nil {
		return
	}

	return NewCLAMDTcpConn(u.Host)
}

func (c *Clamd) SimpleCommand(command string) (chan *domain.ScanResult, error) {
	conn, err := c.newConnection()
	if err != nil {
		return nil, err
	}

	err = conn.SendCommand(command)
	if err != nil {
		return nil, err
	}

	ch, wg, err := conn.ReadResponse()

	go func() {
		wg.Wait()
		conn.Close()
	}()

	return ch, err
}

/*
Check the daemon's state (should reply with PONG).
*/
func (c *Clamd) Ping() error {
	ch, err := c.SimpleCommand("PING")
	if err != nil {
		return err
	}

	response := <-ch

	if response.Raw != "PONG" {
		return errors.New("invalid response")
	}

	return nil
}

/*
Print program and database versions.
*/
func (c *Clamd) Version() (chan *domain.ScanResult, error) {
	dataArrays, err := c.SimpleCommand("VERSION")
	return dataArrays, err
}

func (c *Clamd) ScanStream(r io.Reader, abort chan bool) (chan *domain.ScanResult, error) {
	conn, err := c.newConnection()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			_, allowRunning := <-abort
			if !allowRunning {
				break
			}
		}
		conn.Close()
	}()

	conn.SendCommand("INSTREAM")

	for {
		buf := make([]byte, CHUNK_SIZE)

		nr, err := r.Read(buf)
		if nr > 0 {
			conn.SendChunk(buf[0:nr])
		}

		if err != nil {
			break
		}

	}

	err = conn.SendEOF()
	if err != nil {
		return nil, err
	}

	ch, wg, _ := conn.ReadResponse()

	go func() {
		wg.Wait()
		conn.Close()
	}()

	return ch, nil
}

func NewClamd(address string) *Clamd {
	clamd := &Clamd{Address: address}
	return clamd
}

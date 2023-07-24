package command

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/asnur/clamav_gofiber/domain"
)

type CLAMDConn struct {
	net.Conn
}

const CHUNK_SIZE = 1024
const TCP_TIMEOUT = time.Second * 2

var ResultRegex = regexp.MustCompile(
	`^(?P<path>[^:]+): ((?P<desc>[^:]+)(\((?P<virhash>([^:]+)):(?P<virsize>\d+)\))? )?(?P<status>FOUND|ERROR|OK)$`,
)

func (conn *CLAMDConn) SendCommand(command string) error {
	commandBytes := []byte(fmt.Sprintf("n%s\n", command))

	_, err := conn.Write(commandBytes)
	return err
}

func (conn *CLAMDConn) SendEOF() error {
	_, err := conn.Write([]byte{0, 0, 0, 0})
	return err
}

func (conn *CLAMDConn) SendChunk(data []byte) error {
	var buf [4]byte
	lenData := len(data)
	buf[0] = byte(lenData >> 24)
	buf[1] = byte(lenData >> 16)
	buf[2] = byte(lenData >> 8)
	buf[3] = byte(lenData >> 0)

	a := buf

	b := make([]byte, len(a))
	copy(b, a[:])

	conn.Write(b)

	_, err := conn.Write(data)
	return err
}

func (c *CLAMDConn) ReadResponse() (chan *domain.ScanResult, *sync.WaitGroup, error) {
	var wg sync.WaitGroup

	wg.Add(1)
	reader := bufio.NewReader(c)
	ch := make(chan *domain.ScanResult)

	go func() {
		defer func() {
			close(ch)
			wg.Done()
		}()

		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				return
			}

			if err != nil {
				return
			}

			line = strings.TrimRight(line, " \t\r\n")
			ch <- ParseResult(line)
		}
	}()

	return ch, &wg, nil
}

func ParseResult(line string) *domain.ScanResult {
	res := &domain.ScanResult{}
	res.Raw = line

	matches := ResultRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		res.Description = "Regex had no matches"
		res.Status = domain.RES_PARSE_ERROR
		return res
	}

	for i, name := range ResultRegex.SubexpNames() {
		switch name {
		case "path":
			res.Path = matches[i]
		case "desc":
			res.Description = matches[i]
		case "virhash":
			res.Hash = matches[i]
		case "virsize":
			i, err := strconv.Atoi(matches[i])
			if err == nil {
				res.Size = i
			}
		case "status":
			switch matches[i] {
			case domain.RES_OK:
			case domain.RES_FOUND:
			case domain.RES_ERROR:
				break
			default:
				res.Description = "Invalid status field: " + matches[i]
				res.Status = domain.RES_PARSE_ERROR
				return res
			}
			res.Status = matches[i]
		}
	}

	return res
}

func NewCLAMDTcpConn(address string) (*CLAMDConn, error) {
	conn, err := net.DialTimeout("tcp", address, TCP_TIMEOUT)

	if err != nil {
		if nerr, isOk := err.(net.Error); isOk && nerr.Timeout() {
			return nil, nerr
		}

		return nil, err
	}

	return &CLAMDConn{Conn: conn}, err
}

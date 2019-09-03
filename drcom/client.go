package drcom

import (
	"crypto/md5"
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/logger"
	"hash"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	_codeIn       = byte(0x03)
	_codeOut      = byte(0x06)
	_type         = byte(0x01)
	_eof          = byte(0x00)
	_controlCheck = byte(0x20)
	_adapterNum   = byte(0x05)
	_ipDog        = byte(0x01)
)

const (
	authIP   = "10.100.61.3"
	authPort = "61440"
)

var (
	_delimiter   = []byte{0x00, 0x00, 0x00, 0x00}
	_emptyIP     = []byte{0, 0, 0, 0}
	_primaryDNS  = []byte{10, 10, 10, 10}
	_dhcpServer  = []byte{0, 0, 0, 0}
	_authVersion = []byte{0x6a, 0x00}
	_magic1      = big.NewInt(1968)
	_magic2      = big.NewInt(int64(0xffffffff))
	_magic3      = big.NewInt(int64(711))
	_hostName    = (func() string {
		name, err := os.Hostname()
		if err != nil {
			return "unknown"
		}
		return name
	})()
)

type Config struct {
	MAC      string        `json:"mac"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	Retry    int           `json:"retry"`
	Timeout  time.Duration `json:"timeout"`
}

type Client struct {
	config         *Config
	md5Ctx         hash.Hash
	salt           []byte // [4:8]
	clientIP       []byte // [20:24]
	md5a           []byte
	tail1          []byte
	tail2          []byte
	keepAliveVer   []byte // [28:30]
	conn           *net.UDPConn
	ChallengeTimes int
	Count          int
	timeout        time.Duration
	retry          int
	logoutCh       chan struct{}
}

func New(cfg *Config) *Client {
	addr := fmt.Sprintf("%s:%s", authIP, authPort)
	conn, err := net.DialTimeout("udp", addr, time.Second)
	if err != nil {
		logger.Errorf("net.DialTimeout('udp', %v, time.Second) error(%v)", addr, err)
		os.Exit(1)
	}

	return &Client{
		conn:           conn.(*net.UDPConn),
		config:         cfg,
		md5Ctx:         md5.New(),
		md5a:           make([]byte, 16),
		tail1:          make([]byte, 16),
		tail2:          make([]byte, 4),
		keepAliveVer:   []byte{0xdc, 0x02},
		clientIP:       make([]byte, 4),
		salt:           make([]byte, 4),
		ChallengeTimes: 0,
		Count:          0,
		logoutCh:       make(chan struct{}, 1),
	}
}

func (c *Client) Start() {
	logger.Info("Starting...")

	// Challenge
	logger.Info("Challenging...")
	for i := 0; i < c.config.Retry; i++ {
		if err := c.Challenge(); err != nil {
			logger.Errorf("Challenge Error #%d: %v", c.ChallengeTimes, err)
			if i == c.retry-1 {
				logger.Error("Retried for 3 times! Exiting...")
				os.Exit(1)
			}
		}
	}
	logger.Info("Successfully challenged")

	// Login
	logger.Info("Login...")
	if err := c.Login(); err != nil {
		logger.Errorf("Login error: %v", err)
		os.Exit(1)
	}
	logger.Info("Successfully logged in")

	logger.Info("Starting keepalive daemon...")
	go c.keepalive()
	logger.Info("Successfully started keepalive")
}

// Close close service.
func (c *Client) Close() error {
	close(c.logoutCh)
	_ = c.conn.Close()
	return nil
}

func (c *Client) WriteWithTimeout(b []byte) (err error) {
	if err = c.conn.SetWriteDeadline(time.Now().Add(time.Second * c.config.Timeout)); err != nil {
		return
	}
	_, err = c.conn.Write(b)
	return
}

func (c *Client) ReadWithTimeout(b []byte) (err error) {
	if err = c.conn.SetReadDeadline(time.Now().Add(time.Second * c.config.Timeout)); err != nil {
		return
	}
	_, err = c.conn.Read(b)
	return
}

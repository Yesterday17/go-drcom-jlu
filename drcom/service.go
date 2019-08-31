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
	authServer = "auth.jlu.edu.cn"
	authPort   = "61440"
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
	MAC      string `json:"mac"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Service struct {
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
	logoutCh       chan struct{}
}

// New create service instance and return.
func New(cfg *Config) (s *Service) {
	s = &Service{
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

	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", authServer, authPort))
	if err != nil {
		logger.Errorf("net.ResolveUDPAddr(udp4, %s) error(%v) ", fmt.Sprintf("%s:%s", authServer, authPort), err)
		os.Exit(1)
	}

	conn, err := net.DialTimeout("udp", udpAddr.String(), time.Second)
	if err != nil {
		logger.Errorf("net.DialUDP(udp, %v, %v) error(%v)", nil, udpAddr, err)
		os.Exit(1)
	}

	s.conn = conn.(*net.UDPConn)
	return
}

func (s *Service) Start() {
	logger.Info("Starting...")

	// Challenge
	logger.Info("Challenging...")
	if err := s.Challenge(); err != nil {
		logger.Errorf("Error #%d: %v", s.ChallengeTimes, err)
		return
	}
	logger.Info("Successfully challenged")

	// Login
	logger.Info("Logining...")
	if err := s.Login(); err != nil {
		logger.Errorf("Login error: %v", err)
		return
	}
	logger.Info("Successfully logged in")

	logger.Info("Starting keepalive daemon...")
	go s.aliveproc()
	logger.Info("Successfully started keepalive")
}

func (s *Service) aliveproc() {
	count := 0
	for {
		select {
		case _, ok := <-s.logoutCh:
			if !ok {
				logger.Info("Keepalive exited")
				return
			}
		default:
		}
		count++
		logger.Infof("Sending keepalive #%d", count)
		if err := s.Alive(); err != nil {
			logger.Errorf("drcomSvc.Alive() error(%v)", err)
			time.Sleep(time.Second * 5)
			continue
		}
		logger.Infof("Keepalive #%d success", count)
		time.Sleep(time.Second * 20)
	}
}

func (s *Service) Logout() {
	logger.Info("Logging out...")
	if err := s.Challenge(); err != nil {
		logger.Errorf("drcomSvc.Challenge(%d) error(%v)", s.ChallengeTimes, err)
		return
	}
	if err := s.logout(); err != nil {
		logger.Errorf("service.logout() error(%v)", err)
		return
	}
	logger.Info("Logged out")
}

// Close close service.
func (s *Service) Close() error {
	close(s.logoutCh)
	_ = s.conn.Close()
	return nil
}

package drcom

import (
	"bytes"
	"errors"
	"github.com/Yesterday17/go-drcom-jlu/logger"
	"math/rand"
)

var (
	ErrMACAddrError  = errors.New("invalid mac address")
	ErrIdentifyError = errors.New("invalid username or password")
	ErrUnknown       = errors.New("login failed: unknown error")
)

func (c *Client) Login() (err error) {
	var r, buf []byte

	buf = c.packetLogin()
	if err = c.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	r = make([]byte, 128)
	if err = c.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
		return
	}
	if r[0] != 0x04 {
		if r[0] == 0x05 {
			if r[4] == 0x0B {
				err = ErrMACAddrError
			} else {
				err = ErrIdentifyError
			}
		} else {
			err = ErrUnknown
		}
		return
	}
	// 保存 tail1. 构造 keep38 要用 md5a(在mkptk中保存) 和 tail1
	// 注销也要用 tail1
	copy(c.tail1, r[23:39])
	return
}

func (c *Client) packetLogin() (buf []byte) {
	var md5a, md5b, md5c, mac []byte

	buf = make([]byte, 0, 334+(len(c.config.Password)-1)/4*4)
	buf = append(buf, _codeIn, _type, _eof, byte(len(c.config.Username)+20)) // [0:4]

	// md5a
	md5a = c.md5([]byte{_codeIn, _type}, c.salt, []byte(c.config.Password))
	copy(c.md5a, md5a)
	buf = append(buf, md5a...) // [4:20]

	// username
	user := make([]byte, 36)
	copy(user, c.config.Username)
	buf = append(buf, user...)                    // [20:56]
	buf = append(buf, _controlCheck, _adapterNum) //[56:58]

	// md5a xor mac
	mac, _ = MACHex2Bytes(c.config.MAC)
	for i := 0; i < 6; i++ {
		mac[i] = mac[i] ^ c.md5a[i]
	}
	buf = append(buf, mac...) // [58:64]

	// md5b
	md5b = c.md5([]byte{0x01}, []byte(c.config.Password), []byte(c.salt), []byte{0x00, 0x00, 0x00, 0x00})
	buf = append(buf, md5b...)                      // [64:80]
	buf = append(buf, byte(0x01))                   // [80:81]
	buf = append(buf, c.clientIP...)                // [81:85]
	buf = append(buf, bytes.Repeat(_emptyIP, 3)...) // [85:97]

	// md5c
	tmp := make([]byte, len(buf))
	copy(tmp, buf)
	tmp = append(tmp, []byte{0x14, 0x00, 0x07, 0x0b}...)
	md5c = c.md5(tmp)
	buf = append(buf, md5c[:8]...)   // [97:105]
	buf = append(buf, _ipDog)        // [105:106]
	buf = append(buf, _delimiter...) // [106:110]
	hostname := make([]byte, 32)
	copy(hostname, []byte(_hostName))
	buf = append(buf, hostname...)                       // [110:142]
	buf = append(buf, _primaryDNS...)                    // [142:146]
	buf = append(buf, _dhcpServer...)                    // [146:150]
	buf = append(buf, _emptyIP...)                       // secondary dns, [150:154]
	buf = append(buf, bytes.Repeat(_delimiter, 2)...)    // [154,162]
	buf = append(buf, []byte{0x94, 0x00, 0x00, 0x00}...) // [162,166]
	buf = append(buf, []byte{0x06, 0x00, 0x00, 0x00}...) // [166,170]
	buf = append(buf, []byte{0x02, 0x00, 0x00, 0x00}...) // [170,174]
	buf = append(buf, []byte{0xf0, 0x23, 0x00, 0x00}...) // [174,178]
	buf = append(buf, []byte{0x02, 0x00, 0x00, 0x00}...) // [178,182]
	buf = append(buf, []byte{
		0x44, 0x72, 0x43, 0x4f,
		0x4d, 0x00, 0xcf, 0x07}...) // [182,190]
	buf = append(buf, 0x6a)                              // [190,191]
	buf = append(buf, bytes.Repeat([]byte{0x00}, 55)...) // [191:246]
	buf = append(buf, []byte{
		0x33, 0x64, 0x63, 0x37,
		0x39, 0x66, 0x35, 0x32,
		0x31, 0x32, 0x65, 0x38,
		0x31, 0x37, 0x30, 0x61,
		0x63, 0x66, 0x61, 0x39,
		0x65, 0x63, 0x39, 0x35,
		0x66, 0x31, 0x64, 0x37,
		0x34, 0x39, 0x31, 0x36,
		0x35, 0x34, 0x32, 0x62,
		0x65, 0x37, 0x62, 0x31,
	}...) // [246:286]
	buf = append(buf, bytes.Repeat([]byte{0x00}, 24)...) // [286:310]
	buf = append(buf, _authVersion...)                   // [310:312]
	buf = append(buf, 0x00)                              // [312:313]
	pwdLen := len(c.config.Password)
	if pwdLen > 16 {
		pwdLen = 16
	}
	buf = append(buf, byte(pwdLen)) // [313:314]
	ror := c.ror(c.md5a, []byte(c.config.Password))
	buf = append(buf, ror[:pwdLen]...)       // [314:314+pwdLen]
	buf = append(buf, []byte{0x02, 0x0c}...) // [314+l:316+l]
	tmp = make([]byte, 0, len(buf))
	copy(tmp, buf)
	tmp = append(tmp, []byte{0x01, 0x26, 0x07, 0x11, 0x00, 0x00}...)
	tmp = append(tmp, mac[:4]...)
	sum := c.checkSum(tmp)
	buf = append(buf, sum[:4]...)            // [316+l,320+l]
	buf = append(buf, []byte{0x00, 0x00}...) // [320+l,322+l]
	buf = append(buf, mac...)                // [322+l,328+l]
	zeroCount := (4 - pwdLen%4) % 4
	buf = append(buf, bytes.Repeat([]byte{0x00}, zeroCount)...)
	for i := 0; i < 2; i++ {
		buf = append(buf, byte(rand.Int()))
	}
	return
}

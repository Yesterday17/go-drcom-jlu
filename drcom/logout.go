package drcom

import (
	"errors"
	"github.com/Yesterday17/go-drcom-jlu/logger"
)

func (c *Client) logout() (err error) {
	var r, buf []byte

	buf = c.packetLogout()
	if err = c.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	r = make([]byte, 512)
	if err = c.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
		return
	}
	if r[0] != 0x04 {
		err = errors.New("failed to logout: unknown error")
	}
	return
}

func (c *Client) packetLogout() (buf []byte) {
	var md5, mac []byte

	buf = make([]byte, 0, 80)
	buf = append(buf, _codeOut, _type, _eof, byte(len(c.config.Username)+20))

	// md5
	md5 = c.md5([]byte{_codeOut, _type}, c.salt, []byte(c.config.Password))
	buf = append(buf, md5...)
	tmp := make([]byte, 36)
	copy(tmp, c.config.Username)
	buf = append(buf, tmp...)
	buf = append(buf, _controlCheck, _adapterNum)

	// md5 xor mac
	mac, _ = MACHex2Bytes(c.config.MAC)
	for i := 0; i < 6; i++ {
		mac[i] = mac[i] ^ md5[i]
	}
	buf = append(buf, mac...)     // [58:64]
	buf = append(buf, c.tail1...) // [64:80]
	return
}

package drcom

import (
	"errors"
	"github.com/Yesterday17/go-drcom-jlu/logger"
)

func (s *Service) logout() (err error) {
	var r, buf []byte

	buf = s.packetLogout()
	if err = s.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	r = make([]byte, 512)
	if err = s.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
		return
	}
	if r[0] != 0x04 {
		err = errors.New("failed to logout: unknown error")
	}
	return
}

func (s *Service) packetLogout() (buf []byte) {
	var md5, mac []byte

	buf = make([]byte, 0, 80)
	buf = append(buf, _codeOut, _type, _eof, byte(len(s.config.Username)+20))

	// md5
	md5 = s.md5([]byte{_codeOut, _type}, s.salt, []byte(s.config.Password))
	buf = append(buf, md5...)
	tmp := make([]byte, 36)
	copy(tmp, s.config.Username)
	buf = append(buf, tmp...)
	buf = append(buf, _controlCheck, _adapterNum)

	// md5 xor mac
	mac, _ = MACHex2Bytes(s.config.MAC)
	for i := 0; i < 6; i++ {
		mac[i] = mac[i] ^ md5[i]
	}
	buf = append(buf, mac...)     // [58:64]
	buf = append(buf, s.tail1...) // [64:80]
	return
}

package drcom

import (
	"bytes"
	"errors"
	"github.com/Yesterday17/go-drcom-jlu/logger"
	"math/rand"
)

var ErrChallengeHeadError = errors.New("challenge receive head is not correct")

func (s *Service) Alive() (err error) {
	var r, buf []byte

	buf = s.buf38()
	if err = s.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	r = make([]byte, 128)
	if err = s.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
		return
	}
	s.keepAliveVer[0] = r[28]
	s.keepAliveVer[1] = r[29]
	if s.extra() {
		buf = s.buf40(true, true)
		if err = s.WriteWithTimeout(buf); err != nil {
			logger.Errorf("conn.Write(%v) error(%v)", buf, err)
			return
		}
		r = make([]byte, 512)
		if err = s.ReadWithTimeout(r); err != nil {
			logger.Errorf("conn.Read() error(%v)", err)
			return
		}
		s.Count++
	}
	// 40_1
	buf = s.buf40(true, false)
	if err = s.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	r = make([]byte, 64)
	if err = s.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
		return
	}
	s.Count++
	copy(s.tail2, r[16:20])
	// 40_2
	buf = s.buf40(false, false)
	if err = s.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	if err = s.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
	}
	s.Count++
	return
}

func (s *Service) buf38() (buf []byte) {
	buf = make([]byte, 0, 38)
	buf = append(buf, byte(0xff))                       // [0:1]
	buf = append(buf, s.md5a...)                        // [1:17]
	buf = append(buf, bytes.Repeat([]byte{0x00}, 3)...) // [17:20]
	buf = append(buf, s.tail1...)                       // [20:36]
	for i := 0; i < 2; i++ {                            // [36:38]
		buf = append(buf, byte(rand.Int()))
	}
	return
}

func (s *Service) buf40(first, extra bool) (buf []byte) {
	buf = make([]byte, 0, 40)
	buf = append(buf, []byte{0x07, byte(s.Count), 0x28, 0x00, 0x0b}...) // [0:5]
	// keep40_1   keep40_2
	// 发送  接收  发送  接收
	// 0x01 0x02 0x03 0xx04
	// [5:6]
	if first || extra { //keep40_1 keep40_extra 是 0x01
		buf = append(buf, byte(0x01))
	} else {
		buf = append(buf, byte(0x03))
	}
	// [6:8]
	if extra {
		buf = append(buf, []byte{0x0f, 0x27}...)
	} else {
		buf = append(buf, []byte{s.keepAliveVer[0], s.keepAliveVer[1]}...)
	}
	// [8:10]
	for i := 0; i < 2; i++ {
		buf = append(buf, byte(rand.Int()))
	}
	buf = append(buf, bytes.Repeat([]byte{0x00}, 6)...) //[10:16]
	buf = append(buf, s.tail2...)                       // [16:20]
	buf = append(buf, bytes.Repeat([]byte{0x00}, 4)...) //[20:24]
	if !first {
		tmp := make([]byte, len(buf))
		copy(tmp, buf)
		tmp = append(tmp, s.clientIP...)
		sum := s.crc(tmp)
		buf = append(buf, sum...)                           // [24:28]
		buf = append(buf, s.clientIP...)                    // [28:32]
		buf = append(buf, bytes.Repeat([]byte{0x00}, 8)...) //[32:40]
	}
	if len(buf) < 40 {
		buf = append(buf, bytes.Repeat([]byte{0x00}, 40-len(buf))...)
	}
	return
}

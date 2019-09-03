package drcom

import (
	"bytes"
	"errors"
	"github.com/Yesterday17/go-drcom-jlu/logger"
	"math/rand"
	"time"
)

var ErrChallengeHeadError = errors.New("challenge receive head is not correct")

func (c *Client) Alive() (err error) {
	var r, buf []byte

	buf = c.buf38()
	if err = c.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	r = make([]byte, 128)
	if err = c.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
		return
	}
	c.keepAliveVer[0] = r[28]
	c.keepAliveVer[1] = r[29]
	if c.extra() {
		buf = c.buf40(true, true)
		if err = c.WriteWithTimeout(buf); err != nil {
			logger.Errorf("conn.Write(%v) error(%v)", buf, err)
			return
		}
		r = make([]byte, 512)
		if err = c.ReadWithTimeout(r); err != nil {
			logger.Errorf("conn.Read() error(%v)", err)
			return
		}
		c.Count++
	}
	// 40_1
	buf = c.buf40(true, false)
	if err = c.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	r = make([]byte, 64)
	if err = c.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
		return
	}
	c.Count++
	copy(c.tail2, r[16:20])
	// 40_2
	buf = c.buf40(false, false)
	if err = c.WriteWithTimeout(buf); err != nil {
		logger.Errorf("conn.Write(%v) error(%v)", buf, err)
		return
	}
	if err = c.ReadWithTimeout(r); err != nil {
		logger.Errorf("conn.Read() error(%v)", err)
	}
	c.Count++
	return
}

func (c *Client) buf38() (buf []byte) {
	buf = make([]byte, 0, 38)
	buf = append(buf, byte(0xff))                       // [0:1]
	buf = append(buf, c.md5a...)                        // [1:17]
	buf = append(buf, bytes.Repeat([]byte{0x00}, 3)...) // [17:20]
	buf = append(buf, c.tail1...)                       // [20:36]
	for i := 0; i < 2; i++ {                            // [36:38]
		buf = append(buf, byte(rand.Int()))
	}
	return
}

func (c *Client) buf40(first, extra bool) (buf []byte) {
	buf = make([]byte, 0, 40)
	buf = append(buf, []byte{0x07, byte(c.Count), 0x28, 0x00, 0x0b}...) // [0:5]
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
		buf = append(buf, []byte{c.keepAliveVer[0], c.keepAliveVer[1]}...)
	}
	// [8:10]
	for i := 0; i < 2; i++ {
		buf = append(buf, byte(rand.Int()))
	}
	buf = append(buf, bytes.Repeat([]byte{0x00}, 6)...) //[10:16]
	buf = append(buf, c.tail2...)                       // [16:20]
	buf = append(buf, bytes.Repeat([]byte{0x00}, 4)...) //[20:24]
	if !first {
		tmp := make([]byte, len(buf))
		copy(tmp, buf)
		tmp = append(tmp, c.clientIP...)
		sum := c.crc(tmp)
		buf = append(buf, sum...)                           // [24:28]
		buf = append(buf, c.clientIP...)                    // [28:32]
		buf = append(buf, bytes.Repeat([]byte{0x00}, 8)...) //[32:40]
	}
	if len(buf) < 40 {
		buf = append(buf, bytes.Repeat([]byte{0x00}, 40-len(buf))...)
	}
	return
}

func (c *Client) keepalive() {
	count := 0
	for {
		select {
		case _, ok := <-c.logoutCh:
			if !ok {
				logger.Debug("☑ Exited keepalive daemon")
				return
			}
		default:
			count++
			logger.Debugf("- Sending keepalive #%d", count)
			if err := c.Alive(); err != nil {
				logger.Errorf("drcom.keepalive.Alive() error(%v)", err)
				time.Sleep(time.Second * 5)
				continue
			}
			logger.Debugf("- Keepalive #%d success", count)
			time.Sleep(time.Second * 20)
		}

	}
}

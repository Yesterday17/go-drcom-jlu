package drcom

import (
	"time"
)

func (s *Service) WriteWithTimeout(b []byte) (err error) {
	if err = s.conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		return
	}
	_, err = s.conn.Write(b)
	return
}

func (s *Service) ReadWithTimeout(b []byte) (err error) {
	if err = s.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		return
	}
	_, err = s.conn.Read(b)
	return
}

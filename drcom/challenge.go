package drcom

import (
	"fmt"
	"math/rand"
)

func (s *Service) Challenge() error {
	var (
		response []byte
		packet   = []byte{
			0x01, (byte)(0x02 + s.ChallengeTimes),
			byte(rand.Int()), byte(rand.Int()),
			0x6a, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		}
	)
	if err := s.WriteWithTimeout(packet); err != nil {
		s.ChallengeTimes++
		return fmt.Errorf("conn.Write(%v) error(%v)", packet, err)
	}

	response = make([]byte, 76)
	if err := s.ReadWithTimeout(response); err != nil {
		s.ChallengeTimes++
		return fmt.Errorf("conn.Read() error(%v)", err)
	}

	if response[0] == 0x02 {
		copy(s.salt, response[4:8])
		copy(s.clientIP, response[20:24])
		return nil
	}

	s.ChallengeTimes++
	return ErrChallengeHeadError
}

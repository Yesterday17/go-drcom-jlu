package drcom

import (
	"fmt"
	"math/rand"
)

func (c *Client) Challenge() error {
	var (
		response []byte
		packet   = []byte{
			0x01, (byte)(0x02 + c.ChallengeTimes),
			byte(rand.Int()), byte(rand.Int()),
			0x6a, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		}
	)
	if err := c.WriteWithTimeout(packet); err != nil {
		c.ChallengeTimes++
		return fmt.Errorf("conn.Write(%v) error(%v)", packet, err)
	}

	response = make([]byte, 76)
	if err := c.ReadWithTimeout(response); err != nil {
		c.ChallengeTimes++
		return fmt.Errorf("conn.Read() error(%v)", err)
	}

	if response[0] == 0x02 {
		copy(c.salt, response[4:8])
		copy(c.clientIP, response[20:24])
		return nil
	}

	c.ChallengeTimes++
	return ErrChallengeHeadError
}

package drcom

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
)

const (
	_defaultMACHexLen = 12
)

var (
	// ErrMACAddrLenError mac-len-err
	ErrMACAddrLenError = errors.New("length of mac address is not correct")
)

func (s *Service) md5(items ...[]byte) (ret []byte) {
	for _, v := range items {
		s.md5Ctx.Write(v)
	}
	ret = s.md5Ctx.Sum(nil)
	s.md5Ctx.Reset()
	return
}

// MACHex2Bytes convert mac address to bytes, the input mac format should be 2a:1b:4c:fe:a9:e9.
func MACHex2Bytes(mac string) (res []byte, err error) {
	as := strings.Replace(mac, ":", "", -1)
	if len(as) != _defaultMACHexLen {
		err = ErrMACAddrLenError
		return
	}
	res = make([]byte, 0, 6)
	return hex.DecodeString(as)
}

func (s *Service) ror(md5a, password []byte) (ret []byte) {
	l := len(password)
	ret = make([]byte, l)
	for i := 0; i < l; i++ {
		x := md5a[i] ^ password[i]
		ret[i] = (byte)((x << 3) + (x >> 5))
	}
	return
}

func (s *Service) checkSum(data []byte) (ret []byte) {
	// 1234 = 0x_00_00_04_d2
	sum := []byte{0x00, 0x00, 0x04, 0xd2}
	l := len(data)
	i := 0
	//0123_4567_8901_23
	for ; i+3 < l; i = i + 4 {
		//abcd ^ 3210
		//abcd ^ 7654
		//abcd ^ 1098
		sum[0] ^= data[i+3]
		sum[1] ^= data[i+2]
		sum[2] ^= data[i+1]
		sum[3] ^= data[i]
	}
	if i < l {
		//剩下_23
		//i=12,len=14
		tmp := make([]byte, 4)
		for j := 3; j >= 0 && i < l; j-- {
			//j=3 tmp = 0 0 0 2  i=12  13
			//j=2 tmp = 0 0 3 2  i=13  14
			tmp[j] = data[i]
			i++
		}
		for j := 0; j < 4; j++ {
			sum[j] ^= tmp[j]
		}
	}
	var x = big.NewInt(int64(0))
	tmpBytes := x.SetBytes(sum).Mul(x, _magic1).Add(x, _magic2).Bytes()
	l = len(tmpBytes)
	i = 0
	ret = make([]byte, 4)
	for j := l - 1; j >= 0 && i < 4; j-- {
		ret[i] = tmpBytes[j]
		i++
	}
	return
}

func (s *Service) extra() bool {
	return s.Count%21 == 0
}

func (s *Service) crc(buf []byte) (ret []byte) {
	sum := make([]byte, 2)
	l := len(buf)
	for i := 0; i < l-1; i += 2 {
		sum[0] ^= buf[i+1]
		sum[1] ^= buf[i]
	}
	x := big.NewInt(int64(0))
	tmpBytes := x.SetBytes(sum).Mul(x, _magic3).Bytes()
	ret = make([]byte, 4)
	l = len(tmpBytes)
	for i := 0; i < 4 && l > 0; i++ {
		l--
		ret[i] = tmpBytes[l]
	}
	return
}

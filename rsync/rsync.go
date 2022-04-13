package rsync

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
)

var logtag string = "[Rsync]"

// generate hash table from file in []byte
// hash table contain differential information structured as below
// for one block:
// +---+-----+----------------+---------------------------+
// |key|chunk|rolling checksum|       md5 checksum        |
// +---+-----+----------------+---------------------------+
// | 2 |  4  |        4       |             16            |
// +---+-----+----------------+---------------------------+
// key: high 16 bit of rolling checksum
// chunk: related block index of this hash record
// rolling checksum: 32 bit Adler-32 checksum
// md5 checksum: 128 bit MD5 checksum
func GenerateHashTable(absPath string) []byte {
	var count int64
	var result []byte
	var buff []byte
	data := make([]byte, config.TruncateBlockSize)
	count = 0
	for {
		n, err := fsops.Read(absPath, data, count)

		if n != config.TruncateBlockSize {
			buff = data
		} else if n > 0 {
			buff = data[0:n]
		}
		// calculate record
		chunk := uint32(count / config.TruncateBlockSize)
		rc := getRollingChecksum(buff)
		md5 := common.GetByteMd5(buff)
		key := uint16(rc >> 16)
		record := generateOneRow(key, chunk, rc, md5)

		// add to table
		common.MergeArray(result, record)

		if err != nil {
			break
		}

		count = count + int64(n)
	}
	return result
}

func getRollingChecksum(b []byte) uint32 {
	length := uint32(len(b))
	var s1, s2, i uint32
	s1 = 0
	s2 = 0
	i = 0
	for i = 0; i < (length - 4); i = i + 4 {
		s2 = s2 + 4*(s1+uint32(b[i])) + uint32(3*b[i+1]+2*b[i+2]+b[i+3])
		s1 = s1 + uint32(b[i]+b[i+1]+b[i+2]+b[i+3])
	}
	for ; i < length; i++ {
		s1 = s1 + uint32(b[i])
		s2 = s2 + s1
	}
	return (s1 & 0xffff) + (s2 << 16)
}

func generateOneRow(key uint16, chunk uint32, rc uint32, md5 []byte) []byte {
	var buf bytes.Buffer
	b2 := make([]byte, 2)
	b4 := make([]byte, 4)

	binary.BigEndian.PutUint16(b2, key)
	buf.Write([]byte(b2))
	binary.BigEndian.PutUint32(b4, chunk)
	buf.Write([]byte(b4))
	binary.BigEndian.PutUint32(b4, rc)
	buf.Write([]byte(b4))
	buf.Write(md5)

	return buf.Bytes()
}

func extractFromOneRow(b []byte) (key uint16, chunk uint32, rc uint32, md5 []byte, err error) {
	if len(b) != 26 {
		err = errors.New("invalid length")
		return
	}
	key = binary.BigEndian.Uint16(b[0:2])
	chunk = binary.BigEndian.Uint32(b[2:6])
	rc = binary.BigEndian.Uint32(b[6:10])
	md5 = b[10:26]
	return
}

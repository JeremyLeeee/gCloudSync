package rsync

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gcloudsync/common"
	"gcloudsync/config"
	"gcloudsync/fsops"
	"log"
)

var logtag string = "[Rsync]"

const (
	OpDiffData byte = 24 + iota
	OpLocalData
)

type CheckSums struct {
	key   uint16
	chunk uint32
	rc    uint32
	md5   [16]byte
}

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
func GetHashTable(absPath string) []byte {
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
		key := getKey(rc)
		record := getOneRow(key, chunk, rc, md5)

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

func getKey(rc uint32) uint16 {
	return uint16(rc >> 16)
}

func getOneRow(key uint16, chunk uint32, rc uint32, md5 []byte) []byte {
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

func getLocalDataRecord(start uint32, chunkIndex uint32) []byte {
	var buf bytes.Buffer
	b1 := make([]byte, 1)
	b4 := make([]byte, 4)

	b1[0] = OpLocalData
	buf.Write([]byte(b1))
	binary.BigEndian.PutUint32(b4, start)
	buf.Write([]byte(b4))
	binary.BigEndian.PutUint32(b4, chunkIndex)
	buf.Write([]byte(b4))

	return buf.Bytes()
}

func getDiffDataRecord(start uint32, end uint32, data []byte) []byte {
	var buf bytes.Buffer
	b1 := make([]byte, 1)
	b4 := make([]byte, 4)

	b1[0] = OpDiffData
	buf.Write([]byte(b1))
	binary.BigEndian.PutUint32(b4, start)
	buf.Write([]byte(b4))
	binary.BigEndian.PutUint32(b4, end)
	buf.Write([]byte(b4))
	buf.Write(data)

	return buf.Bytes()
}

func extractFromOneRow(b []byte) (cks CheckSums, err error) {
	if len(b) != 26 {
		err = errors.New("invalid length")
		return
	}
	cks.key = binary.BigEndian.Uint16(b[0:2])
	cks.chunk = binary.BigEndian.Uint32(b[2:6])
	cks.rc = binary.BigEndian.Uint32(b[6:10])
	copy(cks.md5[:], b[10:26])
	return
}

// input does not contain tag
func extractDiff(b []byte) (start int, end int, err error) {
	if len(b) != 8 {
		return 0, 0, errors.New("invalid input length")
	}
	start = int(binary.BigEndian.Uint32(b[0:4]))
	end = int(binary.BigEndian.Uint32(b[4:8]))
	if start > end {
		return 0, 0, errors.New("data bound error")
	}
	return
}

// input does not contain tag
func extractLocal(b []byte) (start int, chunk int, err error) {
	if len(b) != 8 {
		return 0, 0, errors.New("invalid input length")
	}
	start = int(binary.BigEndian.Uint32(b[0:4]))
	chunk = int(binary.BigEndian.Uint32(b[4:8]))
	return
}

// differetial data record structured as below
// +-----+-------+-------+---------------+
// | tag | start |  end  |      data     |
// +-----+-------+-------+---------------+
// |  1  |   4   |   4   |  end - start  |
// +-----+-------+-------+---------------+
// where tag is OpDiffData
// local data record structured as below
// +-----+-------+--------------+
// | tag | start | chunk index  |
// +-----+-------+--------------+
// |  1  |   4   |      4       |
// +-----+-------+--------------+
// where tag is OpLocalData
func GetDiff(table []byte, absPath string) (diff []byte, err error) {
	if len(table)%26 != 0 {
		return nil, errors.New("invalid table len")
	}
	offset := 0
	// construct hash table
	m := make(map[uint16][]CheckSums)
	for {
		row := table[offset : offset+26]
		cks, err := extractFromOneRow(row)
		common.ErrorHandleDebug(logtag, err)
		m[cks.key] = append(m[cks.key], cks)
		offset = offset + 26
		if offset == len(table) {
			break
		}
	}

	// scan file
	data, err := fsops.ReadAll(absPath)
	common.ErrorHandleDebug(logtag, err)
	blockSize := config.TransferBlockSize
	diff = make([]byte, 0)
	pos := 0
	for {
		if offset+blockSize > len(data) {
			// last block
			dr := getDiffDataRecord(uint32(pos), uint32(offset), data[pos:])
			diff = common.MergeArray(diff, dr)
			break
		}
		block := data[offset : offset+blockSize]
		rc := getRollingChecksum(block)
		key := getKey(rc)
		if len(m[key]) == 0 {
			// not match, slide to next block
			offset++
		} else {
			// key match
			// check rolling checksum
			for _, cks := range m[key] {
				if cks.rc == rc {
					// check md5
					md5 := common.GetByteMd5(block)
					if bytes.Equal(cks.md5[:], md5) {
						// match
						// before write local data record, diff data should be write first
						dr := getDiffDataRecord(uint32(pos), uint32(offset), data[pos:offset])
						diff = common.MergeArray(diff, dr)
						// update pos and offset
						offset = offset + blockSize
						pos = offset
						// write local data record
						record := getLocalDataRecord(uint32(offset), cks.chunk)
						diff = common.MergeArray(diff, record)
						break
					}
				}
			} // for
		}
	} // for

	return
}

func ReformFile(diff []byte, absPath string) (err error) {
	// create a temp file with .tmp postfix
	tmpPath := absPath + ".tmp"
	err = fsops.Create(tmpPath)
	common.ErrorHandleDebug(logtag, err)

	pos := 0
	for {
		if pos == len(diff) {
			// reform file finished
			break
		}
		originalData, err := fsops.ReadAll(absPath)
		common.ErrorHandleDebug(logtag, err)

		// get tag
		tag := diff[pos]
		pos++

		if tag == OpDiffData {
			start, end, err := extractDiff(diff[pos : pos+8])
			common.ErrorHandleDebug(logtag, err)
			pos = pos + 8
			// write new data to temp file
			newData := diff[pos : pos+end-start]
			err = fsops.WriteAllAt(tmpPath, newData, start)
			common.ErrorHandleDebug(logtag, err)
			pos = pos + end - start
		} else if tag == OpLocalData {
			start, trunk, err := extractLocal(diff[pos : pos+8])
			common.ErrorHandleDebug(logtag, err)
			pos = pos + 8

			// get block from original file
			begin := trunk * config.TruncateBlockSize
			end := (trunk + 1) * config.TruncateBlockSize
			if end > len(originalData) {
				end = len(originalData)
			}
			originalBlock := originalData[begin:end]

			// write to temp file
			err = fsops.WriteAllAt(tmpPath, originalBlock, start)
			common.ErrorHandleDebug(logtag, err)
		} else {
			log.Panicln(logtag, "invalid tag")
		}
	}
	fsops.CloseCurrentFile()
	fsops.Delete(absPath)
	fsops.Rename(tmpPath, absPath)
	log.Println(logtag, "reform file done.")
	return
}

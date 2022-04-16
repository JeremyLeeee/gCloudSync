package metadata

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gcloudsync/internal/common"
)

var Sig = [14]byte{103, 67, 108, 111, 117, 100, 83, 121, 110, 99, 50, 48, 50, 50}
var logtag string = "[Header]"

type Header struct {
	// header to specify data package
	Signature [14]byte
	Tag       common.SysOp
	Length    uint32
	Last      uint32
}

func NewHeader(len uint32, tag common.SysOp, last uint32) Header {
	// signature is "gCloudSync2022"
	return Header{Signature: Sig, Tag: tag, Length: len, Last: last}
}

func (h Header) ToByteArray() (b []byte, err error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, h); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GetHeaderFromData(b []byte) (header Header, err error) {
	var tempHeader Header
	// log.Println(logtag, "buffer len:", len(b))
	if len(b) < 24 {
		// invalid length for header
		return tempHeader, errors.New("invalid length")
	}

	buf := bytes.NewReader(b[0:24])
	if err := binary.Read(buf, binary.BigEndian, &tempHeader); err != nil {
		return tempHeader, err
	}

	if bytes.Equal(tempHeader.Signature[:], Sig[:]) {
		// valid
		return tempHeader, nil
	} else {
		return tempHeader, errors.New("invalid signature")
	}
}

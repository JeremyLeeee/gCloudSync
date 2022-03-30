package metadata

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gcloudsync/common"
)

type Header struct {
	// header to specify data package
	Signature [14]byte
	Tag       common.SysOp
	Length    uint64
}

func NewHeader(len uint64, tag common.SysOp) Header {
	// signature is "gCloudSync2022"
	sig := [14]byte{103, 67, 108, 111, 117, 100, 83, 121, 110, 99, 50, 48, 50, 50}
	return Header{Signature: sig, Tag: tag, Length: len}
}

func (h Header) ToByteArray() (b []byte, err error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, h); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (h Header) GetHeaderFromData(b []byte) (tag common.SysOp, length uint64, err error) {
	if len(b) != 24 {
		// invalid length for header
		return 0, 0, errors.New("invalid length")
	}
	var tempHeader Header

	buf := bytes.NewReader(b)
	if err := binary.Read(buf, binary.BigEndian, &tempHeader); err != nil {
		return 0, 0, err
	}

	if bytes.Equal(tempHeader.Signature[:], h.Signature[:]) {
		// valid
		return tempHeader.Tag, tempHeader.Length, nil
	} else {
		return 0, 0, errors.New("invalid signature")
	}
}

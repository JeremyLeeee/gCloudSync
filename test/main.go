package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gcloudsync/common"
)

func main() {
	b := []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40, 0xff, 0x01, 0x02, 0x03, 0xbe, 0xef}
	r := bytes.NewReader(b)

	var data struct {
		Mine [3]byte
		PI   float64
		Uate uint8
		Too  common.SysOp
	}

	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	fmt.Println(data.PI)
	fmt.Println(data.Uate)
	fmt.Printf("% x\n", data.Mine)
	fmt.Println(data.Too)
}

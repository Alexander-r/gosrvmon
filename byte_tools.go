package main

import "log"

func Int64ToBytes(v int64) (bytes []byte) {
	var b byte = 0
	b = byte(v & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 8 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 16 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 24 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 32 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 40 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 48 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 56 & 0xFF)
	bytes = append(bytes, b)
	return bytes
}

func BytesToInt64(data []byte) (number int64) {
	number = 0
	if len(data) < 8 {
		return number
	} else {
		number |= int64(data[0])
		number |= int64(data[1]) << 8
		number |= int64(data[2]) << 16
		number |= int64(data[3]) << 24
		number |= int64(data[4]) << 32
		number |= int64(data[5]) << 40
		number |= int64(data[6]) << 48
		number |= int64(data[7]) << 56
		return number
	}
}

func Int64ToBytesReverse(v int64) (bytes []byte) {
	var b byte = 0
	b = byte(v >> 56 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 48 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 40 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 32 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 24 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 16 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v >> 8 & 0xFF)
	bytes = append(bytes, b)
	b = byte(v & 0xFF)
	bytes = append(bytes, b)
	return bytes
}

func BytesToInt64Reverse(data []byte) (number int64) {
	number = 0
	if len(data) < 8 {
		return number
	} else {
		number |= int64(data[0]) << 56
		number |= int64(data[1]) << 48
		number |= int64(data[2]) << 40
		number |= int64(data[3]) << 32
		number |= int64(data[4]) << 24
		number |= int64(data[5]) << 16
		number |= int64(data[6]) << 8
		number |= int64(data[7])
		return number
	}
}

var I64ToB func(v int64) (bytes []byte) = Int64ToBytesReverse
var BToI64 func(data []byte) (number int64) = BytesToInt64Reverse

func CheckByteOrder() {
	var BOM []byte = []byte{254, 255}
	var i uint16 = 65279
	var bytes []byte
	var b byte = 0
	b = byte(i & 0xFF)
	bytes = append(bytes, b)
	b = byte(i >> 8 & 0xFF)
	bytes = append(bytes, b)
	if BOM[0] == bytes[0] && BOM[1] == bytes[1] {
		I64ToB = Int64ToBytes
		BToI64 = BytesToInt64
		return
	}
	if BOM[0] == bytes[1] && BOM[1] == bytes[0] {
		I64ToB = Int64ToBytesReverse
		BToI64 = BytesToInt64Reverse
		return
	}
	log.Println("Unknown byte order")
}

package tarantooled

type request struct {
	requestID   reqID
	body        string
	requestCode byte
	funcCode    uint32
	packedBody  *[]byte
	callBack    *CallBackSQL
	response    *Response
}

func (r *request) packBody() {
	var packet []byte
	packet = append(packet, []byte{
		0xce, 0, 0, 0, 0, // length
		0x82,                   // 2 element map
		0, byte(r.requestCode), // request code
		1,
	}...)
	packet = append(packet, encode4bytes(0xce, uint32(r.requestID))...)
	packet = append(packet, encode4bytes(0xdf, uint32(2))...) //map size 2
	packet = append(packet, encode4bytes(0xce, 0x22)...)      //func
	//packet = append(packet, encodeString("box.execute")...)
	packet = append(packet, encodeString("GetSql")...)
	packet = append(packet, encode4bytes(0xce, 0x21)...)

	packet = append(packet, encode4bytes(0xdd, uint32(1))...)
	// some := body[0x21].([]interface{})
	// for i := 0; i < 2; i++ {
	packet = append(packet, encodeString(r.body)...)
	// }

	l := uint32(len(packet) - 5)
	(packet)[1] = byte(l >> 24)
	(packet)[2] = byte(l >> 16)
	(packet)[3] = byte(l >> 8)
	(packet)[4] = byte(l)
	r.packedBody = &packet
}

// func encodeAuth(v map[uint32]interface{}) []byte {
// 	var result []byte
// 	result = append(result, encode4bytes(0xdf, uint32(2))...)
// 	result = append(result, encode4bytes(0xce, 0x23)...)
// 	result = append(result, encodeString(v[0x23].(string))...)
// 	result = append(result, encode4bytes(0xce, 0x21)...)
// 	result = append(result, encode4bytes(0xdd, uint32(2))...)
// 	some := v[0x21].([]interface{})
// 	for i := 0; i < 2; i++ {
// 		result = append(result, encodeString(some[i].(string))...)
// 	}
// 	return result
// }

func encode4bytes(code byte, v uint32) []byte {
	var result []byte
	return append(result, []byte{
		code,
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}...)
}

func encodeString(v string) []byte {
	var result []byte
	l := len(v)
	result = append(result, encode4bytes(0xdb, uint32(l))...)
	result = append(result, []byte(v)...)
	return result
}

// func encodeInterface(v interface{}) (result []byte) {

// 	switch reflect.ValueOf(v).Type().Kind() {
// 	case 24: //string
// 		return encodeString(v.(string))
// 	case 23: //slice
// 		return encodeSlice(v.([]interface{}))
// 	}
// 	panic("not typed")
// }

// func encodeSlice(v []interface{}) []byte {
// 	var result []byte
// 	l := len(v)
// 	result = append(result, encode4bytes(0xdd, uint32(l))...)
// 	for i := 0; i < l; i++ {
// 		result = append(result, encodeInterface(v[i])...)
// 	}
// 	return result
// }

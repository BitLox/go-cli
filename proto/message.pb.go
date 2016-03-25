package proto

import (
	"encoding/base64"
	"fmt"
)

type SignMessage struct {
	Handle  []byte `protobuf:"bytes,1,req,name=address_handle_extended"`
	Message []byte `protobuf:"bytes,2,opt,name=message_data"`
}

func (m *SignMessage) Reset() {
	m = &SignMessage{}
}

func (m *SignMessage) String() string {
	return fmt.Sprintf("Message to sign: %s", string(m.Message))
}

func (m *SignMessage) ProtoMessage() {}

type SignatureComplete struct {
	Signature []byte `protobuf:"bytes,1,req,name=signature_data_complete"`
}

func (m *SignatureComplete) Reset() {
	m = &SignatureComplete{}
}

func (m *SignatureComplete) String() string {
	return string(m.Base64Bytes())
}

func (m *SignatureComplete) Base64Bytes() []byte {
	b64Len := base64.StdEncoding.EncodedLen(len(m.Signature))
	b64 := make([]byte, b64Len)
	base64.StdEncoding.Encode(b64, m.Signature)
	return b64
}

func (m *SignatureComplete) ProtoMessage() {}

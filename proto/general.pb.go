package proto

import (
	"fmt"
)

// Success response
type Success struct {
}

func (m *Success) Reset() {
	m = &Success{}
}

func (m *Success) String() string {
	return "Success"
}

func (m *Success) ProtoMessage() {}

// Failure response
type Failure struct {
	Code    int32  `protobuf:"varint,1,req,name=error_code"`
	Message []byte `protobuf:"bytes,2,req,name=error_message"`
}

func (m *Failure) Reset() {
	m = &Failure{}
}

func (m *Failure) Error() string {
	return m.String()
}

func (m *Failure) String() string {
	return fmt.Sprintf("Error [%d] %s", m.Code, m.Message)
}

func (m *Failure) ProtoMessage() {}

// address handle
type AddressHandleExtended struct {
	Root  uint32 `protobuf:"varint,1,opt,name=address_handle_root,def=0"`
	Chain uint32 `protobuf:"varint,2,opt,name=address_handle_chain,def=0"`
	Index uint32 `protobuf:"varint,3,opt,name=address_handle_index,def=0"`
}

func (m *AddressHandleExtended) Reset() {
	m = &AddressHandleExtended{}
}

func (m *AddressHandleExtended) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *AddressHandleExtended) ProtoMessage() {}

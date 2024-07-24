package protobuf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"math"
	"nemo/nemo/conf"
	"reflect"

	"nemo/nemo/network"
	"nemo/sys/log"
	"nemo/sys/util"
)

// -------------------------
// | id | protobuf message |
// -------------------------
type Processor struct {
	littleEndian bool
	msgInfo      map[uint16]*MsgInfo
	msgID        map[reflect.Type]uint16
}

type MsgInfo struct {
	msgType       reflect.Type
	msgHandler    network.MsgHandler
	msgRawHandler network.MsgHandler
}

type MsgRaw struct {
	msgID      uint16
	msgRawData []byte
}

func NewProcessor() *Processor {
	p := new(Processor)
	p.littleEndian = conf.GetSYS().LittleEndian
	p.msgInfo = make(map[uint16]*MsgInfo)
	p.msgID = make(map[reflect.Type]uint16)
	return p
}

// GetByteOrder It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) GetByteOrder() bool {
	return p.littleEndian
}

// Register It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) Register(msg any) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatalf("protobuf message pointer required")
	}
	if _, ok := p.msgID[msgType]; ok {
		log.Fatalf("message %s is already registered", msgType)
	}
	if len(p.msgInfo) >= math.MaxUint16 {
		log.Fatalf("too many protobuf messages (max = %v)", math.MaxUint16)
	}

	id := util.StringHash(fmt.Sprintf("%v", msgType))
	i := new(MsgInfo)
	i.msgType = msgType
	p.msgInfo[id] = i
	p.msgID[msgType] = id
}

// SetHandler It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetHandler(msg any, msgHandler network.MsgHandler) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		log.Fatalf("message %s not registered", msgType)
	}
	if p.msgInfo[id] != nil {
		p.msgInfo[id].msgHandler = msgHandler
	}
}

// SetRawHandler It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) SetRawHandler(id uint16, msgRawHandler network.MsgHandler) {
	if p.msgInfo[id] != nil {
		log.Fatalf("message %d is already registered", id)
	}
	p.msgInfo[id] = new(MsgInfo)
	p.msgInfo[id].msgRawHandler = msgRawHandler
}

// Route goroutine safe
func (p *Processor) Route(agent network.Agent, msg any, userData any) error {
	// raw
	if msgRaw, ok := msg.(MsgRaw); ok {
		i, ok := p.msgInfo[msgRaw.msgID]
		if !ok {
			return fmt.Errorf("message %v not registered", msgRaw.msgID)
		}
		if i.msgRawHandler != nil {
			i.msgRawHandler(agent, []any{msgRaw.msgID, msgRaw.msgRawData, userData})
		}
		return nil
	}

	// protobuf
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		return fmt.Errorf("message %s not registered", msgType)
	}
	i := p.msgInfo[id]
	if i.msgHandler != nil {
		i.msgHandler(agent, []any{msg, userData})
	}

	return nil
}

// Unmarshal goroutine safe
func (p *Processor) Unmarshal(data []byte) (any, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	// id
	var id uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	// msg
	i := p.msgInfo[id]
	if i == nil {
		return nil, errors.New(fmt.Sprintf("message didn't define; %v", id))
	}
	if i.msgRawHandler != nil {
		return MsgRaw{id, data[2:]}, nil
	} else {
		msg := reflect.New(i.msgType.Elem()).Interface()
		return msg, proto.Unmarshal(data[2:], msg.(proto.Message))
	}
}

// Marshal goroutine safe
func (p *Processor) Marshal(msg any) ([][]byte, error) {
	msgType := reflect.TypeOf(msg)

	// id
	_id, ok := p.msgID[msgType]
	if !ok {
		err := fmt.Errorf("message %s not registered", msgType)
		return nil, err
	}

	id := make([]byte, 2)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(id, _id)
	} else {
		binary.BigEndian.PutUint16(id, _id)
	}

	// data
	data, err := proto.Marshal(msg.(proto.Message))
	return [][]byte{id, data}, err
}

// Range goroutine safe
func (p *Processor) Range(f func(id uint16, t reflect.Type)) {
	for id, i := range p.msgInfo {
		f(id, i.msgType)
	}
}

func (p *Processor) GetMsgId(msgtype reflect.Type) uint16 {
	return p.msgID[msgtype]
}

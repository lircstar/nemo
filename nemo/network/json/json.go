package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"nemo/nemo/conf"
	"nemo/nemo/network"
	"nemo/sys/log"
	"nemo/sys/util"
	"reflect"
)

type Processor struct {
	msgInfo map[uint16]*MsgInfo
	msgID   map[reflect.Type]uint16
}

type MsgInfo struct {
	msgType       reflect.Type
	msgHandler    network.MsgHandler
	msgRawHandler network.MsgHandler
}

type MsgRaw struct {
	msgID      uint16
	msgRawData json.RawMessage
}

func NewProcessor() *Processor {
	p := new(Processor)
	p.msgInfo = make(map[uint16]*MsgInfo)
	p.msgID = make(map[reflect.Type]uint16)
	return p
}

// Register
// It's dangerous to call the method on routing or marshaling (unmarshaling)
func (p *Processor) Register(msg any) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		log.Fatal("json message pointer required")
	}

	if _, ok := p.msgID[msgType]; ok {
		log.Fatalf("message %s is already registered", msgType)
	}
	if len(p.msgInfo) >= math.MaxUint16 {
		log.Fatalf("too many protobuf messages (max = %v)", math.MaxUint16)
	}

	name := msgType.Elem().Name()
	if name == "" {
		log.Fatalf("unnamed json message %v ", name)
	}

	msgId := util.StringHash(fmt.Sprintf("%v", name))

	i := new(MsgInfo)
	i.msgType = msgType
	p.msgInfo[msgId] = i
	p.msgID[msgType] = msgId
}

// SetHandler
// It's dangerous to call the method on routing or marshaling (unmarshalling)
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

// SetRawHandler
// It's dangerous to call the method on routing or marshaling (unmarshaling)
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

	// json
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		return errors.New("json message pointer required")
	}

	name := msgType.Elem().Name()
	if name == "" {
		return fmt.Errorf("unnamed json message %v ", name)
	}
	msgId := util.StringHash(fmt.Sprintf("%v", name))

	i, ok := p.msgInfo[msgId]
	if !ok {
		return fmt.Errorf("message %v not registered", msgId)
	}
	if i.msgHandler != nil {
		i.msgHandler(agent, []any{msg, userData})
	}

	return nil
}

// Unmarshal goroutine safe
func (p *Processor) Unmarshal(data []byte) (any, error) {
	var m map[uint16]json.RawMessage
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	if len(m) != 1 {
		return nil, errors.New("invalid json data")
	}

	for msgID, data := range m {
		i, ok := p.msgInfo[msgID]
		if !ok {
			return nil, fmt.Errorf("message %v not registered", msgID)
		}

		// msg
		if i.msgRawHandler != nil {
			return MsgRaw{msgID, data}, nil
		} else {
			msg := reflect.New(i.msgType.Elem()).Interface()
			return msg, json.Unmarshal(data, msg)
		}
	}

	panic("json unmashal bug")
}

// Marshal goroutine safe
func (p *Processor) Marshal(msg any) ([][]byte, error) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		return nil, errors.New("json message pointer required")
	}
	name := msgType.Elem().Name()
	if name == "" {
		return nil, fmt.Errorf("unnamed json message %v ", name)
	}

	msgId := util.StringHash(fmt.Sprintf("%v", name))
	if _, ok := p.msgInfo[msgId]; !ok {
		return nil, fmt.Errorf("message %v not registered", msgId)
	}

	// data
	m := map[uint16]any{msgId: msg}
	data, err := json.Marshal(m)

	return [][]byte{data}, err
}

func (p *Processor) GetByteOrder() bool {
	return conf.GetSYS().LittleEndian
}

// Range goroutine safe
func (p *Processor) Range(f func(id uint16, t reflect.Type)) {
	for id, i := range p.msgInfo {
		f(uint16(id), i.msgType)
	}
}

func (p *Processor) GetMsgId(msgType reflect.Type) uint16 {
	return p.msgID[msgType]
}

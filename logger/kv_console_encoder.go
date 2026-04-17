package cherryLogger

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var _bufferPool = buffer.NewPool()

type kvConsoleEncoder struct {
	cfg    zapcore.EncoderConfig
	fields []zapcore.Field
}

func newKVConsoleEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	if cfg.ConsoleSeparator == "" {
		cfg.ConsoleSeparator = "\t"
	}
	return &kvConsoleEncoder{cfg: cfg}
}

func (c *kvConsoleEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	c.fields = append(c.fields, zap.Array(key, marshaler))
	return nil
}

func (c *kvConsoleEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	c.fields = append(c.fields, zap.Object(key, marshaler))
	return nil
}

func (c *kvConsoleEncoder) AddBinary(key string, value []byte) {
	c.fields = append(c.fields, zap.Binary(key, value))
}

func (c *kvConsoleEncoder) AddByteString(key string, value []byte) {
	c.fields = append(c.fields, zap.ByteString(key, value))
}

func (c *kvConsoleEncoder) AddBool(key string, value bool) {
	c.fields = append(c.fields, zap.Bool(key, value))
}

func (c *kvConsoleEncoder) AddComplex128(key string, value complex128) {
	c.fields = append(c.fields, zap.Complex128(key, value))
}

func (c *kvConsoleEncoder) AddComplex64(key string, value complex64) {
	c.fields = append(c.fields, zap.Complex64(key, value))
}

func (c *kvConsoleEncoder) AddDuration(key string, value time.Duration) {
	c.fields = append(c.fields, zap.Duration(key, value))
}

func (c *kvConsoleEncoder) AddFloat64(key string, value float64) {
	c.fields = append(c.fields, zap.Float64(key, value))
}

func (c *kvConsoleEncoder) AddFloat32(key string, value float32) {
	c.fields = append(c.fields, zap.Float32(key, value))
}

func (c *kvConsoleEncoder) AddInt(key string, value int) {
	c.fields = append(c.fields, zap.Int(key, value))
}

func (c *kvConsoleEncoder) AddInt64(key string, value int64) {
	c.fields = append(c.fields, zap.Int64(key, value))
}

func (c *kvConsoleEncoder) AddInt32(key string, value int32) {
	c.fields = append(c.fields, zap.Int32(key, value))
}

func (c *kvConsoleEncoder) AddInt16(key string, value int16) {
	c.fields = append(c.fields, zap.Int16(key, value))
}

func (c *kvConsoleEncoder) AddInt8(key string, value int8) {
	c.fields = append(c.fields, zap.Int8(key, value))
}

func (c *kvConsoleEncoder) AddString(key, value string) {
	c.fields = append(c.fields, zap.String(key, value))
}

func (c *kvConsoleEncoder) AddTime(key string, value time.Time) {
	c.fields = append(c.fields, zap.Time(key, value))
}

func (c *kvConsoleEncoder) AddUint(key string, value uint) {
	c.fields = append(c.fields, zap.Uint(key, value))
}

func (c *kvConsoleEncoder) AddUint64(key string, value uint64) {
	c.fields = append(c.fields, zap.Uint64(key, value))
}

func (c *kvConsoleEncoder) AddUint32(key string, value uint32) {
	c.fields = append(c.fields, zap.Uint32(key, value))
}

func (c *kvConsoleEncoder) AddUint16(key string, value uint16) {
	c.fields = append(c.fields, zap.Uint16(key, value))
}

func (c *kvConsoleEncoder) AddUint8(key string, value uint8) {
	c.fields = append(c.fields, zap.Uint8(key, value))
}

func (c *kvConsoleEncoder) AddUintptr(key string, value uintptr) {
	c.fields = append(c.fields, zap.Uintptr(key, value))
}

func (c *kvConsoleEncoder) AddReflected(key string, value interface{}) error {
	c.fields = append(c.fields, zap.Reflect(key, value))
	return nil
}

func (c *kvConsoleEncoder) OpenNamespace(key string) {}

func (c *kvConsoleEncoder) Clone() zapcore.Encoder {
	return &kvConsoleEncoder{
		cfg:    c.cfg,
		fields: append([]zapcore.Field{}, c.fields...),
	}
}

func (c *kvConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	line := _bufferPool.Get()

	arr := getSliceEncoder()

	if c.cfg.TimeKey != "" && c.cfg.EncodeTime != nil && !entry.Time.IsZero() {
		c.cfg.EncodeTime(entry.Time, arr)
	}
	if c.cfg.LevelKey != "" && c.cfg.EncodeLevel != nil {
		c.cfg.EncodeLevel(entry.Level, arr)
	}
	if entry.LoggerName != "" && c.cfg.NameKey != "" && c.cfg.EncodeName != nil {
		c.cfg.EncodeName(entry.LoggerName, arr)
	}
	if entry.Caller.Defined && c.cfg.CallerKey != "" && c.cfg.EncodeCaller != nil {
		c.cfg.EncodeCaller(entry.Caller, arr)
	}
	if entry.Caller.Defined && c.cfg.FunctionKey != "" {
		arr.AppendString(entry.Caller.Function)
	}

	for i := range arr.elems {
		if i > 0 {
			line.AppendString(c.cfg.ConsoleSeparator)
		}
		fmt.Fprint(line, arr.elems[i])
	}
	putSliceEncoder(arr)

	if len(c.fields) > 0 {
		if line.Len() > 0 {
			line.AppendString(c.cfg.ConsoleSeparator)
		}
		writeKVFields(line, c.fields)
	}

	if c.cfg.MessageKey != "" {
		if line.Len() > 0 {
			line.AppendString(c.cfg.ConsoleSeparator)
		}
		line.AppendString(entry.Message)
	}

	if len(fields) > 0 {
		if line.Len() > 0 {
			line.AppendString(c.cfg.ConsoleSeparator)
		}
		writeKVFields(line, fields)
	}

	if entry.Stack != "" && c.cfg.StacktraceKey != "" {
		line.AppendByte('\n')
		line.AppendString(entry.Stack)
	}

	if c.cfg.LineEnding != "" {
		line.AppendString(c.cfg.LineEnding)
	} else {
		line.AppendString(zapcore.DefaultLineEnding)
	}

	return line, nil
}

func writeKVFields(buf *buffer.Buffer, fields []zapcore.Field) {
	for i, f := range fields {
		if i > 0 {
			buf.AppendByte(' ')
		}
		buf.AppendString(f.Key)
		buf.AppendByte('=')
		switch f.Type {
		case zapcore.StringType:
			buf.AppendString(f.String)
		case zapcore.Int64Type:
			buf.AppendInt(f.Integer)
		case zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
			buf.AppendInt(int64(f.Integer))
		case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
			buf.AppendUint(uint64(f.Integer))
		case zapcore.BoolType:
			if f.Integer == 1 {
				buf.AppendString("true")
			} else {
				buf.AppendString("false")
			}
		case zapcore.Float64Type:
			buf.AppendString(strconv.FormatFloat(float64(f.Integer), 'f', -1, 64))
		case zapcore.Float32Type:
			buf.AppendString(strconv.FormatFloat(float64(f.Integer), 'f', -1, 32))
		case zapcore.DurationType:
			buf.AppendInt(f.Integer)
		case zapcore.TimeType:
			buf.AppendString(time.Unix(0, f.Integer).Format(time.RFC3339))
		case zapcore.ErrorType:
			buf.AppendString(f.Interface.(error).Error())
		case zapcore.ReflectType:
			buf.AppendString(fmt.Sprintf("%v", f.Interface))
		default:
			buf.AppendString(fmt.Sprintf("%v", f.Interface))
		}
	}
}

type sliceArrayEncoder struct {
	elems []interface{}
}

var _sliceEncoderPool = make(chan *sliceArrayEncoder, 10)

func init() {
	for i := 0; i < 10; i++ {
		_sliceEncoderPool <- &sliceArrayEncoder{elems: make([]interface{}, 0, 2)}
	}
}

func getSliceEncoder() *sliceArrayEncoder {
	select {
	case e := <-_sliceEncoderPool:
		return e
	default:
		return &sliceArrayEncoder{elems: make([]interface{}, 0, 2)}
	}
}

func putSliceEncoder(e *sliceArrayEncoder) {
	e.elems = e.elems[:0]
	select {
	case _sliceEncoderPool <- e:
	default:
	}
}

func (s *sliceArrayEncoder) AppendArray(v zapcore.ArrayMarshaler) error {
	enc := &sliceArrayEncoder{elems: make([]interface{}, 0)}
	err := v.MarshalLogArray(enc)
	s.elems = append(s.elems, enc.elems)
	return err
}

func (s *sliceArrayEncoder) AppendObject(v zapcore.ObjectMarshaler) error {
	m := make(map[string]interface{})
	enc := &mapObjectEncoder{Fields: m}
	err := v.MarshalLogObject(enc)
	s.elems = append(s.elems, m)
	return err
}

func (s *sliceArrayEncoder) AppendReflected(v interface{}) error {
	s.elems = append(s.elems, v)
	return nil
}

func (s *sliceArrayEncoder) AppendBool(v bool)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendByteString(v []byte)      { s.elems = append(s.elems, string(v)) }
func (s *sliceArrayEncoder) AppendComplex128(v complex128)  { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendComplex64(v complex64)    { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendDuration(v time.Duration) { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendFloat64(v float64)        { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendFloat32(v float32)        { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt(v int)                { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt64(v int64)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt32(v int32)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt16(v int16)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendInt8(v int8)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendString(v string)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendTime(v time.Time)         { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint(v uint)              { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint64(v uint64)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint32(v uint32)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint16(v uint16)          { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUint8(v uint8)            { s.elems = append(s.elems, v) }
func (s *sliceArrayEncoder) AppendUintptr(v uintptr)        { s.elems = append(s.elems, v) }

type mapObjectEncoder struct {
	Fields map[string]interface{}
}

func (m *mapObjectEncoder) AddArray(key string, v zapcore.ArrayMarshaler) error {
	arr := &sliceArrayEncoder{elems: make([]interface{}, 0)}
	err := v.MarshalLogArray(arr)
	m.Fields[key] = arr.elems
	return err
}

func (m *mapObjectEncoder) AddObject(key string, v zapcore.ObjectMarshaler) error {
	newMap := make(map[string]interface{})
	enc := &mapObjectEncoder{Fields: newMap}
	err := v.MarshalLogObject(enc)
	m.Fields[key] = newMap
	return err
}

func (m *mapObjectEncoder) AddBinary(key string, v []byte)          { m.Fields[key] = v }
func (m *mapObjectEncoder) AddByteString(key string, v []byte)      { m.Fields[key] = string(v) }
func (m *mapObjectEncoder) AddBool(key string, v bool)              { m.Fields[key] = v }
func (m *mapObjectEncoder) AddDuration(key string, v time.Duration) { m.Fields[key] = v }
func (m *mapObjectEncoder) AddComplex128(key string, v complex128)  { m.Fields[key] = v }
func (m *mapObjectEncoder) AddComplex64(key string, v complex64)    { m.Fields[key] = v }
func (m *mapObjectEncoder) AddFloat64(key string, v float64)        { m.Fields[key] = v }
func (m *mapObjectEncoder) AddFloat32(key string, v float32)        { m.Fields[key] = v }
func (m *mapObjectEncoder) AddInt(key string, v int)                { m.Fields[key] = v }
func (m *mapObjectEncoder) AddInt64(key string, v int64)            { m.Fields[key] = v }
func (m *mapObjectEncoder) AddInt32(key string, v int32)            { m.Fields[key] = v }
func (m *mapObjectEncoder) AddInt16(key string, v int16)            { m.Fields[key] = v }
func (m *mapObjectEncoder) AddInt8(key string, v int8)              { m.Fields[key] = v }
func (m *mapObjectEncoder) AddString(key string, v string)          { m.Fields[key] = v }
func (m *mapObjectEncoder) AddTime(key string, v time.Time)         { m.Fields[key] = v }
func (m *mapObjectEncoder) AddUint(key string, v uint)              { m.Fields[key] = v }
func (m *mapObjectEncoder) AddUint64(key string, v uint64)          { m.Fields[key] = v }
func (m *mapObjectEncoder) AddUint32(key string, v uint32)          { m.Fields[key] = v }
func (m *mapObjectEncoder) AddUint16(key string, v uint16)          { m.Fields[key] = v }
func (m *mapObjectEncoder) AddUint8(key string, v uint8)            { m.Fields[key] = v }
func (m *mapObjectEncoder) AddUintptr(key string, v uintptr)        { m.Fields[key] = v }
func (m *mapObjectEncoder) AddReflected(key string, v interface{}) error {
	m.Fields[key] = v
	return nil
}
func (m *mapObjectEncoder) OpenNamespace(key string) {
	ns := make(map[string]interface{})
	m.Fields[key] = ns
	m.Fields = ns
}

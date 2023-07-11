package zaplogger

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const _hex = "0123456789abcdef"

var _sliceEncoderPool = sync.Pool{
	New: func() interface{} {
		return &sliceArrayEncoder{elems: make([]interface{}, 0, 2)}
	},
}

func getSliceEncoder() *sliceArrayEncoder {
	return _sliceEncoderPool.Get().(*sliceArrayEncoder)
}

func putSliceEncoder(e *sliceArrayEncoder) {
	e.elems = e.elems[:0]
	_sliceEncoderPool.Put(e)
}

var bufferPool = buffer.NewPool()

var _textPool = sync.Pool{
	New: func() interface{} {
		return &textEncoder{}
	},
}

func getTextEncoder() *textEncoder {
	return _textPool.Get().(*textEncoder)
}

func putTextEncoder(enc *textEncoder) {
	enc.EncoderConfig = nil
	enc.buf = nil
	_textPool.Put(enc)
}

type textEncoder struct {
	*zapcore.EncoderConfig

	buf       *buffer.Buffer
	separator string
}

// NewTextEncoder creates a key=value encoder.
func NewTextEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &textEncoder{
		EncoderConfig: &cfg,
		buf:           bufferPool.Get(),
		separator:     " ",
	}
}

func (enc *textEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *textEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *textEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *textEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *textEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *textEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

func (enc *textEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *textEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

func (enc *textEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

func (enc *textEncoder) AddReflected(key string, obj interface{}) error {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(marshaled)
	return err
}

func (enc *textEncoder) OpenNamespace(key string) {
}

func (enc *textEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
}

func (enc *textEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
}

func (enc *textEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
}

func (enc *textEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.addElementSeparator()
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	return err
}

func (enc *textEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	enc.addElementSeparator()
	enc.buf.AppendByte('{')
	err := obj.MarshalLogObject(enc)
	enc.buf.AppendByte('}')
	return err
}

func (enc *textEncoder) AppendBool(val bool) {
	enc.addElementSeparator()
	enc.buf.AppendBool(val)
}

func (enc *textEncoder) AppendByteString(val []byte) {
	enc.addElementSeparator()
	enc.buf.AppendByte('"')
	enc.safeAddByteString(val)
	enc.buf.AppendByte('"')
}

func (enc *textEncoder) AppendComplex128(val complex128) {
	enc.addElementSeparator()
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendByte('"')
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, 64)
	enc.buf.AppendByte('+')
	enc.buf.AppendFloat(i, 64)
	enc.buf.AppendByte('i')
	enc.buf.AppendByte('"')
}

func (enc *textEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	if e := enc.EncodeDuration; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendInt64(int64(val))
	}
}

func (enc *textEncoder) AppendInt64(val int64) {
	enc.addElementSeparator()
	enc.buf.AppendInt(val)
}

func (enc *textEncoder) AppendReflected(val interface{}) error {
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}
	enc.addElementSeparator()
	_, err = enc.buf.Write(marshaled)
	return err
}

func (enc *textEncoder) AppendString(val string) {
	enc.addElementSeparator()
	enc.buf.AppendString(val) // replace enc.safeAddString(val)
}

func (enc *textEncoder) AppendTimeLayout(val time.Time, layout string) {
	enc.addElementSeparator()
	enc.buf.AppendTime(val, layout)
}

func (enc *textEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	if e := enc.EncodeTime; e != nil {
		e(val, enc)
	}
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

func (enc *textEncoder) AppendUint64(val uint64) {
	enc.addElementSeparator()
	enc.buf.AppendUint(val)
}

func (enc *textEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }
func (enc *textEncoder) AddFloat32(k string, v float32)     { enc.AddFloat64(k, float64(v)) }
func (enc *textEncoder) AddInt(k string, v int)             { enc.AddInt64(k, int64(v)) }
func (enc *textEncoder) AddInt32(k string, v int32)         { enc.AddInt64(k, int64(v)) }
func (enc *textEncoder) AddInt16(k string, v int16)         { enc.AddInt64(k, int64(v)) }
func (enc *textEncoder) AddInt8(k string, v int8)           { enc.AddInt64(k, int64(v)) }
func (enc *textEncoder) AddUint(k string, v uint)           { enc.AddUint64(k, uint64(v)) }
func (enc *textEncoder) AddUint32(k string, v uint32)       { enc.AddUint64(k, uint64(v)) }
func (enc *textEncoder) AddUint16(k string, v uint16)       { enc.AddUint64(k, uint64(v)) }
func (enc *textEncoder) AddUint8(k string, v uint8)         { enc.AddUint64(k, uint64(v)) }
func (enc *textEncoder) AddUintptr(k string, v uintptr)     { enc.AddUint64(k, uint64(v)) }
func (enc *textEncoder) AppendComplex64(v complex64)        { enc.AppendComplex128(complex128(v)) }
func (enc *textEncoder) AppendFloat64(v float64)            { enc.appendFloat(v, 64) }
func (enc *textEncoder) AppendFloat32(v float32)            { enc.appendFloat(float64(v), 32) }
func (enc *textEncoder) AppendInt(v int)                    { enc.AppendInt64(int64(v)) }
func (enc *textEncoder) AppendInt32(v int32)                { enc.AppendInt64(int64(v)) }
func (enc *textEncoder) AppendInt16(v int16)                { enc.AppendInt64(int64(v)) }
func (enc *textEncoder) AppendInt8(v int8)                  { enc.AppendInt64(int64(v)) }
func (enc *textEncoder) AppendUint(v uint)                  { enc.AppendUint64(uint64(v)) }
func (enc *textEncoder) AppendUint32(v uint32)              { enc.AppendUint64(uint64(v)) }
func (enc *textEncoder) AppendUint16(v uint16)              { enc.AppendUint64(uint64(v)) }
func (enc *textEncoder) AppendUint8(v uint8)                { enc.AppendUint64(uint64(v)) }
func (enc *textEncoder) AppendUintptr(v uintptr)            { enc.AppendUint64(uint64(v)) }

func (enc *textEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *textEncoder) clone() *textEncoder {
	clone := getTextEncoder()
	clone.EncoderConfig = enc.EncoderConfig
	clone.buf = bufferPool.Get()
	clone.separator = enc.separator
	return clone
}

func (enc *textEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()

	arr := getSliceEncoder()
	if enc.TimeKey != "" && enc.EncodeTime != nil {
		enc.EncodeTime(ent.Time, arr)
	}
	if enc.LevelKey != "" && enc.EncodeLevel != nil {
		enc.EncodeLevel(ent.Level, arr)
	}
	if ent.LoggerName != "" && enc.NameKey != "" {
		nameEncoder := enc.EncodeName

		// if no name encoder provided, fall back to FullNameEncoder for backwards
		// compatibility
		if nameEncoder == nil {
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(ent.LoggerName, arr)
	}
	if ent.Caller.Defined {
		if enc.CallerKey != "" && enc.EncodeCaller != nil {
			enc.EncodeCaller(ent.Caller, arr)
		}
		if enc.FunctionKey != "" {
			arr.AppendString(ent.Caller.Function)
		}
	}
	if enc.buf.Len() > 0 {
		arr.AppendByteString(enc.buf.Bytes())
	}
	if final.MessageKey != "" {
		arr.AppendString(ent.Message)
	}
	for i := range arr.elems {
		if i > 0 {
			final.buf.AppendString(enc.separator)
		}
		fmt.Fprint(final.buf, arr.elems[i])

		// Align level
		if i == 1 {
			if ent.Level == zapcore.InfoLevel || ent.Level == zapcore.WarnLevel {
				final.buf.AppendByte(' ')
			}
		}
	}
	putSliceEncoder(arr)

	addFields(final, fields)

	// If there's no stacktrace key, honor that; this allows users to force
	// single-line output.
	if ent.Stack != "" && final.StacktraceKey != "" {
		final.buf.AppendByte('\n')
		final.buf.AppendString(ent.Stack)
	}
	if final.LineEnding != "" {
		final.buf.AppendString(final.LineEnding)
	} else {
		final.buf.AppendString(zapcore.DefaultLineEnding)
	}

	ret := final.buf
	putTextEncoder(final)
	return ret, nil
}

func (enc *textEncoder) truncate() {
	enc.buf.Reset()
}

func (enc *textEncoder) addKey(key string) {
	if enc.buf.Len() > 0 {
		enc.buf.AppendString(enc.separator)
	}
	enc.safeAddString(key)
	enc.buf.AppendByte('=')
}

func (enc *textEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	switch enc.buf.Bytes()[last] {
	case '{', '[', '=', ',':
		return
	default:
		enc.buf.AppendByte(',')
	}
}

func (enc *textEncoder) appendFloat(val float64, bitSize int) {
	enc.addElementSeparator()
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`NaN`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`+Inf`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`-Inf`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *textEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

// safeAddByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (enc *textEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.Write(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (enc *textEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte(b)
	case '\n':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('n')
	case '\r':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('r')
	case '\t':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (enc *textEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

package tnetstr

import (
	"bytes"
	"encoding/base64"

	"errors"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func Marshal(v interface{}) ([]byte, error) {
	e := new(encodeState)
	s, err := e.marshal(v)
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

func Unmarshal(data string, v interface{}) error {
	val, _, err := parse(data)
	if err != nil {
		return err
	}
	switch a := v.(type) {
	case *interface{}:
		*a = val
		return nil
	}
	return nil
}
func parse(data string) (interface{}, string, error) {
	payload, payloadType, remain := parsePayload(data)

	switch payloadType {
	case '#':
		value, err := strconv.ParseInt(payload, 10, 64)
		return value, remain, err
	case '}':
		value, err := parseDict(payload)
		return value, remain, err
	case ']':
		value, err := parseList(payload)
		return value, remain, err
	case '!':
		value := (payload == "true")
		return value, remain, nil
	case '^':
		value, err := strconv.ParseFloat(payload, 64)
		return value, remain, err

	case '~':
		var err error = nil
		if len(payload) != 0 {
			err = errors.New("Payload must be 0 length for null.")
		}
		return interface{}(nil), remain, err
	case ',':
		return payload, remain, nil
	}
	panic("Invalid payload type: " + string(payloadType))
}

func parsePayload(data string) (string, byte, string) {
	lenStr := strings.SplitN(data, ":", 2)
	extra := data[len(lenStr[0])+1:]
	//fmt.Println(extra)
	length, err := strconv.ParseInt(lenStr[0], 10, 64)
	if err != nil {
		panic("length:" + err.Error())
	}

	payload, extra := extra[0:length], extra[length:]
	payloadType, remain := extra[0], extra[1:]

	return payload, payloadType, remain
}

func parseList(data string) ([]interface{}, error) {
	if data == "" {
		return []interface{}{}, nil
	}

	value, extra, err := parse(data)
	if err != nil {
		return nil, err
	}
	result := append([]interface{}(nil), value)

	for extra != "" {
		value, extra, err = parse(extra)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func parsePair(data string) (interface{}, interface{}, string, error) {
	key, extra, err := parse(data)
	if err != nil {
		return nil, nil, "", err
	}
	if extra == "" {
		panic("Unbalanced dictionary store.")
	}
	value, extra, err := parse(extra)
	if err != nil {
		return nil, nil, "", err
	}
	return key, value, extra, nil
}

func parseDict(data string) (map[string]interface{}, error) {
	if data == "" {
		return map[string]interface{}{}, nil
	}

	key, value, extra, err := parsePair(data)
	if err != nil {
		return nil, err
	}
	k, ok := key.(string)
	if !ok {
		panic("Keys can only be strings.")
	}

	result := map[string]interface{}{k: value}

	for extra != "" {
		key, value, extra, err = parsePair(extra)
		if err != nil {
			return nil, err
		}
		result[key.(string)] = value
	}
	return result, nil
}

/*
func dumpDict(data interface{}){
    result = []
    for k,v in data.items():
        result.append(dump(str(k)))
        result.append(dump(v))

    payload = ''.join(result)
    return '%d:' % len(payload) + payload + '}'
}

func dumpList(data interface{}):
    fmt.Println(data)
    for i in data:
        result.append(dump(i))

    payload = ''.join(result)
    return '%d:' % len(payload) + payload + ']'*/

type encodeState struct {
	bytes.Buffer // accumulated output
}

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "tnetstr: unsupported type: " + e.Type.String()
}

func (e *encodeState) marshal(v interface{}) (s string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	return e.reflectValue(reflect.ValueOf(v)), nil
}

func (e *encodeState) error(err error) {
	panic(err)
}

var byteSliceType = reflect.TypeOf([]byte(nil))

func (e *encodeState) reflectValue(v reflect.Value) string {
	if !v.IsValid() {
		return "0:~"
	}

	switch v.Kind() {
	case reflect.Bool:
		x := v.Bool()
		if x {
			return "4:true!"
		} else {
			return "5:false!"
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := strconv.FormatInt(v.Int(), 10)
		l := strconv.Itoa(len(s))
		return l + ":" + s + "#"

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		s := strconv.FormatUint(v.Uint(), 10)
		l := strconv.Itoa(len(s))
		return l + ":" + s + "#"

	case reflect.Float32, reflect.Float64:
		s := strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits())
		l := strconv.Itoa(len(s))
		return l + ":" + s + "^"

	case reflect.String:
		l := strconv.Itoa(len(v.String()))
		return l + ":" + v.String() + ","

	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			e.error(&UnsupportedTypeError{v.Type()})
		}
		if v.IsNil() || v.Len() == 0 {
			return "0:}"
		}

		var sv stringValues = v.MapKeys()
		sort.Sort(sv)
		for _, k := range sv {
			l := strconv.Itoa(len(k.String()))
			s := l + ":" + k.String() + "," + e.reflectValue(v.MapIndex(k))
			return strconv.Itoa(len(s)) + ":" + s + "}"
		}

	case reflect.Array, reflect.Slice:
		var des string

		if v.Type() == byteSliceType {
			s := v.Interface().([]byte)
			// for small buffers, using Encode directly is much faster.
			dst := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
			base64.StdEncoding.Encode(dst, s)
			des = des + string(dst)
		}

		n := v.Len()
		for i := 0; i < n; i++ {
			des = des + e.reflectValue(v.Index(i))
		}
		l := strconv.Itoa(len(des))
		return l + ":" + des + "]"

	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return "0:~"
		}
		return e.reflectValue(v.Elem())

	default:
		e.error(&UnsupportedTypeError{v.Type()})
	}
	return ""
}

type stringValues []reflect.Value

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }

package tnetstr

import (
	"strconv"
	"strings"
	"os"
)

func dump(data interface{}) {
	/*	data.(int64)
			data.(float64)
			data.(string)
			data.(map[string]interface{})
			data.([]interface{})
			data == nil
			data.(bool)

		    if type(data) is long or type(data) is int:
		        out = str(data)
		        return '%d:%s#' % (len(out), out)
		    elif type(data) is float:
		        out = '%f' % data
		        return '%d:%s^' % (len(out), out)
		    elif type(data) is str:
		        return '%d:' % len(data) + data + ',' 
		    elif type(data) is dict:
		        return dumpDict(data)
		    elif type(data) is list:
		        return dumpList(data)
		    elif data == None:
		        return '0:~'
		    elif type(data) is bool:
		        out = repr(data).lower()
		        return '%d:%s!' % (len(out), out)
		    else:
		        assert False, "Can't serialize stuff that's %s." % type(data)*/
}

func parse(data string) (interface{}, string, os.Error) {
	payload, payloadType, remain := parsePayload(data)

	switch payloadType {
	case '#':
		value, err := strconv.Atoi64(payload)
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
		value, err := strconv.Atof64(payload)
		return value, remain, err

	case '~':
		var err os.Error = nil
		if len(payload) != 0 {
			err = os.NewError("Payload must be 0 length for null.")
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
	length, err := strconv.Atoi64(lenStr[0])
	if err != nil {
		panic("length:" + err.String())
	}

	payload, extra := extra[0:length], extra[length:]
	payloadType, remain := extra[0], extra[1:]

	return payload, payloadType, remain
}

func parseList(data string) ([]interface{}, os.Error) {
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

func parsePair(data string) (interface{}, interface{}, string, os.Error) {
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

func parseDict(data string) (map[string]interface{}, os.Error) {
	if data == "" {
		return nil, nil
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
    result = []
    for i in data:
        result.append(dump(i))

    payload = ''.join(result)
    return '%d:' % len(payload) + payload + ']'*/

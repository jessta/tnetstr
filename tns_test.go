package tnetstr

import (
	"fmt"
	"testing"
	"reflect"
)


var FORMAT_EXAMPLES = map[string]interface{}{
	"0:}": map[string]interface{}{},
	"0:]": []interface{}{},
	"51:5:hello,39:11:12345678901#4:this,4:true!0:~4:\x00\x00\x00\x00,]}": map[string]interface{}{"hello": []interface{}{int64(12345678901), "this", true, nil, "\x00\x00\x00\x00"}},
	"5:12345#":                                                            int64(12345),
	"12:this is cool,":                                                    "this is cool",
	"0:,":                                                                 "",
	"0:~":                                                                 nil,
	"4:true!":                                                             true,
	"5:false!":                                                            false,
	"10:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00,": "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00",
	"24:5:12345#5:67890#5:xxxxx,]":                 []interface{}{int64(12345), int64(67890), "xxxxx"},
	"18:3:0.1^3:0.2^3:0.3^]":                       []interface{}{float64(0.1), float64(0.2), float64(0.3)},
	"4:test,":                                      "test",
	"3:123#":                                       int64(123),
	"16:5:hello,5:12345#}":                         map[string]interface{}{"hello": int64(12345)},
	"32:5:hello,5:12345#5:hello,5:56789#]": []interface{}{"hello", int64(12345), "hello", int64(56789)},
	"9:3.1415926^": float64(3.1415926),
}

func TestParse(t *testing.T) {
	defer func() {
		a := recover()
		if a != nil {
			fmt.Println(a)
			t.Fail()
		}
	}()
	for key, val := range FORMAT_EXAMPLES {

		a, s, err := parse(key)
		if err != nil {
			t.Fail()
		}
		bo := reflect.DeepEqual(a, val)
		if !bo {
			fmt.Println("for: " + key)
			fmt.Println(val)
			fmt.Println(a)
			fmt.Println(s)
			t.Fail()
		}
	}
}

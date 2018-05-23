package main

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	texts := [...]string{
		"this is a plain text",
		`{"this": "is", "a": "text",\n"with": "multiple", "line": "s"}`,
		`version:"2"\ndata:\n\tplain: "yml"`,
		"",
		"text",
		"\t",
		"\n",
		"\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
		"0",
		"0x00000",
	}

	for _, text := range texts {
		reader := strings.NewReader(text)
		if content := read(reader); string(content) != text {
			t.Errorf("error reading, expected %s, got %s", text, string(content))
		}
	}
}

func TestMarshal(t *testing.T) {
	test := map[string]string{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}
	if byt, err := marshal(test, true); err != nil {
		t.Errorf("wrong marshal: %v got %s ", err, string(byt))
	}

	expected := "{\n    \"app\": \"a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==\",\n    \"password\": \"c2VjcmV0\"\n}"
	if byt, _ := marshal(test, true); expected != string(byt) {
		t.Errorf("wrong marshal: expected \n%s\n got \n%s\n", expected, string(byt))
	}

	testYml := map[string]interface{}{
		"data": map[string]string{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
	}

	expected = "data:\n  app: a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==\n  password: c2VjcmV0\n"
	if byt, _ := marshal(testYml, false); expected != string(byt) {
		t.Errorf("wrong marshal: expected \n%s\n got \n%s\n", expected, string(byt))
	}

}

func TestUnmarshalJSON(t *testing.T) {
	var j map[string]interface{}
	jsonCase, _ := ioutil.ReadFile("./mock.json")
	expected := map[string]interface{}{
		"apiVersion": "v1",
		"data": map[string]interface{}{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
		"kind": "Secret",
		"metadata": map[string]interface{}{
			"name":      "kubernetes secret decoder",
			"namespace": "ksd",
		},
		"type": "Opaque",
	}
	if err := unmarshal(jsonCase, &j, true); err != nil {
		t.Errorf("must return a valid struct %v", err)
	}
	if !reflect.DeepEqual(expected, j) {
		t.Errorf("json struct does not match.\nexpected\n%v\ngot\n%v", expected, j)
	}
}

func TestUnmarshalYaml(t *testing.T) {
	var y map[string]interface{}
	yamlCase, _ := ioutil.ReadFile("./mock.yml")
	expected := map[string]interface{}{
		"apiVersion": "v1",
		"data": map[interface{}]interface{}{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
		"kind": "Secret",
		"metadata": map[interface{}]interface{}{
			"name":      "kubernetes secret decoder",
			"namespace": "ksd",
		},
		"type": "Opaque",
	}
	if err := unmarshal(yamlCase, &y, false); err != nil {
		t.Errorf("must return a valid struct %v", err)
	}
	if !reflect.DeepEqual(expected, y) {
		t.Errorf("yaml struct does not match.\nexpected\n%v\ngot\n%v", expected, y)
	}
}

func TestDecode(t *testing.T) {
	s := &secret{}
	if err := decode(s); err != nil {
		t.Errorf("wrong decode: %v got %v", err, s)
	}
	s = &secret{
		Data: map[string]string{
			"password": "c2VjcmV",
		},
	}
	if err := decode(s); err == nil {
		t.Errorf("expected `illegal base64 data` got %v", s)
	}
	s = &secret{
		Data: map[string]string{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
	}
	if err := decode(s); err != nil {
		t.Errorf("wrong decode: %v got %v", err, s)
	}
	expected := &secret{
		Data: map[string]string{
			"password": "secret",
			"app":      "kubernetes secret decoder",
		},
	}
	if !reflect.DeepEqual(expected, s) {
		t.Errorf("wrong decode expected %v got %v", expected, s)
	}
}

func TestIsJSONString(t *testing.T) {
	yamlCase, _ := ioutil.ReadFile("./mock.yml")
	wrongTests := [...][]byte{
		nil,
		[]byte(""),
		[]byte("k"),
		[]byte("-"),
		[]byte(`"test": "case"`),
		yamlCase,
	}
	for _, test := range wrongTests {
		if isJSONString(test) {
			t.Errorf("%v must not be a json string", string(test))
		}
	}
	jsonCase, _ := ioutil.ReadFile("./mock.json")
	successCases := [...][]byte{
		[]byte("null"),
		[]byte(`{"valid":"json"}`),
		[]byte(`{"nested": {"json": "string"}}`),
		jsonCase,
	}
	for _, test := range successCases {
		if !isJSONString(test) {
			t.Errorf("%v must be a json string", string(test))
		}
	}
}

func TestParse(t *testing.T) {
	if s, e := parse(strings.NewReader(`{"a"`)); e == nil {
		t.Errorf("expected invalid parse got %v", s)
	}

	// Return same string without data part
	expected := `{\n    "key": "value"\n}`
	if s, e := parse(strings.NewReader(`{"key": "value"}`)); e != nil {
		t.Errorf("expected %v got %v", expected, s)
	}
	if s, e := parse(strings.NewReader(`{ "data": { "password": "c2VjcmV" } }`)); e == nil {
		t.Errorf("expected illegal base64 data got %v", s)
	}
	if s, e := parse(strings.NewReader(`{"data": {"password": "c2VjcmV0"}}`)); e != nil {
		t.Errorf("wrong parse got %v", s)
	}
}
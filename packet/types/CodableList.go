package types

import (
	"sort"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type CodableList map[int]*CompleteCodableValue

func (l CodableList) Append(c *CompleteCodableValue) {
	l[len(l)] = c
}

func (l CodableList) Add(i int, c *CompleteCodableValue) {
	l[i] = c
}

func (l CodableList) Encode() string {
	var encoded string
	firstArg := len(l)
	encoded += CodableValues.EncodeInt(firstArg)
	keys := make([]int, 0)
for k := range l {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		encoded += l[k].Encode()
	}
	return encoded
}

func (l CodableList) Decode(data string) (string, error) {
	length, rest, err := CodableValues.DecodeInt(data)
	if err != nil {
		return "", err
	}
	for i := 1; i <= length; i++ {
		c := &CompleteCodableValue{}
		rest, err = c.Decode(rest)
		if err != nil {
			return "", err
		}
		l.Add(i, c)
	}
	return rest, nil
}

func (l CodableList) Equals(other interface{}) bool {
	ActualValue := other.(CodableList)
	if len(l) != len(ActualValue) {
		return false
	}
	for k, v := range l {
		if !v.Equals(ActualValue[k]) {
			return false
		}
	}
	return true
}

func (l CodableList) String() string {
	var str string = "["
	for _, v := range l {
		str += v.String()
	}
	return str + "]"
}
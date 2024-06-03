package callstack

import "encoding/json"

type UniqueList struct {
	str2num map[string]int64
	strList []string
}

func NewUniqueList() *UniqueList {
	return &UniqueList{
		str2num: make(map[string]int64),
		strList: make([]string, 0),
	}
}

func (sm *UniqueList) Insert(element string) int64 {
	if i, ok := sm.str2num[element]; ok {
		return i
	} else {
		index := int64(len(sm.strList))
		sm.strList = append(sm.strList, element)
		sm.str2num[element] = index
		return index
	}
}

func (sm *UniqueList) MarshalJSON() ([]byte, error) {
	return json.Marshal(sm.strList)
}

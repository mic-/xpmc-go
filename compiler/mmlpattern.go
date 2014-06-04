package compiler

import "../utils"

type MmlPattern struct {
    Name string
    Cmds []int
    HasAnyNote bool
    NumTicks int
}

type MmlPatternMap struct {
    keys [] string
    data []*MmlPattern
}

func (m *MmlPattern) GetCommands() []int {
    return m.Cmds
}

func (m *MmlPatternMap) FindKey(key string) int {
    return utils.PositionOfString(m.keys, key)
}

func (m *MmlPatternMap) Append(key string, pat *MmlPattern) {
    m.keys = append(m.keys, key)
    m.data = append(m.data, pat)
}

func (m *MmlPatternMap) GetNumTicks(key string) int {
    pos := m.FindKey(key)
    if pos >= 0 {
        return m.data[pos].NumTicks
    }
    return 0
}

func (m *MmlPatternMap) HasAnyNote(key string) bool {
    pos := m.FindKey(key)
    if pos >= 0 {
        return m.data[pos].HasAnyNote
    }
    return false
}

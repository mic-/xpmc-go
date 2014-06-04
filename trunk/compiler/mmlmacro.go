package compiler

import "../utils"

// For macro elements
const (
    ARG_REFERENCE = 1
    DEFAULT_PARAM = 2
    CHAR_VERBATIM = 3
)

type MmlMacroElement struct {
    typ int
    val interface{}
}

type MmlMacro struct {
    data []*MmlMacroElement
}

type MmlMacroMap struct {
    keys []string
    data []*MmlMacro
}

func (m *MmlMacroMap) FindKey(key string) int {
    return utils.PositionOfString(m.keys, key)
}

func (m *MmlMacroMap) Append(key string, mac *MmlMacro) {
    m.keys = append(m.keys, key)
    m.data = append(m.data, mac)
}

func (m *MmlMacroMap) GetData(key string) *MmlMacro {
    pos := m.FindKey(key)
    if pos >= 0 {
        return m.data[pos]
    }
    return nil
}

func (m *MmlMacro) AppendArgumentRef(x int) {
    m.data = append(m.data, &MmlMacroElement{ARG_REFERENCE, x})
}

func (m *MmlMacro) AppendChar(x byte) {
    m.data = append(m.data, &MmlMacroElement{CHAR_VERBATIM, x})
}

func (m *MmlMacro) AppendDefaultParam(x string) {
    m.data = append(m.data, &MmlMacroElement{DEFAULT_PARAM, x})
}

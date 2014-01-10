/*
 * Package effects
 *
 * Part of XPMC.
 *
 * /Mic, 2012-2013
 */
 
package effects

import (
    "../utils"
)


type EffectMap struct {
    data []*utils.ParamList
    keys []int
    refCount []int
    extraData []map[string]interface{}
}


func (m *EffectMap) FindKey(key int) int {
    return utils.PositionOfInt(m.keys, key)
}

func (m *EffectMap) GetKeys() []int {
    return m.keys
}

func (m *EffectMap) GetKeyAt(pos int) int {
    if pos < len(m.keys) {
        return m.keys[pos]
    }
    return -1
}

func (m *EffectMap) GetDataAt(pos int) *utils.ParamList {
    if pos < len(m.data) {
        return m.data[pos]
    }
    return nil
}

/* Checks if an effect with the given parameter list already exists, and if so
 * returns its key
 */
func (m *EffectMap) GetKeyFor(lst *utils.ParamList) int {
    for i, key := range m.keys {
        if lst.Equal(m.data[i]) {
            return key
        }
    }
    return -1
}

func (m *EffectMap) GetData(key int) *utils.ParamList {
    pos := m.FindKey(key)
    if pos != -1 {
        return m.data[pos]
    }
    return nil
}

func (m *EffectMap) Append(key int, lst *utils.ParamList) {
    m.keys = append(m.keys, key)
    m.data = append(m.data, lst)
    m.refCount = append(m.refCount, 0)
    m.extraData = append(m.extraData, map[string]interface{}{})
}

func (m *EffectMap) PutExtraInt(key int, name string, val int) {
    pos := m.FindKey(key)
    if (pos != -1) {
        m.extraData[pos][name] = val
    }
}

func (m *EffectMap) GetExtraInt(key int, name string) int {
    pos := m.FindKey(key)
    if (pos != -1) {
        return m.extraData[pos][name].(int)
    }
    return 0
}

func (m *EffectMap) PutExtraString(key int, name string, val string) {
    pos := m.FindKey(key)
    if (pos != -1) {
        m.extraData[pos][name] = val
    }
}

func (m *EffectMap) GetExtraString(key int, name string) string {
    pos := m.FindKey(key)
    if (pos != -1) {
        return m.extraData[pos][name].(string)
    }
    return ""
}

func (m *EffectMap) IsReferenced(key int) bool {
    pos := m.FindKey(key)
    if (pos != -1) {
        return (m.refCount[pos] > 0)
    }
    return false
}

func (m *EffectMap) AddRef(key int) {
    pos := m.FindKey(key)
    if (pos != -1) {
        m.refCount[pos] += 1
    }
}

func (m *EffectMap) IsEmpty(key int) bool {
    pos := m.FindKey(key)
    if (pos != -1) {
        return m.data[pos].IsEmpty()
    }
    return true
}

func (m *EffectMap) Len() int {
    return len(m.keys)
}

func NewEffectMap() *EffectMap {
    e := &EffectMap{}
    return e
}

var Arpeggios *EffectMap
var Vibratos *EffectMap
var VolumeMacros *EffectMap
var DutyMacros *EffectMap
var PanMacros *EffectMap
var PitchMacros *EffectMap
var PulseMacros *EffectMap
var FeedbackMacros *EffectMap
var MODs *EffectMap
var ADSRs *EffectMap
var Filters *EffectMap
var Portamentos *EffectMap
var Waveforms *EffectMap
var WaveformMacros *EffectMap
var PCMs *EffectMap


func Init() {
    Arpeggios = NewEffectMap()
    Vibratos = NewEffectMap()
    VolumeMacros = NewEffectMap()
    DutyMacros = NewEffectMap()
    PanMacros = NewEffectMap()
    PitchMacros = NewEffectMap()
    PulseMacros = NewEffectMap()
    FeedbackMacros = NewEffectMap()
    MODs = NewEffectMap()
    ADSRs = NewEffectMap()
    Filters = NewEffectMap()
    Portamentos = NewEffectMap()
    Waveforms = NewEffectMap()
    WaveformMacros = NewEffectMap()
    PCMs = NewEffectMap()
}

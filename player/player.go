/*
 * Package player
 *
 * Part of XPMC.
 *
 * /Mic, 2013
 */
 
package player

import (
    "../utils"
)

const (
    EFFECT_STEP_MASK = 0x80
    EFFECT_STEP_EVERY_FRAME = 0
    EFFECT_STEP_EVERY_NOTE = 0x80
)

const (
    ARPEGGIO_CUMULATIVE = 0
    ARPEGGIO_ABSOLUTE = 1
)

const (
    NEW_EFFECT_VALUE = 1
    NEW_NOTE = 2
)

type volume struct {
    Vol int
    Op []int
}

type PlaybackChannel struct {
    DataPtr int
    DataPos int
    Delay int
    Note int
    NoteOffs int
    Octave int
    Duty int
    Freq int
    FreqOffs int
    FreqOffsLatch int
    Volume volume
    VolOffs int
    VolOffsLatch int
    VolMac *VolSlideEffect
    ArpMac *ArpeggioEffect
    Arp2Mac *ArpeggioEffect
    EpMac *FreqSlideEffect
    MpMac *VibratoEffect
    
    Operator int
    /* 19 loops,
    -- 20 loopIndex,
    -- 21 detune,
    -- 22 csMac,
    -- 23 mode,
    -- 24 feedback,
    -- 25 adsr,
    -- 26 operator,
    -- 27 mult,
    -- 28 ams,
    -- 29 fms,
    -- 30 lfo,
    -- 31 rateScale,
    -- 32 dtMac,
    -- 33 fbMac,
    -- 34 pattern,
    -- 35 chnNum,
    -- 36 oldPos,*/
    PrevNote int
    DelayLatch int
    Transpose int

    FreqChange int
    VolChange int
}

func NewPlaybackChannel(num int) *PlaybackChannel {
    return &PlaybackChannel{
        DataPtr: 0,
        DataPos: 0,
        Delay: 0,
        Note: 0,
        NoteOffs: 0,
        Octave: 0,
        Duty: 0,
        Freq: 0,
        FreqOffs: 0,
        FreqOffsLatch: 0,
        Volume: volume{0, []int{}},
        VolOffs: 0,
        VolOffsLatch: 0,
        VolMac: NewVolSlideEffect(),
        ArpMac: NewArpeggioEffect(ARPEGGIO_CUMULATIVE),
        Arp2Mac: NewArpeggioEffect(ARPEGGIO_ABSOLUTE),
        EpMac: NewFreqSlideEffect(),
        MpMac: NewVibratoEffect(),
        /*NewDutyEffect(),
        NewFeedbackEffect(),
        NewPanningEffect(),*/
    }
}        
type IEffectMacro interface {
    Step(*PlaybackChannel, int)
    Disable()
}

type EffectMacro struct {
    Enabled bool
    ID int
    Params utils.ParamList
    SubType int
}

// @EN
type ArpeggioEffect struct {
    *EffectMacro
}

// @v
type VolSlideEffect struct {
    *EffectMacro
}

// @EP
type FreqSlideEffect struct {
    *EffectMacro
}

// @MP
type VibratoEffect struct {
    *EffectMacro
}

func (e *EffectMacro) Step(c *PlaybackChannel, trigger int) bool {
    if e.Enabled {
        if trigger == EFFECT_STEP_EVERY_NOTE && ((e.ID & EFFECT_STEP_MASK) == EFFECT_STEP_EVERY_FRAME) {
            e.Params.MoveToStart()
        }
        if trigger == EFFECT_STEP_EVERY_NOTE || ((e.ID & EFFECT_STEP_MASK) == EFFECT_STEP_EVERY_FRAME) {
            e.Params.Step()
            return true
        }           
    }
    return false
}

func (e *EffectMacro) Disable() {
    e.Enabled = false
}

func (e *ArpeggioEffect) Update(c *PlaybackChannel, trigger int) {
    old := c.NoteOffs
    if e.SubType == ARPEGGIO_ABSOLUTE || (e.ID & EFFECT_STEP_MASK) != trigger {
        c.NoteOffs = e.Params.Peek()
    } else {
        c.NoteOffs += e.Params.Peek()
    }
    if old != c.NoteOffs && c.FreqChange != NEW_NOTE {
        c.FreqChange = NEW_EFFECT_VALUE
    }
}

func (e *ArpeggioEffect) Step(c *PlaybackChannel, trigger int) {
    if e.EffectMacro.Step(c, trigger) {
        e.Update(c, trigger)
    }
}

func (e *FreqSlideEffect) Update(c *PlaybackChannel, trigger int) {
    if (e.ID & 0x80) != trigger {
        c.FreqOffs = e.Params.Peek()
    } else {
        c.FreqOffs += e.Params.Peek()
    }
    if c.FreqChange != NEW_NOTE {
        c.FreqChange = NEW_EFFECT_VALUE
    }
}

func (e *VolSlideEffect) Update(c *PlaybackChannel, trigger int) {
    if len(c.Volume.Op) > 0 {
        if c.Operator != 0 {
            c.Volume.Op[c.Operator - 1] = e.Params.Peek()
        } else {
            for i := 0; i < len(c.Volume.Op); i++ {
                c.Volume.Op[i] = e.Params.Peek()
            }
        }
    } else {
        c.Volume.Vol = e.Params.Peek()
    }
    c.VolChange = 1
}


func NewArpeggioEffect(subType int) *ArpeggioEffect {
    e := &ArpeggioEffect{new (EffectMacro)}
    e.SubType = subType
    return e
}

func NewVolSlideEffect() *VolSlideEffect {
    e := &VolSlideEffect{new (EffectMacro)}
    return e
}

func NewFreqSlideEffect() *FreqSlideEffect {
    e := &FreqSlideEffect{new (EffectMacro)}
    return e
}

func NewVibratoEffect() *VibratoEffect {
    e := &VibratoEffect{new (EffectMacro)}
    return e
}
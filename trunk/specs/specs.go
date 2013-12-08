/*
 * Package specs
 *
 * Part of XPMC.
 * Contains data/functions for dealing with the specifications
 * of different sound chips.
 *
 * /Mic, 2012-2013
 */
 
package specs

const (
    CHIP_SN76489    = 1
    CHIP_YM2151     = 2
    CHIP_YM2413     = 3
    CHIP_YM2612     = 4
    CHIP_YM3812     = 5
    CHIP_AY_3_8910  = 6
    CHIP_SCC        = 7
    CHIP_SID        = 8
    CHIP_POKEY      = 9
    CHIP_MIKEY      = 10
    CHIP_YMF292     = 11
    CHIP_SPC        = 12
    CHIP_GBAPU      = 13
    CHIP_HUC6280    = 14
    CHIP_2A03       = 15
    CHIP_VRC6       = 16
    CHIP_N106       = 17
    CHIP_T6W28      = 18
    CHIP_UNKNOWN    = 99
)


type Specs struct {
    Duty []int          // Supports @
    VolChange []int     // Supports volume change
    FM []int            // Supports FM operations (contains the # of operators)
    ADSR []int          // Supports ADSR envelopes
    Filter []int        // Supports filter
    RingMod []int       // Supports ring modulation
    WaveTable []int     // Supports WT
    PCM []int           // Supports XPCM
    ToneEnv []int       // Supports @te
    VolEnv []int        // Supports @ve
    Detune []int        // Supports detuning
    MinOct []int        // Min octave
    MaxOct []int        // Max octave
    MaxVol []int        // Max volume
    MinNote []int       // Lowest playable note
    ID int              // Identifier
    IDs []int           
    Name string
}


/* ISpecs interface implementation *
/***********************************/

func (s *Specs) GetDuty() []int {
    return s.Duty
}

func (s *Specs) GetVolChange() []int {
    return s.VolChange
}

func (s *Specs) GetFM() []int {
    return s.FM
}

func (s *Specs) GetADSR() []int {
    return s.ADSR
}

func (s *Specs) GetFilter() []int {
    return s.Filter
}

func (s *Specs) GetRingMod() []int {
    return s.RingMod
}

func (s *Specs) GetWaveTable() []int {
    return s.WaveTable
}

func (s *Specs) GetPCM() []int {
    return s.PCM
}

func (s *Specs) GetToneEnv() []int {
    return s.ToneEnv
}

func (s *Specs) GetVolEnv() []int {
    return s.VolEnv
}

func (s *Specs) GetDetune() []int {
    return s.Detune
}

func (s *Specs) GetMinOct() []int {
    return s.MinOct
}

func (s *Specs) GetMaxOct() []int {
    return s.MaxOct
}

func (s *Specs) GetMaxVol() []int {
    return s.MaxVol
}

func (s *Specs) GetMinNote() []int {
    return s.MinNote
}

func (s *Specs) GetID() int {
    return s.ID
}

func (s *Specs) GetIDs() []int {
    return s.IDs
}

/**************/


/* 2A03 (NES) *
/**************/

var Specs2A03 = Specs{
    Duty:       []int{ 3,   3,  -1,   1, -1},   
    VolChange:  []int{ 1,   1,   0,   1,  0},   
    WaveTable:  []int{ 0,   0,   1,   0,  0},   
    PCM:        []int{ 0,   0,   0,   0,  1},   
    ToneEnv:    []int{ 64,  64,  0,   0,  0},   
    VolEnv:     []int{ 0,   0,   0,   7,  0},   
    Detune:     []int{ 1,   1,   0,   0,  0},   
    MinOct:     []int{ 1,   1,   0,   0,  0},   
    MaxOct:     []int{ 9,   9,   9,   9,  9},   
    MaxVol:     []int{15,  15,   0,  15,  0},   
    MinNote:    []int{10,  10,  10,   1,  1},
    ID:         CHIP_2A03,
}


/* AY-3-8910 (Amstrad CPC, MSX etc) *
/************************************/

var SpecsAY_3_8910 = Specs{
    Duty:       []int{ 7,   7,   7},    
    VolChange:  []int{ 1,   1,   1},    
    FM:         []int{ 0,   0,   0},    
    ADSR:       []int{ 0,   0,   0},    
    Filter:     []int{ 0,   0,   0},
    RingMod:    []int{ 0,   0,   0},
    WaveTable:  []int{ 0,   0,   0},    
    PCM:        []int{ 0,   0,   0},    
    ToneEnv:    []int{15,  15,  15},    
    VolEnv:     []int{15,  15,  15},
    Detune:     []int{ 1,   1,   1},    
    MinOct:     []int{ 1,   1,   1},    
    MaxOct:     []int{ 7,   7,   7},    
    MaxVol:     []int{15,  15,  15},    
    MinNote:    []int{ 1,   1,   1},    
    ID:         CHIP_AY_3_8910,
}


/* Gameboy (DMG/CGB/SGB) APU *
/*****************************/

var SpecsGBAPU = Specs{
    Duty:       []int{ 3,   3,  -1,   1},       
    VolChange:  []int{ 1,   1,   1,   1},   
    FM:         []int{ 0,   0,   0,   0},       
    ADSR:       []int{ 0,   0,   0,   0},   
    Filter:     []int{ 0,   0,   0,   0},   
    RingMod:    []int{ 0,   0,   0,   0},   
    WaveTable:  []int{ 0,   0,   1,   0},   
    PCM:        []int{ 0,   0,   0,   0},   
    ToneEnv:    []int{ 7,   0,   0,   0},   
    VolEnv:     []int{ 7,   7,   0,   7},   
    Detune:     []int{ 1,   1,   0,   0},   
    MinOct:     []int{ 2,   2,   1,   1},   
    MaxOct:     []int{ 7,   7,   7,  11},   
    MaxVol:     []int{15,  15,   3,  15},   
    MinNote:    []int{ 1,   1,   1,   1},   // C 
    ID:         CHIP_GBAPU,
}


/* HuC6280 (PC-Engine) *
/***********************/

var SpecsHuC6280 = Specs{
    Duty:       []int{ 0,   0,   0,   0,   1,   1}, 
    VolChange:  []int{ 1,   1,   1,   1,   1,   1}, 
    WaveTable:  []int{ 1,   1,   1,   1,   1,   1},
    PCM:        []int{ 0,   0,   0,   1,   0,   0}, 
    Detune:     []int{ 1,   1,   1,   1,   1,   1}, 
    MinOct:     []int{ 1,   1,   1,   0,   1,   1}, 
    MaxOct:     []int{12,  12,  12,  12,  12,  12}, 
    MaxVol:     []int{31,  31,  31,  31,  31,  31}, 
    MinNote:    []int{ 1,   1,   1,   1,   1,   1},
    ID:         CHIP_HUC6280,
}


/* POKEY (Atari XL/XE etc) *
/***************************/

var SpecsPokey = Specs{
    Duty:       []int{ 7,   7,   7,   7},   
    VolChange:  []int{ 1,   1,   1,   1},   
    FM:         []int{ 0,   0,   0,   0},   
    ADSR:       []int{ 0,   0,   0,   0},   
    Filter:     []int{ 1,   1,   1,   1},   
    RingMod:    []int{ 0,   0,   0,   0},   
    WaveTable:  []int{ 0,   0,   0,   0},   
    PCM:        []int{ 0,   0,   0,   0},   
    ToneEnv:    []int{ 0,   0,   0,   0},   
    VolEnv:     []int{ 0,   0,   0,   0},   
    Detune:     []int{ 1,   1,   1,   1},   
    MinOct:     []int{ 1,   1,   1,   1},   
    MaxOct:     []int{ 9,   9,   9,   9},   
    MaxVol:     []int{15,  15,  15,  15},   
    MinNote:    []int{ 1,   1,   1,   1},   
    ID:         CHIP_POKEY,
}   



/* Konami SCC *
/**************/

var SpecsSCC = Specs{
    Duty:       []int{ -1,  -1,  -1,  -1, -1},
    VolChange:  []int{  1,   1,   1,   1,  1},
    FM:         []int{  0,   0,   0,   0,  0},
    ADSR:       []int{  0,   0,   0,   0,  0},
    Filter:     []int{  0,   0,   0,   0,  0},
    RingMod:    []int{  0,   0,   0,   0,  0},
    WaveTable:  []int{  1,   1,   1,   1,  1},
    PCM:        []int{  0,   0,   0,   0,  0},
    ToneEnv:    []int{  0,   0,   0,   0,  0},
    VolEnv:     []int{  0,   0,   0,   0,  0},
    Detune:     []int{  1,   1,   1,   1,  1},
    MinOct:     []int{  1,   1,   1,   1,  1},
    MaxOct:     []int{  7,   7,   7,   7,  7},
    MaxVol:     []int{ 15,  15,  15,  15, 15},
    MinNote:    []int{  1,   1,   1,   1,  1},
    ID:         CHIP_SCC,
}


/* MOS 6581/8580 SID (C64) *
/***************************/

var SpecsSID = Specs{
    Duty:       []int{ 3,   3,   3},    
    VolChange:  []int{ 1,   1,   1},    
    FM:         []int{ 0,   0,   0},    
    ADSR:       []int{ 1,   1,   1},    
    Filter:     []int{ 1,   1,   1},    
    RingMod:    []int{ 1,   1,   1},    
    WaveTable:  []int{ 0,   0,   0},
    PCM:        []int{ 0,   0,   0},
    ToneEnv:    []int{ 0,   0,   0},
    VolEnv:     []int{ 0,   0,   0},
    Detune:     []int{ 1,   1,   1},    
    MinOct:     []int{ 0,   0,   0},
    MaxOct:     []int{ 7,   7,   7},
    MaxVol:     []int{15,  15,  15},
    MinNote:    []int{ 1,   1,   1},
    ID:         CHIP_SID,
}


/* SN76489 (SMS, GameGear, ColecoVision etc) *
/*********************************************/

var SpecsSN76489 = Specs{
    Duty:       []int{-1,  -1,  -1,   1},   
    VolChange:  []int{ 1,   1,   1,   1},   
    FM:         []int{ 0,   0,   0,   0},   
    ADSR:       []int{ 0,   0,   0,   0},   
    Filter:     []int{ 0,   0,   0,   0},   
    RingMod:    []int{ 0,   0,   0,   0},   
    WaveTable:  []int{ 0,   0,   0,   0},   
    PCM:        []int{ 0,   0,   0,   0},   
    ToneEnv:    []int{ 0,   0,   0,   0},   
    VolEnv:     []int{ 0,   0,   0,   0},   
    Detune:     []int{ 1,   1,   1,   0},   
    MinOct:     []int{ 2,   2,   2,   2},   
    MaxOct:     []int{ 7,   7,   7,   7},   
    MaxVol:     []int{15,  15,  15,  15},   
    MinNote:    []int{10,  10,  10,  10},   // A
    ID:         CHIP_SN76489,
}


/* SPC (S-DSP) (SNES) *
/**********************/

var SpecsSPC = Specs{
    Duty:       []int{  1,   1,   1,   1,   1,   1,   1,   1},  
    VolChange:  []int{  1,   1,   1,   1,   1,   1,   1,   1},  
    FM:         []int{  0,   0,   0,   0,   0,   0,   0,   0},  
    ADSR:       []int{  1,   1,   1,   1,   1,   1,   1,   1},  
    Filter:     []int{  1,   1,   1,   1,   1,   1,   1,   1},  
    RingMod:    []int{  0,   0,   0,   0,   0,   0,   0,   0},  
    WaveTable:  []int{  1,   1,   1,   1,   1,   1,   1,   1},  
    PCM:        []int{  1,   1,   1,   1,   1,   1,   1,   1},  
    ToneEnv:    []int{  0,   0,   0,   0,   0,   0,   0,   0},  
    VolEnv:     []int{  2,   2,   2,   2,   2,   2,   2,   2},  
    Detune:     []int{  1,   1,   1,   1,   1,   1,   1,   1},  
    MinOct:     []int{  1,   1,   1,   0,   1,   1,   1,   1},  
    MaxOct:     []int{ 12,  12,  12,  12,  12,  12,  12,  12},  
    MaxVol:     []int{127, 127, 127, 127, 127, 127, 127, 127},  
    MinNote:    []int{  1,   1,   1,   1,   1,   1,   1,   1},
    ID:         CHIP_SPC,
}


/* T6W28 (NeoGeo Pocket / Color) *
/*********************************/

var SpecsT6W28 = Specs{
    Duty:       []int{-1,  -1,  -1,   1},   
    VolChange:  []int{ 1,   1,   1,   1},   
    PCM:        []int{ 0,   0,   0,   0},   
    Detune:     []int{ 1,   1,   1,   0},   
    MinOct:     []int{ 2,   2,   2,   2},   
    MaxOct:     []int{ 7,   7,   7,   7},   
    MaxVol:     []int{15,  15,  15,  15},   
    MinNote:    []int{10,  10,  10,  10},   // A
    ID:         CHIP_T6W28,
}


/* Konami VRC6 (NES) *
/*********************/

var SpecsVRC6 = Specs{
    Duty:       []int{ 8,   8,  64},    
    VolChange:  []int{ 1,   1,   0},    
    FM:         []int{ 0,   0,   0},    
    ADSR:       []int{ 0,   0,   0},    
    Filter:     []int{ 0,   0,   0},    
    RingMod:    []int{ 0,   0,   0},    
    WaveTable:  []int{ 0,   0,   1},    
    PCM:        []int{ 0,   0,   0},    
    ToneEnv:    []int{ 0,   0,   0},    
    VolEnv:     []int{ 0,   0,   0},    
    Detune:     []int{ 1,   1,   1},    
    MinOct:     []int{ 0,   0,   0},    
    MaxOct:     []int{ 9,   9,   9},    
    MaxVol:     []int{15,  15,   0},    
    MinNote:    []int{10,  10,  10},
    ID:         CHIP_VRC6,
}


var SpecsYM2151 = Specs{
    Duty:       []int{  7,   7,   7,   7,   7,   7,   7,   7},   // Supports @
    VolChange:  []int{  1,   1,   1,   1,   1,   1,   1,   1},   // Supports volume change
    FM:         []int{  1,   1,   1,   1,   1,   1,   1,   1},   // Supports FM
    ADSR:       []int{  1,   1,   1,   1,   1,   1,   1,   1},   // Supports ADSR
    Detune:     []int{  1,   1,   1,   1,   1,   1,   1,   1},
    MinOct:     []int{  1,   1,   1,   1,   1,   1,   1,   1},   // Min octave
    MaxOct:     []int{  7,   7,   7,   7,   7,   7,   7,   7},   // Max octave
    MaxVol:     []int{127, 127, 127, 127, 127, 127, 127, 127},   // Maximum volume
    MinNote:    []int{  1,   1,   1,   1,   1,   1,   1,   1},   // Min playable note
    ID:         CHIP_YM2151,
}

/* YM2413 (SMS etc) *
/********************/

var SpecsYM2413 = Specs{
    Duty:       []int{15,  15,  15,  15,  15,  15,  15,  15,  15},  
    VolChange:  []int{ 1,   1,   1,   1,   1,   1,   1,   1,   1},  
    FM:         []int{ 2,   2,   2,   2,   2,   2,   2,   2,   2},  
    ADSR:       []int{ 1,   1,   1,   1,   1,   1,   1,   1,   1},  
    Filter:     []int{ 0,   0,   0,   0,   0,   0,   0,   0,   0},  
    RingMod:    []int{ 0,   0,   0,   0,   0,   0,   0,   0,   0},  
    WaveTable:  []int{ 0,   0,   0,   0,   0,   0,   0,   0,   0},  
    PCM:        []int{ 0,   0,   0,   0,   0,   0,   0,   0,   0},  
    ToneEnv:    []int{ 1,   1,   1,   1,   1,   1,   1,   1,   1},  
    VolEnv:     []int{63,  63,  63,  63,  63,  63,  63,  63,  63},  
    Detune:     []int{ 1,   1,   1,   1,   1,   1,   1,   1,   1},  
    MinOct:     []int{ 1,   1,   1,   1,   1,   1,   1,   1,   1},  
    MaxOct:     []int{ 7,   7,   7,   7,   7,   7,   7,   7,   7},  
    MaxVol:     []int{15,  15,  15,  15,  15,  15,  15,  15,  15},  
    MinNote:    []int{ 1,   1,   1,   1,   1,   1,   1,   1,   1},  
    ID:         CHIP_YM2413,
}


/* YM2612 (Sega Genesis etc) *
/*****************************/

var SpecsYM2612 = Specs{
    Duty:       []int{  7,   7,   7,   7,   7,   7},    
    VolChange:  []int{  1,   1,   1,   1,   1,   1},    
    FM:         []int{  4,   4,   4,   4,   4,   4},    
    ADSR:       []int{  1,   1,   1,   1,   1,   1},    
    Filter:     []int{  0,   0,   0,   0,   0,   0},    
    RingMod:    []int{  0,   0,   0,   0,   0,   0},    
    WaveTable:  []int{  0,   0,   0,   0,   0,   1},    
    PCM:        []int{  0,   0,   0,   0,   0,   1},
    ToneEnv:    []int{  0,   0,   0,   0,   0,   0},
    VolEnv:     []int{  0,   0,   0,   0,   0,   0},
    Detune:     []int{  1,   1,   1,   1,   1,   1},
    MinOct:     []int{  1,   1,   1,   1,   1,   0},    
    MaxOct:     []int{  7,   7,   7,   7,   7,   7},    
    MaxVol:     []int{127, 127, 127, 127, 127, 127},    
    MinNote:    []int{  1,   1,   1,   1,   1,   1},    
    ID:         CHIP_YM2612,
}


var ChannelSpecs Specs



// Adds data to slice s at position pos, extending the length of the
// slice if necessary.
//
// E.g.  x is {1,2,3,4}
//       y is {6,7,8,9}
//       insertspec(&x, 3, y) would set x to {1,2,3,6,7,8,9}
//
func insertspec(s *[]int, pos int, data []int) {
    if pos + len(data) >= len(*s) {
        *s = append(*s, make([]int, pos + len(data) - len(*s))...)
    }
    for i, d := range data { 
        (*s)[pos + i] = d
    }
}   


func SetChannelSpecs(dest *Specs, firstPhysChan int, firstLogicalChan int, s Specs) {
    insertspec(&dest.Duty,      firstLogicalChan, s.Duty[firstPhysChan:])
    insertspec(&dest.VolChange, firstLogicalChan, s.VolChange[firstPhysChan:])
    insertspec(&dest.FM,        firstLogicalChan, s.FM[firstPhysChan:])
    insertspec(&dest.ADSR,      firstLogicalChan, s.ADSR[firstPhysChan:])
    insertspec(&dest.Filter,    firstLogicalChan, s.Filter[firstPhysChan:])
    insertspec(&dest.RingMod,   firstLogicalChan, s.RingMod[firstPhysChan:])
    insertspec(&dest.WaveTable, firstLogicalChan, s.WaveTable[firstPhysChan:])
    insertspec(&dest.PCM,       firstLogicalChan, s.PCM[firstPhysChan:])
    insertspec(&dest.ToneEnv,   firstLogicalChan, s.ToneEnv[firstPhysChan:])
    insertspec(&dest.VolEnv,    firstLogicalChan, s.VolEnv[firstPhysChan:])
    insertspec(&dest.Detune,    firstLogicalChan, s.Detune[firstPhysChan:])
    insertspec(&dest.MinOct,    firstLogicalChan, s.MinOct[firstPhysChan:])
    insertspec(&dest.MaxOct,    firstLogicalChan, s.MaxOct[firstPhysChan:])
    insertspec(&dest.MaxVol,    firstLogicalChan, s.MaxVol[firstPhysChan:])
    insertspec(&dest.MinNote,   firstLogicalChan, s.MinNote[firstPhysChan:])
    
    tempIDs := make([]int, len(s.FM[firstPhysChan:]))
    for i, _ := range tempIDs {
        tempIDs[i] = s.ID
    }
    insertspec(&dest.IDs,       firstLogicalChan, tempIDs)
}




/*
 * Package targets
 *
 * Part of XPMC.
 * Contains data/functions describing different output targets.
 *
 * /Mic, 2012
 */
 
package targets

import (
    "fmt"
    "../specs"
    //"../timing"
)

const (
    TARGET_UNKNOWN = 0
    TARGET_SMS = 1      // SEGA Master System 
    TARGET_NES = 2      // Nintendo Entertainment System
    TARGET_GBA = 3      // Gameboy Advance
    TARGET_NDS = 4      // Nintendo DS
    TARGET_GBC = 5      // Gameboy Color (and DMG) 
    TARGET_SMD = 6      // SEGA Megadrive 
    TARGET_SAT = 7      // SEGA Saturn
    TARGET_XGS = 8      // XGameStation Micro Edition
    TARGET_SGG = 9      // SEGA Game Gear 
    TARGET_CPS = 10     // Capcom Play System (should be replaced by TARGET_VGM)
    TARGET_X68 = 11     // X68000
    TARGET_AST = 12     // Atari ST
    TARGET_C64 = 13     // Commodore 64 
    TARGET_PCE = 14     // NEC PC Engine / TurboGrafx 16
    TARGET_ZXS = 15     // ZX Spectrum
    TARGET_PC4 = 16     // PC 4k synth
    TARGET_CLV = 17     // ColecoVision 
    TARGET_KSS = 18
    TARGET_CPC = 19     // Amstrad CPC 
    TARGET_AT8 = 20     // Atari 8-bit (XL/XE etc)
    TARGET_MSX = 21     // MSX
    TARGET_VGM = 22     // For all-out fantasy machine chip fest tracks
    TARGET_SFC = 23     // Super Famicom (and SNES)
    TARGET_LYX = 24     // Atari Lynx
    TARGET_NGP = 25     // NeoGeo Pocket Color
        
    TARGET_LAST = 26
)



type ITarget interface {
    GetAdsrLen() int
    GetAdsrMax() int
    GetChannelSpecs() *specs.Specs
    GetChannelNames() string
    GetID() int
    GetMaxLoopDepth() int
    GetMaxTempo() int
    GetMaxVolume() int
    GetMaxWavLength() int
    GetMaxWavSample() int
    GetMinVolume() int
    GetMinWavLength() int
    GetMinWavSample() int
    Init()
    Output(outputVgm int)
    SupportsPAL() bool
    SupportsPan() bool
    SupportsPCM() bool
    SupportsWaveTable() bool
}

type Target struct {
    MaxTempo int
    MinVolume int
    SupportsPanning int
    MaxLoopDepth int
    SupportsPal bool
    MinWavLength int
    MaxWavLength int
    MinWavSample int
    MaxWavSample int
    AdsrLen int
    AdsrMax int
    ChannelSpecs specs.Specs
    MachineSpeed int
    ID int
}

type TargetAt8 struct {
    Target
}

type TargetGBC struct {
    Target
}

type TargetGen struct {
    Target
}

type TargetKSS struct {
    Target
}

type TargetNES struct {
    Target
}

type TargetNGP struct {
    Target
}

type TargetSGG struct {
    Target
}

type TargetSMS struct {
    Target
}


func NewTarget(tID int) ITarget {
    switch tID {
    case TARGET_AT8:
        return &TargetAt8{}
    case TARGET_GBC:
        return &TargetGBC{}
    case TARGET_SMD:
        return &TargetGen{}
    case TARGET_KSS:
        return &TargetKSS{}
    case TARGET_SGG:
        return &TargetSGG{}
    case TARGET_SMS:
        return &TargetSMS{}
    }
    return ITarget(nil)
}

        
func (t *Target) Init() {
    // Stub to fulfill the ITarget interface
}

func (t *Target) Output(outputVgm int) {
    // Stub to fulfill the ITarget interface
}

func (t *Target) GetID() int {
    return t.ID
}

func (t *Target) GetAdsrLen() int {
    return t.AdsrLen
}

func (t *Target) GetAdsrMax() int {
    return t.AdsrMax
}

func (t *Target) GetChannelNames() string {
    names := ""
    for i, _ := range t.ChannelSpecs.Duty {
        names += fmt.Sprintf("%c", 'A'+i)
    }
    return names
}

func (t *Target) GetChannelSpecs() *specs.Specs {
    return &t.ChannelSpecs
}

func (t *Target) GetMaxTempo() int {
    return t.MaxTempo
}

func (t *Target) GetMaxLoopDepth() int {
    return t.MaxLoopDepth
}

func (t *Target) GetMinVolume() int {
    return t.MinVolume
}

func (t *Target) GetMaxVolume() int {
    maxVol := t.ChannelSpecs.MaxVol[0]
    for _, v := range t.ChannelSpecs.MaxVol {
        if v > maxVol {
            maxVol = v
        }
    }
    return maxVol
}

func (t *Target) GetMinWavLength() int {
    return t.MinWavLength
}

func (t *Target) GetMaxWavLength() int {
    return t.MaxWavLength
}

func (t *Target) GetMinWavSample() int {
    return t.MinWavSample
}

func (t *Target) GetMaxWavSample() int {
    return t.MaxWavSample
}

func (t *Target) SupportsPAL() bool {
    return t.SupportsPal
}

func (t *Target) SupportsPan() bool {
    return t.SupportsPanning > 0
}

func (t *Target) SupportsPCM() bool {
    supported := false
    for _, p := range t.ChannelSpecs.PCM {
        if p != 0 {
            supported = true
            break
        }
    }
    return supported
}

func (t *Target) SupportsWaveTable() bool {
    supported := false
    for _, w := range t.ChannelSpecs.WaveTable {
        if w != 0 {
            supported = true
            break
        }
    }
    return supported
}


func DefineSymbol(sym string, val int) {
}


/* Atari 8-bit (XL/XE etc) *
 ***************************/

func (t *TargetAt8) Init() {
    DefineSymbol("AT8", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsPokey)      // A..D
    
    //activeChannels    = repeat(0, length(supportedChannels))
    
    t.ID                = TARGET_AT8
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.SupportsPal       = true
    //timing.UpdateFreq     = 50.0  // Use PAL by default
}


/* Gameboy / Color *
 *******************/

func (t *TargetGBC) Init() {
    DefineSymbol("DMG", 1)
    DefineSymbol("GBC", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsGBAPU)      // A..D

    t.ID                = TARGET_GBC
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.MinWavLength      = 32
    t.MaxWavLength      = 32
    t.MinWavSample      = 0
    t.MaxWavSample      = 15
}


/* Sega Genesis / Megadrive *
 ****************************/

func (t *TargetGen) Init() {
    DefineSymbol("GEN", 1)
    DefineSymbol("SMD", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 4, specs.SpecsYM2612)     // E..J

    //activeChannels        = repeat(0, length(supportedChannels))

    t.ID                = TARGET_SMD
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.SupportsPal       = true
    t.AdsrLen           = 5
    t.AdsrMax           = 63
    t.MinWavLength      = 1
    t.MaxWavLength      = 2097152 // 2MB
    t.MinWavSample      = 0
    t.MaxWavSample      = 255
    t.MachineSpeed      = 3579545
}


/* KSS (MSX executable music) *
 ******************************/

func (t *TargetKSS) Init() {
    DefineSymbol("KSS", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 4, specs.SpecsAY_3_8910)  // E..G
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 7, specs.SpecsSCC)        // H..L
    //specs.SetChannelSpecs(&t.ChannelSpecs, 0, 12, specs.SpecsYM2151)  // M..T
    
    //activeChannels    = repeat(0, length(supportedChannels))  
    
    t.ID                = TARGET_KSS
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.AdsrLen           = 5
    t.AdsrMax           = 63
    t.MinWavLength      = 32
    t.MaxWavLength      = 32
    t.MinWavSample      = -128
    t.MaxWavSample      = 127
}


/* NES / Famicom *
 *****************/

func (t *TargetNES) Init() {
    DefineSymbol("NES", 1)       
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.Specs2A03)       // A..E
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 5, specs.SpecsVRC6)       // F..H
    
    //activeChannels        = repeat(0, length(supportedChannels))  
    
    t.ID                = TARGET_NES
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.MaxLoopDepth      = 2
    t.SupportsPal       = true
}


/* NeoGeo Pocket / Color *
 *************************/

func (t *TargetNGP) Init() {
    DefineSymbol("NGP", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsT6W28)      // A..D
    
    t.ID                = TARGET_NGP
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.MachineSpeed      = 3072000
}


/* Sega Gamegear *
 *****************/

func (t *TargetSGG) Init() {
    DefineSymbol("SGG", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    
    t.ID                = TARGET_SGG
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.MachineSpeed      = 3579545
}


/* Sega Master System *
 **********************/

func (t *TargetSMS) Init() {
    DefineSymbol("SMS", 1)       
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 4, specs.SpecsYM2413)     // E..M
    
    //activeChannels        = repeat(0, length(supportedChannels))  
    
    t.ID                = TARGET_SMS
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.MaxLoopDepth      = 2
    t.MachineSpeed      = 3579545
    t.SupportsPal       = true
}



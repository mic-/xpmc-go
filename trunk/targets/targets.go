/*
 * Package targets
 *
 * Part of XPMC.
 * Contains data/functions describing different output targets.
 *
 * /Mic, 2012-2014
 */
 
package targets

import (
    "fmt"
    "os"
    "../specs"
    "../utils"
    "../effects"
)

import . "../defs"

const (
    TARGET_UNKNOWN = 0
    TARGET_SMS = 1      // SEGA Master System 
    TARGET_NES = 2      // Nintendo Entertainment System
    TARGET_GBA = 3      // Gameboy Advance
    TARGET_NDS = 4      // Nintendo DS
    TARGET_GBC = 5      // Gameboy Color (and DMG) 
    TARGET_SMD = 6      // SEGA Megadrive / Genesis
    TARGET_SAT = 7      // SEGA Saturn
    TARGET_XGS = 8      // XGameStation Micro Edition
    TARGET_SGG = 9      // SEGA Game Gear 
    TARGET_CPS = 10     // Capcom Play System (should be replaced by TARGET_VGM)
    TARGET_X68 = 11     // X68000
    TARGET_AST = 12     // Atari ST
    TARGET_C64 = 13     // Commodore 64 
    TARGET_PCE = 14     // NEC PC-Engine / TurboGrafx 16
    TARGET_ZXS = 15     // ZX Spectrum
    TARGET_PC4 = 16     // PC 4k synth
    TARGET_CLV = 17     // ColecoVision 
    TARGET_KSS = 18     // KSS (MSX music rips)
    TARGET_CPC = 19     // Amstrad CPC 
    TARGET_AT8 = 20     // Atari 8-bit (XL/XE etc)
    TARGET_MSX = 21     // MSX
    TARGET_VGM = 22     // For all-out fantasy machine chip fest tracks
    TARGET_SFC = 23     // Super Famicom (and SNES)
    TARGET_LYX = 24     // Atari Lynx
    TARGET_NGP = 25     // NeoGeo Pocket Color
        
    TARGET_LAST = 26
)

const (
    FORMAT_WLA_DX = 0
    FORMAT_GAS_68K = 1
)


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
    CompilerItf ICompiler
    MachineSpeed int
    ID int
}

type TargetAST struct {
    Target
}

type TargetAt8 struct {
    Target
}

type TargetC64 struct {
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

type TargetPCE struct {
    Target
}

type TargetSGG struct {
    Target
}

type TargetSMS struct {
    Target
}


func NewTarget(tID int, icomp ICompiler) ITarget {
    var t ITarget = ITarget(nil)
    
    switch tID {
    case TARGET_AST:
        t = &TargetAST{}

    case TARGET_AT8:
        t = &TargetAt8{}
    
    case TARGET_C64:
        t = &TargetC64{}
    
    case TARGET_GBC:
        t = &TargetGBC{}
    
    case TARGET_KSS:
        t = &TargetKSS{}

    case TARGET_NES:
        t = &TargetNES{}
    
    case TARGET_PCE:
        t = &TargetPCE{}
    
    case TARGET_SGG:
        t = &TargetSGG{}
    
    case TARGET_SMD:
        t = &TargetGen{}
    
    case TARGET_SMS:
        t = &TargetSMS{}
    }
    
    if t != nil {
        t.SetCompilerItf(icomp)
    }
    
    return t
}


/* Maps target name strings to TARGET_* int constants (e.g.
 * "sms" -> TARGET_SMS).
 */
func NameToID(targetName string) int {
    switch targetName {
    case "ast":
        return TARGET_AST;

    case "at8":
        return TARGET_AT8;
    
    case "c64":
        return TARGET_C64;
    
    case "gbc":
        return TARGET_GBC;
    
    case "kss":
        return TARGET_KSS;

    case "nes":
        return TARGET_NES;
    
    case "pce":
        return TARGET_PCE;
    
    case "sgg":
        return TARGET_SGG;
    
    case "smd", "gen":
        return TARGET_SMD;
    
    case "sms":
        return TARGET_SMS;
    }
    return TARGET_UNKNOWN;
}

        
func (t *Target) Init() {
    // Stub to fulfill the ITarget interface
}

func (t *Target) Output(outputVgm int) {
    // Stub to fulfill the ITarget interface
}


func (t *Target) SetCompilerItf(icomp ICompiler) {
    t.CompilerItf = icomp
}

/* Returns the ID (one of the TARGET_* constants) of this
 * target.
 */
func (t *Target) GetID() int {
    return t.ID
}

/* Returns the number of parameters used for ADSR envelopes
 * on this target.
 */
func (t *Target) GetAdsrLen() int {
    return t.AdsrLen
}

/* Returns the maximum value supported for any of the ADSR envelope
 * parameters on this target.
 */
func (t *Target) GetAdsrMax() int {
    return t.AdsrMax
}

func (t *Target) GetChannelNames() string {
    names := ""
    for i, _ := range t.ChannelSpecs.Duty {
        if t.ChannelSpecs.IDs[i] != specs.CHIP_UNKNOWN {
            names += fmt.Sprintf("%c", 'A'+i)
        }
    }
    return names
}

func (t *Target) GetChannelSpecs() ISpecs {
    return &t.ChannelSpecs
}

/* Returns which of the specified chip's channels the given channel
 * number corresponds to for this target.
 * For example, on the SEGA Genesis target the first 4 channels (A..D)
 * are the PSG (SN76489) channels and the last 6 channels (E..J) are
 * the FM (YM2612) channels. ChipChannel for channel E and CHIP_YM2612
 * would therefore return 0, channel D and CHIP_SN76489 would return 3,
 * and so on.
 */
func (t *Target) ChipChannel(chn, chipId int) int {
    for i, chnChipId := range t.ChannelSpecs.IDs {
        if chipId == chnChipId {
            return chn - i
        }
    }
    return -1
}

/* Returns the maximum tempo (quarter notes per minute) supported
 * by this target.
 */
func (t *Target) GetMaxTempo() int {
    return t.MaxTempo
}

/* Returns the maximum loop depth (the depth to which []-loops can
 * be nested) for this target.
 */
func (t *Target) GetMaxLoopDepth() int {
    return t.MaxLoopDepth
}

func (t *Target) GetMinVolume() int {
    return t.MinVolume
}

func (t *Target) GetMaxVolume() int {
    maxVol := t.ChannelSpecs.MaxVol[0]
    for i, v := range t.ChannelSpecs.MaxVol {
        if v > maxVol && t.ChannelSpecs.IDs[i] != specs.CHIP_UNKNOWN {
            maxVol = v
        }
    }
    return maxVol
}

/* Gets the minimum supported length for WT samples
 * for this target (for targets like the GBC and PCE that
 * have a fixed wavetable size).
 */
func (t *Target) GetMinWavLength() int {
    return t.MinWavLength
}

func (t *Target) GetMaxWavLength() int {
    return t.MaxWavLength
}

func (t *Target) GetMinWavSample() int {
    return t.MinWavSample
}

/* Gets the maximum supported amplitude for WT samples for
 * this target.
 */
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



/********************************************************************************/


/* Packs the parameters for an ADSR envelope into the format used by the given chip.
 */
func packADSR(adsr []interface{}, chipType int) []interface{} {
    packedAdsr := []interface{}{}
    
    switch chipType {
    case specs.CHIP_YM2413:
        packedAdsr = make([]interface{}, 2)
        packedAdsr[0] = adsr[0].(int) * 0x10 + adsr[1].(int)
        packedAdsr[1] = (adsr[2].(int) ^ 15) * 0x10 + adsr[3].(int)
    case specs.CHIP_YM2151, specs.CHIP_YM2612:
        packedAdsr = make([]interface{}, 4)
        packedAdsr[0] = adsr[1].(int)
        packedAdsr[1] = adsr[2].(int)
        packedAdsr[2] = adsr[3].(int)
        packedAdsr[3] = (adsr[4].(int) / 2) + ((adsr[3].(int) ^ 31) / 2) * 0x10
    }
    
    return packedAdsr
}


/* Packs the parameters for a MOD modulator macro into the format used by the given chip.
 */
func packMOD(modParams []interface{}, chipType int) []interface{} {
    packedMod := []interface{}{}
    
    switch chipType {
    case specs.CHIP_YM2151:
        packedMod = make([]interface{}, 5)
        packedMod[0] = modParams[0].(int)
        packedMod[1] = modParams[1].(int)
        packedMod[2] = modParams[2].(int) | 0x80
        packedMod[3] = modParams[3].(int) + modParams[4].(int) * 0x10
        packedMod[4] = modParams[5].(int) + 0xC0
    case specs.CHIP_YM2612:
        packedMod = make([]interface{}, 2)
        packedMod[0] = modParams[0].(int)
        packedMod[1] = modParams[1].(int) * 8 + modParams[2].(int)
    }
    
    return packedMod
}


/* Outputs the pattern data and addresses.
 */
func (t *Target) outputPatterns(outFile *os.File, outputFormat int) int {
    patSize := 0
    
    switch outputFormat {
    case FORMAT_WLA_DX:
        patterns := t.CompilerItf.GetPatterns()
        for n, pat := range patterns {
            outFile.WriteString(fmt.Sprintf("xpmp_pattern%d:", n))
            cmds := pat.GetCommands()
            for j, cmd := range cmds {
                if (j % 16) == 0 {
                    outFile.WriteString("\n.db ")
                }              
                outFile.WriteString(fmt.Sprintf("$%02x", cmd & 0xFF))
                if j < len(cmds)-1 && (j % 16) != 15 {
                    outFile.WriteString(",")
                }
            }
            outFile.WriteString("\n")
            patSize += len(cmds)
        }

        outFile.WriteString("\nxpmp_pattern_tbl:\n")
        for n := range patterns {
            outFile.WriteString(fmt.Sprintf(".dw xpmp_pattern%d\n", n))
            patSize += 2
        }
        outFile.WriteString("\n")
        
    case FORMAT_GAS_68K:
        patterns := t.CompilerItf.GetPatterns()
        for n, pat := range patterns {
            outFile.WriteString(fmt.Sprintf("xpmp_pattern%d:", n))
            cmds := pat.GetCommands()
            for j, cmd := range cmds {
                if (j % 16) == 0 {
                    outFile.WriteString("\ndc.b ")
                }              
                outFile.WriteString(fmt.Sprintf("0x%02x", cmd & 0xFF))
                if j < len(cmds)-1 && (j % 16) != 15 {
                    outFile.WriteString(",")
                }
            }
            outFile.WriteString("\n")
            patSize += len(cmds)
        }

        outFile.WriteString("\n.globl xpmp_pattern_tbl\n")
        outFile.WriteString("xpmp_pattern_tbl:\n")
        for n := range patterns {
            outFile.WriteString(fmt.Sprintf("dc.w xpmp_pattern%d\n", n))
            patSize += 2
        }
        outFile.WriteString("\n")
    }
    
    return patSize
}


/* Outputs the channel data (the actual notes, volume commands, effect invokations, etc)
 * for all channels and all songs.
 */
func (t *Target) outputChannelData(outFile *os.File, outputFormat int) int {
    songDataSize := 0
    
    switch outputFormat {
    case FORMAT_WLA_DX:
        songs := t.CompilerItf.GetSongs()
        for n, sng := range songs {
            channels := sng.GetChannels()
            if n > 0 {
                fmt.Printf("\n")
            }
            for _, chn := range channels {  
                if chn.IsVirtual() {
                    continue       
                }
                outFile.WriteString(fmt.Sprintf("xpmp_s%d_channel_%s:", n, chn.GetName()))
                commands := chn.GetCommands()
                for j, cmd := range commands {
                    if (j % 16) == 0 {
                        outFile.WriteString("\n.db ")
                    }
                    outFile.WriteString(fmt.Sprintf("$%02x", cmd & 0xFF))
                    songDataSize++
                    if j < len(commands)-1 && (j % 16) != 15 {
                       outFile.WriteString(",")
                    }
                }
                outFile.WriteString("\n")
                fmt.Printf("Song %d, Channel %s: %d bytes, %d / %d ticks\n", sng.GetNum(), chn.GetName(), len(commands), utils.Round2(float64(chn.GetTicks())), utils.Round2(float64(chn.GetLoopTicks())))
            }
        }

        outFile.WriteString("\nxpmp_song_tbl:\n")
        for n, sng := range songs {
            channels := sng.GetChannels()
            for _, chn := range channels { 
                if chn.IsVirtual() {
                    continue
                }
                outFile.WriteString(fmt.Sprintf(".dw xpmp_s%d_channel_%s\n", n, chn.GetName()))
                songDataSize += 2
            }
        }
        
    case FORMAT_GAS_68K:
        songs := t.CompilerItf.GetSongs()
        for n, sng := range songs {
            channels := sng.GetChannels()
            if n > 0 {
                fmt.Printf("\n")
            }
            for _, chn := range channels {  
                if chn.IsVirtual() {
                    continue       
                }          
                outFile.WriteString(fmt.Sprintf("xpmp_s%d_channel_%s:", n, chn.GetName()))

                commands := chn.GetCommands()
                for j, cmd := range commands {
                    if (j % 16) == 0 {
                        outFile.WriteString("\ndc.b ")
                    }
                    outFile.WriteString(fmt.Sprintf("0x%02x", cmd & 0xFF))
                    songDataSize++
                    if j < len(commands)-1 && (j % 16) != 15 {
                       outFile.WriteString(",")
                    }
                }
                outFile.WriteString("\n")
                fmt.Printf("Song %d, Channel %s: %d bytes, %d / %d ticks\n", sng.GetNum(), chn.GetName(), len(commands), utils.Round2(float64(chn.GetTicks())), utils.Round2(float64(chn.GetLoopTicks())))
            }
        }

        outFile.WriteString("\n.globl xpmp_song_tbl")
        outFile.WriteString("\nxpmp_song_tbl:\n")
        for n, sng := range songs {
            channels := sng.GetChannels()
            for _, chn := range channels { 
                if chn.IsVirtual() {
                    continue
                }
                outFile.WriteString(fmt.Sprintf("dc.w xpmp_s%d_channel_%s\n", n, chn.GetName()))
                songDataSize += 2       // ToDo: should be += 4 like in the original code?
            }
        }
        outFile.WriteString("dc.w 0\n")
        songDataSize += 2
    }
    
    return songDataSize
}


/* Outputs a zero-terminated string with a total size of exactLength bytes.
 * The string will be padded if it's too short, and truncated if it's too
 * long.
 */
func outputStringWithExactLength(outFile *os.File, str string, exactLength int) {
    if len(str) >= exactLength {
        outFile.WriteString(".db \"" + str[:exactLength-1] + "\", 0\n")
    } else {
        outFile.WriteString(".db \"" + str + "\"")
        for i := 0; i < exactLength - len(str); i++ {
            outFile.WriteString(", 0")
        }
        outFile.WriteString("\n")
    }
}


/* Outputs the tables for the standard effects (the ones that are common for most/all targets).
 */
func outputStandardEffects(outFile *os.File, outputFormat int) int {
    tableSize := outputTable(outFile, outputFormat, "xpmp_dt_mac", effects.DutyMacros,   true,  1, 0x80)
    tableSize += outputTable(outFile, outputFormat, "xpmp_v_mac",  effects.VolumeMacros, true,  1, 0x80)
    tableSize += outputTable(outFile, outputFormat, "xpmp_EP_mac", effects.PitchMacros,  true,  1, 0x80)
    tableSize += outputTable(outFile, outputFormat, "xpmp_EN_mac", effects.Arpeggios,    true,  1, 0x80)
    tableSize += outputTable(outFile, outputFormat, "xpmp_MP_mac", effects.Vibratos,     false, 1, 0x80)
    tableSize += outputTable(outFile, outputFormat, "xpmp_CS_mac", effects.PanMacros,    true,  1, 0x80)
    return tableSize
}


func (t *Target) outputCallbacks(outFile *os.File, outputFormat int) int {
    callbacksSize := 0

    switch outputFormat {
    case FORMAT_WLA_DX:
        outFile.WriteString("xpmp_callback_tbl:\n")
        for _, cb := range t.CompilerItf.GetCallbacks() {
            outFile.WriteString(".dw " + cb + "\n")
            callbacksSize += 2
        }
        outFile.WriteString("\n")
    }

    utils.INFO("Size of callback table: %d bytes", callbacksSize)  
    
    return callbacksSize
}


func (t *Target) outputEffectFlags(outFile *os.File, outputFormat int) {
    switch outputFormat {
    case FORMAT_WLA_DX:
        songs := t.CompilerItf.GetSongs()
        numChannels := len(songs[0].GetChannels())
        for _, effName := range EFFECT_STRINGS {
            for c := 0; c < numChannels; c++ {
                for _, sng := range songs {
                    channels := sng.GetChannels()
                    if channels[c].IsUsingEffect(effName) {
                        outFile.WriteString(fmt.Sprintf(".DEFINE XPMP_CHN%d_USES_", channels[c].GetNum()) + effName + "\n")
                        break
                    }
                }
            }
        }
    }
}


func outputTable(outFile *os.File, outputFormat int, tblName string, effMap *effects.EffectMap, canLoop bool, scaling int, loopDelim int) int {
    var bytesWritten, dat int
    
    bytesWritten = 0
    
    hexPrefix := "$"
    byteDecl := ".db"
    wordDecl := ".dw"

    switch outputFormat {
    case FORMAT_GAS_68K:
        hexPrefix = "0x"
        byteDecl = "dc.b"
        wordDecl = "dc.w"
    }        
    
    if effMap.Len() > 0 {
        for _, key := range effMap.GetKeys() {
            outFile.WriteString(fmt.Sprintf(tblName + "_%d:", key))
            effectData := effMap.GetData(key)
            for j, param := range effectData.MainPart {
                dat = (param.(int) * scaling) & 0xFF
                if canLoop && (dat == loopDelim) {
                    dat++
                }

                if canLoop && j == len(effectData.MainPart)-1 && len(effectData.LoopedPart) == 0 {
                    if j > 0 {
                        outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, loopDelim))
                    }
                    outFile.WriteString(fmt.Sprintf("\n" + tblName + "_%d_loop:\n", key))
                    outFile.WriteString(fmt.Sprintf("%s %s%02x, %s%02x", byteDecl, hexPrefix, dat, hexPrefix, loopDelim))
                    bytesWritten += 3
                } else if j == 0 {
                    outFile.WriteString(fmt.Sprintf("\n%s %s%02x", byteDecl, hexPrefix, dat))
                    bytesWritten += 1
                } else {
                    outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, dat))
                    bytesWritten += 1
                }
            }
            if canLoop && len(effectData.LoopedPart) > 0 {
                if len(effectData.MainPart) > 0 {
                    outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, loopDelim))
                    bytesWritten += 1
                }
                outFile.WriteString(fmt.Sprintf("\n" + tblName + "_%d_loop:\n", key))
                for j, param := range effectData.LoopedPart {
                    dat = (param.(int) * scaling) & 0xFF
                    if dat == loopDelim && canLoop {
                        dat++
                    }
                    if j == 0 {
                        outFile.WriteString(fmt.Sprintf("%s %s%02x", byteDecl, hexPrefix, dat))
                    } else {
                        outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, dat))
                    }
                    bytesWritten += 1
                }
                outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, loopDelim))
                bytesWritten += 1
            }
            outFile.WriteString("\n")
        }
        outFile.WriteString(tblName + "_tbl:\n")
        for _, key := range effMap.GetKeys() {
            outFile.WriteString(fmt.Sprintf("%s " + tblName + "_%d\n", wordDecl, key))
            bytesWritten += 2
        }
        if canLoop {
            outFile.WriteString(tblName + "_loop_tbl:\n")
            for _, key := range effMap.GetKeys() {
                outFile.WriteString(fmt.Sprintf("%s " + tblName + "_%d_loop\n", wordDecl, key))
                bytesWritten += 2
            }
        }
        outFile.WriteString("\n")
    } else {
        outFile.WriteString(tblName + "_tbl:\n")
        if canLoop {
            outFile.WriteString(tblName + "_loop_tbl:\n")
        }
        outFile.WriteString("\n")
    }
        
    return bytesWritten
}



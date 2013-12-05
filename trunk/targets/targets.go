/*
 * Package targets
 *
 * Part of XPMC.
 * Contains data/functions describing different output targets.
 *
 * /Mic, 2012-2013
 */
 
package targets

import (
    "fmt"
    "os"
    "time"
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


func NewTarget(tID int, icomp ICompiler) ITarget {
    var t ITarget = ITarget(nil)
    
    switch tID {
    case TARGET_AT8:
        t = &TargetAt8{}
    case TARGET_GBC:
        t = &TargetGBC{}
    case TARGET_SMD:
        t = &TargetGen{}
    case TARGET_KSS:
        t = &TargetKSS{}
    case TARGET_SGG:
        t = &TargetSGG{}
    case TARGET_SMS:
        t = &TargetSMS{}
    }
    
    if t != nil {
        t.SetCompilerItf(icomp)
    }
    
    return t
}

func NameToID(targetName string) int {
    switch targetName {
    case "at8":
        return TARGET_AT8;
    case "gbc":
        return TARGET_GBC;
    case "kss":
        return TARGET_KSS;
    case "sgg":
        return TARGET_SGG;
    case "smd":
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
    for i := range t.ChannelSpecs.Duty {
        names += fmt.Sprintf("%c", 'A'+i)
    }
    return names
}

func (t *Target) GetChannelSpecs() ISpecs {
    return &t.ChannelSpecs
}

/* Returns which of the specified chip's channels the given channel
 * number corresponds to for this target.
 */
func (t *Target) ChipChannel(chn, chipId int) int {
    for i, chnChipId := range t.ChannelSpecs.IDs {
        if chipId == chnChipId {
            return chn - i
        }
    }
    return -1
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


/* Atari 8-bit (XL/XE etc) *
 ***************************/

func (t *TargetAt8) Init() {
    utils.DefineSymbol("AT8", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsPokey)      // A..D
      
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
    utils.DefineSymbol("DMG", 1)
    utils.DefineSymbol("GBC", 1)
    
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

func (t *TargetGBC) Output(outputVgm int) {
    fmt.Printf("TargetGBC.Output\n")

    /*atom factor,f2,r2,smallestDiff
    integer f, tableSize, cbSize, songSize, wavSize, patSize, numSongs
    sequence closest,s*/

    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + ".asm")
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + ".asm")
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")

    songSize := 0
    /*numSongs = 0
    for i = 1 to length(songs) do
        if sequence(songs[i]) then
            numSongs += 1
        end if
    end for*/
    
    outFile.WriteString(
    ".IFDEF XPMP_MAKE_GBS\n\n" +
    ".MEMORYMAP\n" +
    "\tDEFAULTSLOT 1\n" +
    "\tSLOTSIZE $4000\n" +
    "\tSLOT 0 $0000\n" +
    "\tSLOT 1 $4000\n" +
    ".ENDME\n\n")
    outFile.WriteString(
    ".ROMBANKSIZE $4000\n" +
    ".ROMBANKS 2\n" +
    ".BANK 0 SLOT 0\n" +
    ".ORGA $00\n\n")
    
    outFile.WriteString(
    ".db \"GBS\"\n" +
    ".db 1\t\t; Version\n" +
    //sprintf(".db %d\t\t; Number of songs", numSongs) & CRLF &
    ".db 1\t\t; Start song\n" +
    ".dw $0400\t; Load address\n" +
    ".dw $0400\t; Init address\n" +
    ".dw $0408\t; Play address\n" +
    ".dw $fffe\t; Stack pointer\n" +
    ".db 0\n" +
    ".db 0\n")
    
    /*if length(songTitle) >= 32 then
        puts(outFile, ".db \"" & songTitle[1..31] & "\", 0" & CRLF)
    else
        puts(outFile, ".db \"" & songTitle & "\"")
        for i = 1 to 32 - length(songTitle) do
            puts(outFile, ", 0")
        end for
        puts(outFile, CRLF)
    end if
    if length(songComposer) >= 32 then
        puts(outFile, ".db \"" & songComposer[1..31] & "\", 0" & CRLF)
    else
        puts(outFile, ".db \"" & songComposer & "\"")
        for i = 1 to 32 - length(songComposer) do
            puts(outFile, ", 0")
        end for
        puts(outFile, CRLF)
    end if
    if length(songProgrammer) >= 32 then
        puts(outFile, ".db \"" & songProgrammer[1..31] & "\", 0" & CRLF)
    else
        puts(outFile, ".db \"" & songProgrammer & "\"")
        for i = 1 to 32 - length(songProgrammer) do
            puts(outFile, ", 0")
        end for
        puts(outFile, CRLF)
    end if  
    puts(outFile, ".INCBIN \"gbs.bin\"" & CRLF & CRLF)
    puts(outFile, ".ELSE" & CRLF & CRLF) 


    if sum(assoc_get_references(volumeMacros)) = 0 then 
        puts(outFile, ".DEFINE XPMP_VMAC_NOT_USED" & CRLF)
    end if
    if sum(assoc_get_references(pitchMacros)) = 0 then
        puts(outFile, ".DEFINE XPMP_EPMAC_NOT_USED" & CRLF)
    end if
    if sum(assoc_get_references(vibratos)) = 0 then
        puts(outFile, ".DEFINE XPMP_MPMAC_NOT_USED" & CRLF)
    end if
    if not usesEN[1] then
        puts(outFile, ".DEFINE XPMP_ENMAC_NOT_USED" & CRLF)
    end if
    if not usesEN[2] then
        puts(outFile, ".DEFINE XPMP_EN2MAC_NOT_USED" & CRLF)
    end if*/

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
    
    if t.CompilerItf.GetGbNoiseType() == 1 {
        outFile.WriteString(".DEFINE XPMP_ALT_GB_NOISE\n")
    }
    if t.CompilerItf.GetGbVolCtrlType() == 1 {
        outFile.WriteString(".DEFINE XPMP_ALT_GB_VOLCTRL\n")
    }
    
    tableSize := outputWlaTable(outFile, "xpmp_dt_mac", effects.DutyMacros,   true, 1, 0x80)
    tableSize += outputWlaTable(outFile, "xpmp_v_mac",  effects.VolumeMacros, true, 1, 0x80)
    //tableSize += outputWlaTable(outFile, "xpmp_VS_mac", effects.VolumeSlides, true, 1, 0x80)
    tableSize += outputWlaTable(outFile, "xpmp_EP_mac", effects.PitchMacros,  true, 1, 0x80)
    tableSize += outputWlaTable(outFile, "xpmp_EN_mac", effects.Arpeggios,    true, 1, 0x80)
    tableSize += outputWlaTable(outFile, "xpmp_MP_mac", effects.Vibratos,     false, 1, 0x80)
    tableSize += outputWlaTable(outFile, "xpmp_CS_mac", effects.PanMacros,    true, 1, 0x80)
    
    /*tableSize += output_wla_table("xpmp_WT_mac", waveformMacros, 1, 1, #80)
    
    wavSize = 0
    puts(outFile, "xpmp_waveform_data:")
    for i = 1 to length(waveforms[ASSOC_KEY]) do
        for j = 1 to length(waveforms[ASSOC_DATA][i][LIST_MAIN]) by 2 do
            if j = 1 then
                puts(outFile, CRLF & ".db ")
            end if              
            printf(outFile, "$%02x", waveforms[ASSOC_DATA][i][LIST_MAIN][j]*#10 + waveforms[ASSOC_DATA][i][LIST_MAIN][j+1])
            wavSize += 1
            if j < length(waveforms[ASSOC_DATA][i][LIST_MAIN]) - 1 then
                puts(outFile, ",")
            end if

        end for
    end for
    puts(outFile, {13, 10, 13, 10})
    
    cbSize = 0
    puts(outFile, "xpmp_callback_tbl:" & CRLF)
    for i = 1 to length(callbacks) do
        puts(outFile, ".dw " & callbacks[i] & CRLF)
        cbSize += 2
    end for
    puts(outFile, CRLF)*/

    utils.INFO(fmt.Sprintf("Size of effect tables: %d bytes\n", tableSize))
    //INFO(fmt.Sprintf("Size of waveform table: %d bytes\n", wavSize))

    /*patSize = 0
    for n = 1 to length(patterns[2]) do
        printf(outFile, "xpmp_pattern%d:", n)
        for j = 1 to length(patterns[2][n]) do
            if remainder(j, 16) = 1 then
                puts(outFile, CRLF & ".db ")
            end if              
            printf(outFile, "$%02x", and_bits(patterns[2][n][j], #FF))
            if j < length(patterns[2][n]) and remainder(j, 16) != 0 then
                puts(outFile, ",")
            end if

        end for
        puts(outFile, CRLF)
        patSize += length(patterns[2][n])
    end for

    puts(outFile, {13, 10} & "xpmp_pattern_tbl:" & CRLF)
    for n = 1 to length(patterns[2]) do
        printf(outFile, ".dw xpmp_pattern%d" & {13, 10}, n)
        patSize += 2
    end for
    puts(outFile, {13, 10})

    if verbose then
        printf(1, "Size of patterns table: %d bytes\n", patSize)
    end if*/
  
    for n, sng := range songs {
        channels := sng.GetChannels()
        for _, chn := range channels {  // ToDo: don't iterate over the last channel (pattern)
            outFile.WriteString(fmt.Sprintf("xpmp_s%d_channel_%s:", n, chn.GetName()))
            commands := chn.GetCommands()
            for j, cmd := range commands {
                if (j % 16) == 0 {
                    outFile.WriteString("\n.db ")
                }
                outFile.WriteString(fmt.Sprintf("$%02x", cmd & 0xFF))
                songSize++
                if j < len(commands)-1 && (j % 16) != 15 {
                   outFile.WriteString(",")
                }
            }
            outFile.WriteString("\n")
            fmt.Printf("Song %d, Channel %s: %d bytes, %d / %d ticks\n", n, chn.GetName(), len(commands), 0, 0)  // ToDo: replace zeroes with: round2(songLen[n][i]), round2(songLoopLen[n][i])})
        }
    }
    
    outFile.WriteString("\nxpmp_song_tbl:\n")
    for n, sng := range songs {
        channels := sng.GetChannels()
        for _, chn := range channels {  // ToDo: don't iterate over the last channel (pattern)
            outFile.WriteString(fmt.Sprintf(".dw xpmp_s%d_channel_%s\n", n, chn.GetName()))
            songSize += 2
        }
    }

    utils.INFO(fmt.Sprintf("Total size of song(s): %d bytes\n", songSize + tableSize)) // ToDo: + patSize + cbSize + wavSize)
    
    outFile.WriteString(".ENDIF")
    outFile.Close()
}


/* Sega Genesis / Megadrive *
 ****************************/

func (t *TargetGen) Init() {
    utils.DefineSymbol("GEN", 1)
    utils.DefineSymbol("SMD", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 4, specs.SpecsYM2612)     // E..J

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
    utils.DefineSymbol("KSS", 1)
    
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
    utils.DefineSymbol("NES", 1)       
    
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
    utils.DefineSymbol("NGP", 1)
    
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
    utils.DefineSymbol("SGG", 1)
    
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
    utils.DefineSymbol("SMS", 1)       
    
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



// Output table data in WLA-DX format 
func outputWlaTable(outFile *os.File, tblName string, effMap *effects.EffectMap, canLoop bool, scaling int, loopDelim int) int {
    var bytesWritten, dat int
    
    bytesWritten = 0
    
    if effMap.Len() > 0 {
        for _, key := range effMap.GetKeys() {
            outFile.WriteString(fmt.Sprintf(tblName + "_%d:", key))
            effectData := effMap.GetData(key)
            for j, param := range effectData.MainPart {
                dat = (param * scaling) & 0xFF
                if canLoop && (dat == loopDelim) {
                    dat++
                }

                if canLoop && j == len(effectData.MainPart)-1 && len(effectData.LoopedPart) == 0 {
                    if j > 0 {
                        outFile.WriteString(fmt.Sprintf(", $%02x", loopDelim))
                    }
                    outFile.WriteString(fmt.Sprintf("\n" + tblName + "_%d_loop:\n", key))
                    outFile.WriteString(fmt.Sprintf(".db $%02x, $%02x", dat, loopDelim))
                    bytesWritten += 3
                } else if j == 0 {
                    outFile.WriteString(fmt.Sprintf("\n.db $%02x", dat))
                    bytesWritten += 1
                } else {
                    outFile.WriteString(fmt.Sprintf(", $%02x", dat))
                    bytesWritten += 1
                }
            }
            if canLoop && len(effectData.LoopedPart) > 0 {
                if len(effectData.MainPart) > 0 {
                    outFile.WriteString(fmt.Sprintf(", $%02x", loopDelim))
                    bytesWritten += 1
                }
                outFile.WriteString(fmt.Sprintf("\n" + tblName + "_%d_loop:\n", key))
                for j, param := range effectData.LoopedPart {
                    dat = (param * scaling) & 0xFF
                    if dat == loopDelim && canLoop {
                        dat++
                    }
                    if j == 0 {
                        outFile.WriteString(fmt.Sprintf(".db $%02x", dat))
                    } else {
                        outFile.WriteString(fmt.Sprintf(", $%02x", dat))
                    }
                    bytesWritten += 1
                }
                outFile.WriteString(fmt.Sprintf(", $%02x", loopDelim))
                bytesWritten += 1
            }
            outFile.WriteString("\n")
        }
        outFile.WriteString(tblName + "_tbl:\n")
        for _, key := range effMap.GetKeys() {
            outFile.WriteString(fmt.Sprintf(".dw " + tblName + "_%d\n", key))
            bytesWritten += 2
        }
        if canLoop {
            outFile.WriteString(tblName + "_loop_tbl:\n")
            for _, key := range effMap.GetKeys() {
                outFile.WriteString(fmt.Sprintf(".dw " + tblName + "_%d_loop\n", key))
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



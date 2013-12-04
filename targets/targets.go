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
    "../specs"
    "../utils"
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

    /*s = date()
    printf(outFile, "; Written by XPMC at %02d:%02d:%02d on " & WEEKDAYS[s[7]] & " " & MONTHS[s[2]] & " %d, %d." & {13, 10, 13, 10},
           s[4..6] & {s[3], s[1] + 1900})

    numSongs = 0
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
    
    /*puts(outFile,
    ".db \"GBS\"" & CRLF &
    ".db 1\t\t; Version" & CRLF &
    sprintf(".db %d\t\t; Number of songs", numSongs) & CRLF &
    ".db 1\t\t; Start song" & CRLF &
    ".dw $0400\t; Load address" & CRLF &
    ".dw $0400\t; Init address" & CRLF &
    ".dw $0408\t; Play address" & CRLF &
    ".dw $fffe\t; Stack pointer" & CRLF &
    ".db 0" & CRLF &
    ".db 0" & CRLF)
    
    if length(songTitle) >= 32 then
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
    end if

    for i = 1 to length(supportedChannels)-1 do
        for j = 1 to length(usesEffect[i]) do
            if usesEffect[i][j] then
                printf(outFile, ".DEFINE XPMP_CHN%d_USES_%s" & CRLF, {i - 1, EFFECT_STRINGS[j]})
            end if
        end for
    end for
    
    if gbNoise = 1 then
        puts(outFile, ".DEFINE XPMP_ALT_GB_NOISE" & CRLF)
    end if
    if gbVolCtrl = 1 then
        puts(outFile, ".DEFINE XPMP_ALT_GB_VOLCTRL" & CRLF)
    end if
    
    tableSize  = output_wla_table("xpmp_dt_mac", dutyMacros,   1, 1, #80)
    tableSize += output_wla_table("xpmp_v_mac",  volumeMacros, 1, 1, #80)
    tableSize += output_wla_table("xpmp_VS_mac", volumeSlides, 1, 1, #80)
    tableSize += output_wla_table("xpmp_EP_mac", pitchMacros,  1, 1, #80)
    tableSize += output_wla_table("xpmp_EN_mac", arpeggios,    1, 1, #80)
    tableSize += output_wla_table("xpmp_MP_mac", vibratos,     0, 1, #80)
    tableSize += output_wla_table("xpmp_CS_mac", panMacros,    1, 1, #80)
    tableSize += output_wla_table("xpmp_WT_mac", waveformMacros, 1, 1, #80)
    
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
    puts(outFile, CRLF)

    if verbose then
        printf(1, "Size of effect tables: %d bytes\n", tableSize)
    end if
    if verbose then
        printf(1, "Size of waveform table: %d bytes\n", wavSize)
    end if

    patSize = 0
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
    end if
    
    songSize = 0
    for n = 1 to length(songs) do
        if sequence(songs[n]) then
            for i = 1 to length(supportedChannels)-1 do
                printf(outFile, "xpmp_s%d_channel_" & supportedChannels[i] & ":", n)
                for j = 1 to length(songs[n][i]) do
                    if remainder(j, 16) = 1 then
                        puts(outFile, CRLF & ".db ")
                    end if              
                    printf(outFile, "$%02x", and_bits(songs[n][i][j], #FF))
                    songSize += 1
                    if j < length(songs[n][i]) and remainder(j, 16) != 0 then
                        puts(outFile, ",")
                    end if

                end for
                puts(outFile, CRLF)
                printf(1, "Song %d, Channel " & supportedChannels[i] & ": %d bytes, %d / %d ticks\n", {n, length(songs[n][i]), round2(songLen[n][i]), round2(songLoopLen[n][i])})
            end for
        end if
    end for
    
    puts(outFile, {13, 10} & "xpmp_song_tbl:" & CRLF)
    for n = 1 to length(songs) do
        if sequence(songs[n]) then
            for i = 1 to length(supportedChannels)-1 do
                printf(outFile, ".dw xpmp_s%d_channel_" & supportedChannels[i] & CRLF, n)
                songSize += 2
            end for
        end if
    end for

    if verbose then
        printf(1, "Total size of song(s): %d bytes\n", songSize + patSize + tableSize + cbSize + wavSize)
    end if*/
    
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



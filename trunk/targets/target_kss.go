/*
 * Package targets
 * Target KSS (MSX)
 *
 * Part of XPMC.
 * Contains data/functions specific to the KSS output target
 *
 * /Mic, 2012-2014
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


/* KSS (MSX executable music) *
 ******************************/

func (t *TargetKSS) Init() {
    utils.DefineSymbol("KSS", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 4, specs.SpecsAY_3_8910)  // E..G
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 7, specs.SpecsSCC)        // H..L
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 12, specs.SpecsYM2151)    // M..T
    
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


func (t *TargetKSS) Output(outputVgm int) {
    fmt.Printf("TargetKSS.Output\n")

    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + ".asm")
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + ".asm")
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")
  
    usesPSG := 0
    usesSCC := 0
    extraChips := 0
    songs := t.CompilerItf.GetSongs()
    for _, sng := range songs {
        channels := sng.GetChannels()
        for _, chn := range channels {
            if chn.IsUsed() {
                switch chn.GetChipID() {
                case specs.CHIP_AY_3_8910:
                    usesPSG = 1
                case specs.CHIP_SCC:
                    usesSCC = 1
                case specs.CHIP_SN76489:
                    extraChips |= 6
                case specs.CHIP_YM2151:
                    extraChips |= 3
                }
            }
        }
    }
    
    envelopes := make([][]int, len(effects.ADSRs.GetKeys()))
    for i, key := range effects.ADSRs.GetKeys() {
        envelopes[i] = packADSR(effects.ADSRs.GetData(key).MainPart, specs.CHIP_YM2151)
    }
    
    mods := make([][]int, len(effects.MODs.GetKeys()))
    for i, key := range effects.MODs.GetKeys() {
        mods[i] = packMOD(effects.MODs.GetData(key).MainPart, specs.CHIP_YM2151)
    }
    
    outFile.WriteString( 
        ".IFDEF XPMP_MAKE_KSS\n" +
        ".memorymap\n" +
        "defaultslot 0\n" +
        "slotsize $8010\n" +
        "slot 0 0\n" +
        ".endme\n\n" +
        ".rombanksize $8010\n" +
        ".rombanks 1\n\n" +
        ".orga   $0000\n" +
        ".db      \"KSCC\"   ; Magic string\n" +
        ".dw   $0000   ; Load address\n" +
        ".dw   $8000   ; Data length\n" +
        ".dw   $7FF0   ; Driver initialize function\n" +
        ".dw   $0000   ; Play address\n" +
        ".db   $00   ; No. of banks\n" +
        ".db   $00   ; extra\n" +
        ".db   $00   ; reserved\n")

    outFile.WriteString(fmt.Sprintf(
        ".db   $%02x   ; Extra chips\n\n" +
        ".incbin \"kss.bin\"\n\n" +
        ".ELSE\n\n", extraChips))
    
    t.outputEffectFlags(outFile, FORMAT_WLA_DX)

    if usesPSG != 0 {
        outFile.WriteString(".DEFINE XPMP_USES_AY\n")
    }
    if usesSCC != 0 {
        outFile.WriteString(".DEFINE XPMP_USES_SCC\n")
    }
    if (extraChips & 2) == 2 {
        if (extraChips & 1) == 1 {
            outFile.WriteString(".DEFINE XPMP_USES_FMUNIT\n")
        } else {            
            outFile.WriteString(".DEFINE XPMP_USES_SN76489\n")
        }
    }
    
    tableSize := outputStandardEffects(outFile, FORMAT_WLA_DX)
    tableSize += outputTable(outFile, FORMAT_WLA_DX, "xpmp_FB_mac", effects.FeedbackMacros, true,  1, 0x80)
    tableSize += outputTable(outFile, FORMAT_WLA_DX, "xpmp_WT_mac", effects.WaveformMacros, true,  1, 0x80)
    tableSize += outputTable(outFile, FORMAT_WLA_DX, "xpmp_ADSR",   effects.ADSRs,          false, 1, 0)    // ToDo: use packed envelopes
    tableSize += outputTable(outFile, FORMAT_WLA_DX, "xpmp_MOD",    effects.MODs,           false, 1, 0)    // ToDo: use packed mods
    
    patSize := t.outputPatterns(outFile, FORMAT_WLA_DX)
    utils.INFO("Size of patterns table: %d bytes\n", patSize)

    songSize := t.outputChannelData(outFile, FORMAT_WLA_DX)
    utils.INFO("Total size of song(s): %d bytes\n", songSize + tableSize + patSize) // + cbSize + wavSize)
    
    outFile.Close()
}


/*
 * Package targets
 * Target GBC
 *
 * Part of XPMC.
 * Contains data/functions specific to the GBC output target
 *
 * /Mic, 2012-2015
 */
 
package targets

import (
    "fmt"
    "os"
    "strconv"
    "time"
    "../defs"
    "../specs"
    "../utils"
    "../effects"
)

import . "../utils"

/* Gameboy Color (and DMG) 
 */
func (t *TargetGBC) Init() {
    t.Target.Init()
    t.Target.SetOutputSyntax(SYNTAX_WLA_DX)
    
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
    
    t.CompilerItf.SetMetaCommandHandler("GB-VOLUME-CONTROL", handleGbVolCtrl)
    t.CompilerItf.SetMetaCommandHandler("GB-NOISE", handleGbNoiseCtrl)
}

    
func (t *TargetGBC) Output(outputFormat int) {
    utils.DEBUG("TargetGBC.Output")

    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + ".asm")
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + ".asm")
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")
    
    // Output the GBS header
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
    fmt.Sprintf(".db %d\t\t; Number of songs\n", len(t.CompilerItf.GetSongs())) +
    ".db 1\t\t; Start song\n" +
    ".dw $0400\t; Load address\n" +
    ".dw $0400\t; Init address\n" +
    ".dw $0408\t; Play address\n" +
    ".dw $fffe\t; Stack pointer\n" +
    ".db 0\n" +
    ".db 0\n")
    
    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetTitle(), 32)
    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetComposer(), 32)
    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetProgrammer(), 32)   

    outFile.WriteString(".INCBIN \"gbs.bin\"\n\n")
    outFile.WriteString(".ELSE\n\n") 

    t.outputEffectFlags(outFile)
    
    if t.GetExtraInt("NoiseCtrl", 0) == 1 { //t.CompilerItf.GetGbNoiseType() == 1 {
        outFile.WriteString(".DEFINE XPMP_ALT_GB_NOISE\n")
    }
    if t.GetExtraInt("VolCtrl", 0) == 1 { //t.CompilerItf.GetGbVolCtrlType() == 1 {
        outFile.WriteString(".DEFINE XPMP_ALT_GB_VOLCTRL\n")
    }
    
    tableSize := t.outputStandardEffects(outFile)
    
    // ToDo: output waveform macros (WTM)
    /*tableSize += output_wla_table("xpmp_WT_mac", waveformMacros, 1, 1, #80)*/
    
    wavSize := 0
    outFile.WriteString("xpmp_waveform_data:")
    for _, key := range effects.Waveforms.GetKeys() {
        params := effects.Waveforms.GetData(key).MainPart
        for j := 0; j < len(params); j += 2 {
            if j == 0 {
                outFile.WriteString("\n.db ")
            }             
            // Pack two 4-bit samples into one byte
            outFile.WriteString(fmt.Sprintf("$%02x", params[j].(int) * 0x10 + params[j+1].(int)))
            wavSize++
            if j < len(params) - 1 {
                outFile.WriteString(",")
            }
        }
    }
    outFile.WriteString("\n\n")
    
    cbSize := 0
    outFile.WriteString("xpmp_callback_tbl:\n")
    for _, cb := range t.CompilerItf.GetCallbacks() {
        outFile.WriteString(".dw " + cb + "\n")
        cbSize += 2
    }
    outFile.WriteString("\n")

    utils.INFO("Size of effect tables: %d bytes", tableSize)
    utils.INFO("Size of waveform table: %d bytes", wavSize)

    patSize := t.outputPatterns(outFile)
    utils.INFO("Size of pattern table: %d bytes\n", patSize)
  
    songSize := t.outputChannelData(outFile)
    utils.INFO("Total size of song(s): %d bytes\n", songSize + tableSize + wavSize + cbSize + patSize )
    
    outFile.WriteString(".ENDIF")
    outFile.Close()
}


func handleGbVolCtrl(cmd string, itarget defs.ITarget) {
    s := Parser.GetString()
    ctl, err := strconv.Atoi(s)
    if err == nil {
        if ctl == 1 {
            itarget.PutExtraInt("VolCtrl", 1)
        } else if ctl == 0 {
            itarget.PutExtraInt("VolCtrl", 0)
        } else {
            WARNING(cmd + ": Expected 0 or 1, got: " + s)
        }
    } else {
        ERROR(cmd + ": Expected 0 or 1, got: " + s)
    }
}

func handleGbNoiseCtrl(cmd string, itarget defs.ITarget) {
    s := Parser.GetString()
    val, err := strconv.Atoi(s)
    if err == nil {
        if val == 1 {
            itarget.PutExtraInt("NoiseCtrl", 1)
            itarget.GetCompilerItf().GetCurrentSong().GetChannels()[3].SetMaxOctave(5)
        } else if val == 0 {
            itarget.PutExtraInt("NoiseCtrl", 0)
            itarget.GetCompilerItf().GetCurrentSong().GetChannels()[3].SetMaxOctave(11)
        } else {
            WARNING(cmd + ": Expected 0 or 1, defaulting to 0: " + s)
        }
    } else {
        ERROR(cmd + ": Expected 0 or 1, got: " + s)
    }
}


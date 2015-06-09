/*
 * Package targets
 * Target SMS (Sega Master System)
 *
 * Part of XPMC.
 * Contains data/functions specific to the SMS output target
 *
 * /Mic, 2012-2015
 */
 
package targets

import (
    "fmt"
    "os"
    "time"
    "../specs"
    "../utils"
    "../effects"
    "../timing"
)


func (t *TargetSMS) Init() {
    t.Target.Init()
    t.Target.SetOutputSyntax(SYNTAX_WLA_DX)

    utils.DefineSymbol("SMS", 1)       
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 4, specs.SpecsYM2413)     // E..M
    
    //activeChannels        = repeat(0, length(supportedChannels))  
    
    t.ID                = TARGET_SMS
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.MaxLoopDepth      = 2
    t.MachineSpeed      = 3579545
    t.AdsrLen           = 4
    t.AdsrMax           = 15
    t.SupportsPal       = true
}


func (t *TargetSMS) Output(outputFormat int) {
    utils.DEBUG("TargetSMS.Output")

    fileEnding := ".asm"
    outputVgm := false
    if outputFormat == OUTPUT_VGM {
        fileEnding = ".vgm"
        outputVgm = true
    } else if outputFormat == OUTPUT_VGZ {
        fileEnding = ".vgz"
        outputVgm = true
    }

    if outputVgm {
        // ToDo: output VGM/VGZ
        return
    }

    envelopes := make([][]interface{}, len(effects.ADSRs.GetKeys()))
    for i, key := range effects.ADSRs.GetKeys() {
        envelopes[i] = packADSR(effects.ADSRs.GetData(key).MainPart, specs.CHIP_YM2413)
        effects.ADSRs.GetData(key).MainPart = make([]interface{}, len(envelopes[i]))
        copy(effects.ADSRs.GetData(key).MainPart, envelopes[i])        
    }
    
    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + fileEnding)
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + fileEnding)
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")

    palNtscString := ".db 0"
    if timing.UpdateFreq == 50 {
        outFile.WriteString(".DEFINE XPMP_50_HZ\n")
        t.MachineSpeed = 3546893
        palNtscString = ".db 1"
    } else {
        t.MachineSpeed = 3579545
    }

    if t.CompilerItf.GetSongs()[0].GetSmsTuning() {
        outFile.WriteString(".DEFINE XPMP_TUNE_SMS\n")
    }

    // Output the SGC header
    outFile.WriteString(
    ".IFDEF XPMP_MAKE_SGC\n\n" +
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

    systemType := 0     // SMS
    if !t.SupportsPal {
        systemType = 1  // GG
    }
    outFile.WriteString(
    ".db \"SGC\"\n" +
    ".db $1A\n" +
    ".db 1\t\t; Version\n" +
    palNtscString + "\n" + 
    ".db 0, 0\n" +
    ".dw $0400\t; Load address\n" +
    ".dw $0400\t; Init address\n" +
    ".dw $0408\t; Play address\n" +
    ".dw $dff0\t; Stack pointer\n" +
    ".dw 0\t\t; Reserved\n" +
    ".dw $040C\t; RST 08\n" +
    ".dw $040C\t; RST 10\n" +
    ".dw $040C\t; RST 18\n" +
    ".dw $040C\t; RST 20\n" +
    ".dw $040C\t; RST 28\n" +
    ".dw $040C\t; RST 30\n" +
    ".dw $040C\t; RST 38\n" +
    ".db 0, 0, 1, 2\t; Mapper setting (none)\n" +
    ".db 0\t\t; Start song\n" +
    fmt.Sprintf(".db %d\t\t; Number of songs\n", len(t.CompilerItf.GetSongs())) +
    ".db 0, 0\t; Sound effects (none)\n" +
    fmt.Sprintf(".db %d\t\t; System type\n", systemType) +
    ".dw 0,0,0,0,0,0,0,0,0,0,0 ; Reserved\n" +
    ".db 0")

    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetTitle(), 32)
    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetComposer(), 32)
    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetProgrammer(), 32)  
    
    outFile.WriteString(".INCBIN \"sgc.bin\"\n\n")
    outFile.WriteString(".ELSE\n\n") 
    
    t.outputEffectFlags(outFile)
    
    usesFM := false
    songs := t.CompilerItf.GetSongs()
    for _, sng := range songs {
        usesFM = sng.UsesChip(specs.CHIP_YM2413)
        if usesFM {
            break
        }
    }
    if usesFM {
        outFile.WriteString(".DEFINE XPMP_ENABLE_FM\n")
    }
        
    tableSize := t.outputStandardEffects(outFile)
    if usesFM {
        tableSize += t.outputTable(outFile, "xpmp_ADSR",   effects.ADSRs, false, 1, 0)  
    }    
    /*outFile.WriteString("xpmp_ADSR_tbl:\n")
    if usesFM {
        for _, enve := range envelopes {
            outFile.WriteString(fmt.Sprintf(".db $%02x,$%02x\n", enve[0], enve[1]))
            tableSize += 2
        }
    }*/
    outFile.WriteString("\n")
    utils.INFO("Size of effect tables: %d bytes", tableSize)

    cbSize := t.outputCallbacks(outFile)
        
    patSize := t.outputPatterns(outFile)
    utils.INFO("Size of patterns table: %d bytes\n", patSize)
        
    songSize := t.outputChannelData(outFile)  
    utils.INFO("Total size of song(s): %d bytes", songSize + tableSize + cbSize + patSize)

    outFile.Close()
}

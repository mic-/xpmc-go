/*
 * Package targets
 * Target SMS (Sega Game Gear)
 *
 * Part of XPMC.
 * Contains data/functions specific to the SGG output target
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
)


/* Sega Gamegear *
 *****************/

func (t *TargetSGG) Init() {
    t.Target.Init()
    t.Target.SetOutputSyntax(SYNTAX_WLA_DX)

    utils.DefineSymbol("SGG", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    
    t.ID                = TARGET_SGG
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.MachineSpeed      = 3579545
}


func (t *TargetSGG) Output(outputFormat int) {
    utils.DEBUG("TargetSGG.Output")

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
  
    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + fileEnding)
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + fileEnding)
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")

    outFile.WriteString(".DEFINE XPMP_GAME_GEAR\n")

    palNtscString := ".db 0"
    t.MachineSpeed = 3579545

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
         
    tableSize := t.outputStandardEffects(outFile)  
    outFile.WriteString("\n")
    utils.INFO("Size of effect tables: %d bytes", tableSize)

    patSize := t.outputPatterns(outFile)
    utils.INFO("Size of patterns table: %d bytes\n", patSize)
 
    songSize := t.outputChannelData(outFile) 
    utils.INFO("Total size of song(s): %d bytes", songSize + tableSize)

    outFile.Close()    
}

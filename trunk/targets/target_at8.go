package targets

import (
    "fmt"
    "os"
    "time"
    "../specs"
    "../utils"
)


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



/* Output data suitable for the Atari 8-bit (400/800/XE/XL) playback library (WLA-DX).
 */
func (t *TargetAt8) Output(outputVgm int) {
    fmt.Printf("TargetGBC.Output\n")

    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + ".asm")
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + ".asm")
    }

    saphdr, err := os.Create("sapheader.txt")
    if err != nil {
        utils.ERROR("Unable to create file: sapheader.txt");    
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")
   
    saphdr.WriteString("SAP\n" + 
        "AUTHOR \"" + t.CompilerItf.GetSongs()[0].GetProgrammer() + "\"\n" +
        "NAME \"" + t.CompilerItf.GetSongs()[0].GetTitle() + "\"\n" +
        fmt.Sprintf("DATE \"%d\"\n", now.Year()) +
        fmt.Sprintf("SONGS %d\n", len(t.CompilerItf.GetSongs())) +
        "DEFSONG 0\n" +
        "TYPE B\n" +
        "INIT 2000\n" +
        "PLAYER 2011\n")
    saphdr.Close()

    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetTitle(), 32)
    outputStringWithExactLength(outFile, t.CompilerItf.GetSongs()[0].GetComposer(), 32)
    
    outFile.WriteString(".DEFINE XPMP_AT8\n")
    
    t.outputEffectFlags(outFile, FORMAT_WLA_DX)
     
    tableSize := outputStandardEffects(outFile, FORMAT_WLA_DX)
    outFile.WriteString("\n")
    utils.INFO("Size of effect tables: %d bytes", tableSize)

    cbSize := t.outputCallbacks(outFile, FORMAT_WLA_DX)

    patSize := t.outputPatterns(outFile, FORMAT_WLA_DX)
    utils.INFO("Size of patterns table: %d bytes\n", patSize)
        
    songSize := t.outputChannelData(outFile, FORMAT_WLA_DX)  
    utils.INFO("Total size of song(s): %d bytes", songSize + tableSize + cbSize + patSize)

    outFile.Close()
}





/*
 * Package targets
 * Target AST (Atari ST)
 *
 * Part of XPMC.
 * Contains data/functions specific to the Atari ST output target
 *
 * /Mic, 2014
 */
 
package targets

import (
    "os"
    "time"
    "../specs"
    "../utils"
    "../timing"
)


func (t *TargetAST) Init() {
    utils.DefineSymbol("AST", 1) 
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsAY_3_8910)   // A..C
    
    t.ID                = TARGET_AST
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPal       = true
    t.MaxLoopDepth      = 2
    t.MachineSpeed      = 2000000
    timing.UpdateFreq   = 50.0  // Use PAL as default
}


/* Output data suitable for the Atari ST
 */
func (t *TargetAST) Output(outputVgm int) {
    utils.DEBUG("TargetAST.Output")

    fileEnding := ".s"
    if outputVgm == 1 {
        fileEnding = ".vgm"
    } else if outputVgm == 2 {
        fileEnding = ".vgz"
    } else if outputVgm == 3 {
        fileEnding = ".ym"
    }

    if outputVgm != 0 {
        // ToDo: output VGM/VGZ or YM
        return
    }
    
    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + fileEnding)
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + fileEnding)
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")    

    // ToDo: finish
    
    outFile.Close()
}

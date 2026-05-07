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
    "xpmc-go/specs"
    "xpmc-go/utils"
    "xpmc-go/timing"
)


func (t *TargetAST) Init() {
    t.Target.Init()
    t.Target.SetOutputSyntax(SYNTAX_GAS_68K)
    
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
func (t *TargetAST) Output(outputFormat int) {
    utils.DEBUG("TargetAST.Output")

    fileEnding := ".s"
    if outputFormat == OUTPUT_YM {
        fileEnding = ".ym"
    }

    if t.Target.OutputVGM(outputFormat) {
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

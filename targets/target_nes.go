/*
 * Package targets
 * Target NES
 *
 * Part of XPMC.
 * Contains data/functions specific to the NES output target
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


func (t *TargetNES) Init() {
    utils.DefineSymbol("NES", 1) 
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.Specs2A03)   // A..E
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 5, specs.SpecsVRC6)   // F..H
    
    t.ID                = TARGET_NES
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.SupportsPal       = true
    t.MaxLoopDepth      = 2
    t.MinWavLength      = 8
    t.MaxWavLength      = 32
    t.MinWavSample      = 0
    t.MaxWavSample      = 15
    timing.UpdateFreq   = 60.0  // Use NTSC as default
}


/* Output data suitable for the NES/Famicom (WLA-DX)
 */
func (t *TargetNES) Output(outputVgm int) {
    utils.DEBUG("TargetNES.Output")
    
    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + ".asm")
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + ".asm")
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")    

    // ToDo: finish
    
    outFile.Close()
}

    
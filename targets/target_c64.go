/*
 * Package targets
 * Target C64 (Commodore 64)
 *
 * Part of XPMC.
 * Contains data/functions specific to the C64 output target
 *
 * /Mic, 2012-2014
 */
 
package targets

import (
    "../specs"
    "../utils"
    "../timing"
)


func (t *TargetC64) Init() {
    utils.DefineSymbol("C64", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSID)    // A..C
    
    t.ID                = TARGET_C64
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.MaxLoopDepth      = 2
    t.SupportsPal       = true
    timing.UpdateFreq   = 50.0
}



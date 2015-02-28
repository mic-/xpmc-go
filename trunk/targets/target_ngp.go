package targets

import (
    "../specs"
    "../utils"
)


/* NeoGeo Pocket / Color *
 *************************/

func (t *TargetNGP) Init() {
    t.Target.Init()

    utils.DefineSymbol("NGP", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsT6W28)      // A..D
    
    t.ID                = TARGET_NGP
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.MachineSpeed      = 3072000
}

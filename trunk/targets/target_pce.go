package targets

import (
    "../specs"
    "../utils"
)


func (t *TargetPCE) Init() {
    utils.DefineSymbol("PCE", 1)
    utils.DefineSymbol("TGX", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsHuC6280)      // A..F
    
    t.ID                = TARGET_PCE
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.MinWavLength      = 32
    t.MaxWavLength      = 32
    t.MinWavSample      = 0
    t.MaxWavSample      = 31
}


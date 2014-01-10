/*
 * Package targets
 * Target GEN/SMD (Sega Genesis / Megadrive)
 *
 * Part of XPMC.
 * Contains data/functions specific to the GEN output target
 *
 * /Mic, 2012-2014
 */
 
package targets

import (
    "os"
    "time"
    "../specs"
    "../utils"
    "../effects"
    "../timing"
)


/* Sega Genesis / Megadrive *
 ****************************/

func (t *TargetGen) Init() {
    utils.DefineSymbol("GEN", 1)
    utils.DefineSymbol("SMD", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsSN76489)    // A..D
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 4, specs.SpecsYM2612)     // E..J

    t.ID                = TARGET_SMD
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.MaxLoopDepth      = 2
    t.SupportsPal       = true
    t.AdsrLen           = 5
    t.AdsrMax           = 63
    t.MinWavLength      = 1
    t.MaxWavLength      = 2097152 // 2MB
    t.MinWavSample      = 0
    t.MaxWavSample      = 255
    t.MachineSpeed      = 3579545
}


/* Output data suitable for the SEGA Genesis (Megadrive) playback library
 */
func (t *TargetGen) Output(outputVgm int) {
    fileEnding := ".asm"
    if outputVgm == 1 {
        fileEnding = ".vgm"
    } else if outputVgm == 2 {
        fileEnding = ".vgz"
    }

    if outputVgm != 0 {
        // ToDo: output VGM/VGZ
        return
    }
  
    outFile, err := os.Create(t.CompilerItf.GetShortFileName() + fileEnding)
    if err != nil {
        utils.ERROR("Unable to open file: " + t.CompilerItf.GetShortFileName() + fileEnding)
    }

    now := time.Now()
    outFile.WriteString("; Written by XPMC on " + now.Format(time.RFC1123) + "\n\n")
    
    // Convert ADSR envelopes to the format used by the YM2612
    envelopes := make([][]interface{}, len(effects.ADSRs.GetKeys()))
    for i, key := range effects.ADSRs.GetKeys() {
        envelopes[i] = packADSR(effects.ADSRs.GetData(key).MainPart, specs.CHIP_YM2612)
        effects.ADSRs.GetData(key).MainPart = make([]interface{}, len(envelopes[i]))
        copy(effects.ADSRs.GetData(key).MainPart, envelopes[i])
    }

    // Convert MODmodulation parameters to the format used by the YM2612
    mods := make([][]interface{}, len(effects.MODs.GetKeys()))
    for i, key := range effects.MODs.GetKeys() {
        mods[i] = packMOD(effects.MODs.GetData(key).MainPart, specs.CHIP_YM2612)
        effects.MODs.GetData(key).MainPart = make([]interface{}, len(mods[i]))
        copy(effects.MODs.GetData(key).MainPart, mods[i])
    } 
   
    
    /* ToDo: translate
    for i = 1 to length(feedbackMacros[1]) do
        feedbackMacros[ASSOC_DATA][i][LIST_MAIN] = (feedbackMacros[ASSOC_DATA][i][LIST_MAIN])*8
        feedbackMacros[ASSOC_DATA][i][LIST_LOOP] = (feedbackMacros[ASSOC_DATA][i][LIST_LOOP])*8
    end for*/
    
    /*numSongs = 0
    for i = 1 to length(songs) do
        if sequence(songs[i]) then
            numSongs += 1
        end if
    end for*/
             
    if timing.UpdateFreq == 50 {
        outFile.WriteString(".equ XPMP_50_HZ, 1\n")
        t.MachineSpeed = 3546893
    } else {
        t.MachineSpeed = 3579545
    }

    tableSize := outputStandardEffects(outFile, FORMAT_GAS_68K)
    tableSize += outputTable(outFile, FORMAT_GAS_68K, "xpmp_FB_mac", effects.FeedbackMacros, true,  1, 0x80)
    tableSize += outputTable(outFile, FORMAT_GAS_68K, "xpmp_ADSR",   effects.ADSRs,          false, 1, 0)  
    tableSize += outputTable(outFile, FORMAT_GAS_68K, "xpmp_MOD",    effects.MODs,           false, 1, 0)  
    
    /*tableSize += output_m68kas_table("xpmp_VS_mac", volumeSlides, 1, 1, 0)     
    tableSize += output_m68kas_table("xpmp_FB_mac", feedbackMacros,1, 1, 0)
    tableSize += output_m68kas_table("xpmp_ADSR",   adsrs,        0, 1, 0)
    tableSize += output_m68kas_table("xpmp_MOD",    mods,         0, 1, 0)*/

    cbSize := 0
        
    utils.INFO("Size of effect tables: %d bytes\n", tableSize)

    patSize := t.outputPatterns(outFile, FORMAT_GAS_68K)
    utils.INFO("Size of patterns table: %d bytes\n", patSize)
    
    songSize := t.outputChannelData(outFile, FORMAT_GAS_68K) 

    utils.INFO("Total size of song(s): %d bytes\n", songSize + patSize + tableSize + cbSize)

    outFile.Close()
}

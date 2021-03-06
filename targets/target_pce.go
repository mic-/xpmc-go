package targets

import (
    "fmt"
    "os"
    "time"
    "../specs"
    "../utils"
    "../effects"
)


func (t *TargetPCE) Init() {
    t.Target.Init()
    t.Target.SetOutputSyntax(SYNTAX_WLA_DX)
    
    utils.DefineSymbol("PCE", 1)
    utils.DefineSymbol("TGX", 1)
    
    specs.SetChannelSpecs(&t.ChannelSpecs, 0, 0, specs.SpecsHuC6280)      // A..F
    
    t.ID                = TARGET_PCE
    t.MaxTempo          = 300
    t.MinVolume         = 0
    t.SupportsPanning   = 1
    t.SupportsPal       = true
    t.MaxLoopDepth      = 2
    t.MinWavLength      = 32
    t.MaxWavLength      = 32
    t.MinWavSample      = 0
    t.MaxWavSample      = 31
}

/* Output data suitable for the PC-Engine / TurboGrafx-16 (WLA-DX)
 */
func (t *TargetPCE) Output(outputFormat int) {
    utils.DEBUG("TargetPCE.Output")
    
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
 
    t.outputEffectFlags(outFile)
    tableSize := t.outputStandardEffects(outFile)  
    outFile.WriteString("\n")
    utils.INFO("Size of effect tables: %d bytes", tableSize)
    
    tableSize += t.outputTable(outFile, "xpmp_WT_mac", effects.WaveformMacros, true, 1, 0x80)
    tableSize += t.outputTable(outFile, "xpmp_MOD",   effects.MODs, false, 1, 0)

    wavSize := 0
    outFile.WriteString("xpmp_waveform_data:")
    for _, key := range effects.Waveforms.GetKeys() {
        params := effects.Waveforms.GetData(key).MainPart
        for j := 0; j < len(params); j += 1 {
            if j == 0 {
                outFile.WriteString("\n.db ")
            }             
            outFile.WriteString(fmt.Sprintf("$%02x", params[j].(int)))
            wavSize++
            if j < len(params) - 1 {
                outFile.WriteString(",")
            }
        }
    }
    outFile.WriteString("\n\n")
    utils.INFO("Size of waveform table: %d bytes", wavSize)
    
    cbSize := t.outputCallbacks(outFile)


    outFile.WriteString("xpmp_pcm_table:\n")
    for i := 1; i <= 12; i++ {
        outFile.WriteString(fmt.Sprintf(".dw xpmp_pcm%d\n", i - 1))
        outFile.WriteString(fmt.Sprintf(".dw :xpmp_pcm%d\n", i - 1))
    }
    outFile.WriteString("\n")
        
    patSize := t.outputPatterns(outFile)
    utils.INFO("Size of pattern table: %d bytes", patSize)
    
    songSize := t.outputChannelData(outFile) 

    pcmSize := 0
    pcmBank := 2
    outFile.WriteString(fmt.Sprintf("\n.bank %d slot 6\n.org $0000\n", pcmBank))
    pcmBank++
    for i := 1; i <= 12; i++ {
        outFile.WriteString(fmt.Sprintf("xpmp_pcm%d:", i - 1))
        pcmNum := -1
        for j, key := range effects.PCMs.GetKeys() {
            if key == i - 1 {
                pcmNum = j
                break
            }
        }
       
        if pcmNum >= 0 {
            params := effects.PCMs.GetDataAt(pcmNum).LoopedPart[0].([]int)
            params = append(params, 0x80 * 8)
            column := 1
            for j, smp := range params {
                if (column % 32) == 1 {
                    outFile.WriteString("\n.db ")
                }              
                outFile.WriteString(fmt.Sprintf("$%02x", smp / 8))
                pcmSize++
                if (pcmSize % 0x2000) == 0 {
                    outFile.WriteString(fmt.Sprintf("\n.bank %d slot 6\n.org $0000", pcmBank))
                    pcmBank++
                    column = 0
                }
                if j < len(params) - 1 && (column % 32) != 0 {
                    outFile.WriteString(",")
                }
                column += 1
            }
        }
        outFile.WriteString("\n")
    }
    outFile.WriteString("\n\n")
    utils.INFO("Size of XPCM data: %d bytes", pcmSize)

    utils.INFO("Total size of song(s): %d bytes\n", songSize + patSize + tableSize + cbSize + wavSize + pcmSize)
   
    outFile.Close()
}


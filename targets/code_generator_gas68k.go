package targets

import (
    "fmt"
    "os"
    "../effects"
    "../utils"
)

import . "../defs"


func (cg *CodeGeneratorGas68k) OutputCallbacks(outFile *os.File) int {
    callbacksSize := 0

    // ToDo: implement

    utils.INFO("Size of callback table: %d bytes", callbacksSize)  
    
    return callbacksSize
}


/* Outputs the channel data (the actual notes, volume commands, effect invokations, etc)
 * for all channels and all songs.
 */
func (cg *CodeGeneratorGas68k) OutputChannelData(outFile *os.File) int {
    songDataSize := 0
    
    songs := cg.itarget.GetCompilerItf().GetSongs()
    for n, sng := range songs {
        channels := sng.GetChannels()
        if n > 0 {
            fmt.Printf("\n")
        }
        for _, chn := range channels {  
            if chn.IsVirtual() {
                continue       
            }          
            outFile.WriteString(fmt.Sprintf("xpmp_s%d_channel_%s:", n, chn.GetName()))

            commands := chn.GetCommands()
            for j, cmd := range commands {
                if (j % 16) == 0 {
                    outFile.WriteString("\ndc.b ")
                }
                outFile.WriteString(fmt.Sprintf("0x%02x", cmd & 0xFF))
                songDataSize++
                if j < len(commands)-1 && (j % 16) != 15 {
                   outFile.WriteString(",")
                }
            }
            outFile.WriteString("\n")
            fmt.Printf("Song %d, Channel %s: %d bytes, %d / %d ticks\n", sng.GetNum(), chn.GetName(), len(commands), utils.Round2(float64(chn.GetTicks())), utils.Round2(float64(chn.GetLoopTicks())))
        }
    }

    outFile.WriteString("\n.globl xpmp_song_tbl")
    outFile.WriteString("\nxpmp_song_tbl:\n")
    for n, sng := range songs {
        channels := sng.GetChannels()
        for _, chn := range channels { 
            if chn.IsVirtual() {
                continue
            }
            outFile.WriteString(fmt.Sprintf("dc.w xpmp_s%d_channel_%s\n", n, chn.GetName()))
            songDataSize += 2       // ToDo: should it be += 4 like in the original code?
        }
    }
    outFile.WriteString("dc.w 0\n")
    songDataSize += 2
    
    return songDataSize
}


func (cg *CodeGeneratorGas68k) OutputEffectFlags(outFile *os.File) {
    songs := cg.itarget.GetCompilerItf().GetSongs()
    numChannels := len(songs[0].GetChannels())
    
    for _, effName := range EFFECT_STRINGS {
        for c := 0; c < numChannels; c++ {
            for _, sng := range songs {
                channels := sng.GetChannels()
                if channels[c].IsUsingEffect(effName) {
                    // ToDo: translate: outFile.WriteString(fmt.Sprintf(".DEFINE XPMP_CHN%d_USES_", channels[c].GetNum()) + effName + "\n")
                    break
                }
            }
        }
    }
}


/* Outputs the pattern data and addresses.
 */
func (cg *CodeGeneratorGas68k) OutputPatterns(outFile *os.File) int {
    patSize := 0
    
    patterns := cg.itarget.GetCompilerItf().GetPatterns()
    for n, pat := range patterns {
        outFile.WriteString(fmt.Sprintf("xpmp_pattern%d:", n))
        cmds := pat.GetCommands()
        for j, cmd := range cmds {
            if (j % 16) == 0 {
                outFile.WriteString("\ndc.b ")
            }              
            outFile.WriteString(fmt.Sprintf("0x%02x", cmd & 0xFF))
            if j < len(cmds)-1 && (j % 16) != 15 {
                outFile.WriteString(",")
            }
        }
        outFile.WriteString("\n")
        patSize += len(cmds)
    }

    outFile.WriteString("\n.globl xpmp_pattern_tbl\n")
    outFile.WriteString("xpmp_pattern_tbl:\n")
    for n := range patterns {
        outFile.WriteString(fmt.Sprintf("dc.w xpmp_pattern%d\n", n))
        patSize += 2
    }
    outFile.WriteString("\n")
    
    return patSize
}


func (cg *CodeGeneratorGas68k) OutputTable(outFile *os.File, tblName string, effMap *effects.EffectMap, canLoop bool, scaling int, loopDelim int) int {
    var bytesWritten, dat int
    
    bytesWritten = 0
    
    hexPrefix := "0x"
    byteDecl := "dc.b"
    wordDecl := "dc.w"

    if effMap.Len() > 0 {
        for _, key := range effMap.GetKeys() {
            outFile.WriteString(fmt.Sprintf(tblName + "_%d:", key))
            effectData := effMap.GetData(key)
            for j, param := range effectData.MainPart {
                dat = (param.(int) * scaling) & 0xFF
                if canLoop && (dat == loopDelim) {
                    dat++
                }

                if canLoop && j == len(effectData.MainPart)-1 && len(effectData.LoopedPart) == 0 {
                    if j > 0 {
                        outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, loopDelim))
                    }
                    outFile.WriteString(fmt.Sprintf("\n" + tblName + "_%d_loop:\n", key))
                    outFile.WriteString(fmt.Sprintf("%s %s%02x, %s%02x", byteDecl, hexPrefix, dat, hexPrefix, loopDelim))
                    bytesWritten += 3
                } else if j == 0 {
                    outFile.WriteString(fmt.Sprintf("\n%s %s%02x", byteDecl, hexPrefix, dat))
                    bytesWritten += 1
                } else {
                    outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, dat))
                    bytesWritten += 1
                }
            }
            if canLoop && len(effectData.LoopedPart) > 0 {
                if len(effectData.MainPart) > 0 {
                    outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, loopDelim))
                    bytesWritten += 1
                }
                outFile.WriteString(fmt.Sprintf("\n" + tblName + "_%d_loop:\n", key))
                for j, param := range effectData.LoopedPart {
                    dat = (param.(int) * scaling) & 0xFF
                    if dat == loopDelim && canLoop {
                        dat++
                    }
                    if j == 0 {
                        outFile.WriteString(fmt.Sprintf("%s %s%02x", byteDecl, hexPrefix, dat))
                    } else {
                        outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, dat))
                    }
                    bytesWritten += 1
                }
                outFile.WriteString(fmt.Sprintf(", %s%02x", hexPrefix, loopDelim))
                bytesWritten += 1
            }
            outFile.WriteString("\n")
        }
        outFile.WriteString(tblName + "_tbl:\n")
        for _, key := range effMap.GetKeys() {
            outFile.WriteString(fmt.Sprintf("%s " + tblName + "_%d\n", wordDecl, key))
            bytesWritten += 2
        }
        if canLoop {
            outFile.WriteString(tblName + "_loop_tbl:\n")
            for _, key := range effMap.GetKeys() {
                outFile.WriteString(fmt.Sprintf("%s " + tblName + "_%d_loop\n", wordDecl, key))
                bytesWritten += 2
            }
        }
        outFile.WriteString("\n")
    } else {
        outFile.WriteString(tblName + "_tbl:\n")
        if canLoop {
            outFile.WriteString(tblName + "_loop_tbl:\n")
        }
        outFile.WriteString("\n")
    }
        
    return bytesWritten
}


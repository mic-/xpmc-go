package targets

import (
    "fmt"
    "os"
    "../effects"
    "../utils"
)

import . "../defs"


func (cg *CodeGeneratorWla) OutputCallbacks(outFile *os.File) int {
    callbacksSize := 0

    outFile.WriteString("xpmp_callback_tbl:\n")
    for _, cb := range cg.itarget.GetCompilerItf().GetCallbacks() {
        outFile.WriteString(".dw " + cb + "\n")
        callbacksSize += 2
    }
    outFile.WriteString("\n")

    utils.INFO("Size of callback table: %d bytes", callbacksSize)  
    
    return callbacksSize
}


func (cg *CodeGeneratorWla) OutputEffectFlags(outFile *os.File) {
    songs := cg.itarget.GetCompilerItf().GetSongs()
    numChannels := len(songs[0].GetChannels())
    
    for _, effName := range EFFECT_STRINGS {
        for c := 0; c < numChannels; c++ {
            for _, sng := range songs {
                channels := sng.GetChannels()
                if channels[c].IsUsingEffect(effName) {
                    outFile.WriteString(fmt.Sprintf(".DEFINE XPMP_CHN%d_USES_", channels[c].GetNum()) + effName + "\n")
                    break
                }
            }
        }
    }
}


/* Outputs the pattern data and addresses.
 */
func (cg *CodeGeneratorWla) OutputPatterns(outFile *os.File) int {
    patSize := 0
    
    patterns := cg.itarget.GetCompilerItf().GetPatterns()
    for n, pat := range patterns {
        outFile.WriteString(fmt.Sprintf("xpmp_pattern%d:", n))
        cmds := pat.GetCommands()
        for j, cmd := range cmds {
            if (j % 16) == 0 {
                outFile.WriteString("\n.db ")
            }              
            outFile.WriteString(fmt.Sprintf("$%02x", cmd & 0xFF))
            if j < len(cmds)-1 && (j % 16) != 15 {
                outFile.WriteString(",")
            }
        }
        outFile.WriteString("\n")
        patSize += len(cmds)
    }

    outFile.WriteString("\nxpmp_pattern_tbl:\n")
    for n := range patterns {
        outFile.WriteString(fmt.Sprintf(".dw xpmp_pattern%d\n", n))
        patSize += 2
    }
    outFile.WriteString("\n")
        
    return patSize
}


/* Outputs the channel data (the actual notes, volume commands, effect invokations, etc)
 * for all channels and all songs.
 */
func (cg *CodeGeneratorWla) OutputChannelData(outFile *os.File) int {
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
                    outFile.WriteString("\n.db ")
                }
                outFile.WriteString(fmt.Sprintf("$%02x", cmd & 0xFF))
                songDataSize++
                if j < len(commands)-1 && (j % 16) != 15 {
                   outFile.WriteString(",")
                }
            }
            outFile.WriteString("\n")
            fmt.Printf("Song %d, Channel %s: %d bytes, %d / %d ticks\n",
                sng.GetNum(), chn.GetName(), len(commands), utils.Round2(float64(chn.GetTicks())), utils.Round2(float64(chn.GetLoopTicks())))
        }
    }

    outFile.WriteString("\nxpmp_song_tbl:\n")
    for n, sng := range songs {
        channels := sng.GetChannels()
        for _, chn := range channels { 
            if chn.IsVirtual() {
                continue
            }
            outFile.WriteString(fmt.Sprintf(".dw xpmp_s%d_channel_%s\n", n, chn.GetName()))
            songDataSize += 2
        }
    }
    
    return songDataSize
}


func (cg *CodeGeneratorWla) OutputTable(outFile *os.File, tblName string, effMap *effects.EffectMap, canLoop bool, scaling int, loopDelim int) int {
    var bytesWritten, dat int
    
    bytesWritten = 0
    
    hexPrefix := "$"
    byteDecl := ".db"
    wordDecl := ".dw"

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


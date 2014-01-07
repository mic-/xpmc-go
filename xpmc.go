package main

import (
    "fmt"
    "os"
    "strings"
    "./compiler"
    "./defs"
    "./targets"
    "./timing"
    "./utils"
//    "./player"
)


var target int


func showHelp(what string) {
    switch what {
    case "EN":
        fmt.Println("EN:\tArpeggio macro: { num | num}. num is zero or more numbers in\n\tthe range -63-63.\n")
    default:
        fmt.Println("Usage: xpmc.exe [options] target input [output]\n")
        fmt.Println("Options:")
        fmt.Println("\t-h\tShow this information") 
        fmt.Println("\t-v\tVerbose mode")
        fmt.Println("\t-w\tTreat warnings as errors")
        fmt.Println("\nTarget:")
        fmt.Println("\t-at8\tAtari 8-bit")
        fmt.Println("\t-c64\tCommodore 64")
        fmt.Println("\t-clv\tColecoVision")
        fmt.Println("\t-cpc\tAmstrad CPC")
        fmt.Println("\t-cps\tCPS-1")
        //fmt.Println(1, "\t-gba\tGameboy Advance")
        fmt.Println("\t-gbc\tGameboy / Gameboy Color")
        fmt.Println("\t-gen\tSEGA Genesis")
        fmt.Println("\t-kss\tKSS")
        fmt.Println("\t-nes\tNintendo Entertainment System")
        fmt.Println("\t-pce\tPC-Engine")
        //fmt.Println(1, "\t-nds\tNintendo DS")
        //fmt.Println(1, "\t-sat\tSEGA Saturn")
        fmt.Println("\t-sgg\tSEGA Game Gear")
        fmt.Println("\t-sms\tSEGA Master System")

    }
}


/*integer effectsRemoved
effectsRemoved = 0
func removeUnusedEffects(effectMap *effects.EffectMap, effectCmd int)
    integer n
    n = 1
    
    pos := 0
    for pos < effectMap.Len() {
        key := effectMap.GetKeyAt(pos)
        if !effectMap.IsReferenced(key) {     
            for i = 1 to length(songs) do
                if sequence(songs[i]) then
                    for j, chn := range songs[i].Channels {
                        for k = 1 to length(songs[i][j]) do
                            if songs[i][j][k] == effectCmd {
                                if (songs[i][j][k + 1] & 0x7F) > pos {
                                    songs[i][j][k + 1] = ((songs[i][j][k + 1] & 0x7F) - 1) | (songs[i][j][k + 1] & 0x80)
                                }
                            }
                        end for
                    }
                end if
            end for
            tbl[ASSOC_KEY]   = substr2(tbl[ASSOC_KEY],   1, n-1) & substr2(tbl[ASSOC_KEY],   n+1, length(tbl[ASSOC_KEY]))
            tbl[ASSOC_DATA]  = substr2(tbl[ASSOC_DATA],  1, n-1) & substr2(tbl[ASSOC_DATA],  n+1, length(tbl[ASSOC_DATA]))
            tbl[ASSOC_REF]   = substr2(tbl[ASSOC_REF],   1, n-1) & substr2(tbl[ASSOC_REF],   n+1, length(tbl[ASSOC_REF]))
            tbl[ASSOC_EXTRA] = substr2(tbl[ASSOC_EXTRA], 1, n-1) & substr2(tbl[ASSOC_EXTRA], n+1, length(tbl[ASSOC_EXTRA]))
            effectsRemoved += 1
        } else {
            pos += 1
        }
    }
}    */


func main() {
    timing.UpdateFreq = 60.0
    timing.UseFractionalDelays = true

    if timing.UseFractionalDelays {
        timing.SupportedLengths = defs.EXTENDED_LENGTHS()
    }
    
    utils.OldParsers = utils.NewGenericStack()
    
    utils.Verbose(false)
    utils.WarningsAreErrors(false)
    target = targets.TARGET_UNKNOWN

    comp := &compiler.Compiler{}
    
    for i, arg := range os.Args {
        if i >= 1 {
            if arg[0] == '-' {
                if arg == "-v" {
                    utils.Verbose(true);
                } else if arg == "-w" {
                    utils.WarningsAreErrors(true);
                } else if arg == "-h" || arg == "-help" {
                    showHelp("")
                    return
                } else if targets.NameToID(arg[1:]) != targets.TARGET_UNKNOWN {
                    target = targets.NameToID(arg[1:])
                } 
            } else {
                if target == targets.TARGET_UNKNOWN {
                    fmt.Printf("Error: No target platform specified.\nRun the compiler with the -h option to see a list of targets.")
                    return
                }
                comp.Init(target)

                inputFileName := arg
                lastDot := strings.LastIndexAny(inputFileName, ".")
                lastSlash := strings.LastIndexAny(inputFileName, "/\\")
                
                if lastDot >= 0 && lastSlash < 0 {
                    comp.ShortFileName = inputFileName[:lastDot]
                } else {
                    comp.ShortFileName = inputFileName
                    inputFileName += ".mml"
                }

                if i < len(os.Args)-1 {
                    comp.ShortFileName = os.Args[i + 1]
                    lastDot = strings.LastIndexAny(os.Args[i + 1], ".")
                    if lastDot >= 0 {
                        comp.ShortFileName = os.Args[i+1][:lastDot]
                        //writeVGM = equal(lower(fileNames[2][n..length(fileNames[2])]), ".vgm")
                        //writeVGM += equal(lower(fileNames[2][n..length(fileNames[2])]), ".vgz") * 2
                        //writeWAV = equal(lower(fileNames[2][n..length(fileNames[2])]), ".wav")
                    }
                }
                
                //fmt.Printf("compiler.SFN = " + comp.ShortFileName + "\n")
                
                comp.CompileFile(inputFileName);
    
                // Insert END markers or jumps at the end of the command sequence for each channel
                for _, song := range comp.Songs {
                    for _, chn := range song.Channels {
                        if chn.IsVirtual() {
                            continue
                        }
                        chn.LoopTicks = chn.Ticks - chn.LoopTicks
                        if chn.LoopPoint == -1 {
                            chn.AddCmd([]int{defs.CMD_END})
                        } else {
                            if !chn.HasAnyNote {
                                chn.AddCmd([]int{defs.CMD_END})
                            } else {
                                chn.AddCmd([]int{defs.CMD_JMP, chn.LoopPoint & 0xFF, chn.LoopPoint / 0x100})
                            }
                        }
                    }
                }
 
                ticks := -1
                for _, song := range comp.Songs {
                    for _, chn := range song.Channels {
                        if chn.IsVirtual() {
                            continue
                        }
                        if chn.HasAnyNote {
                            if ticks == -1 {
                                ticks = chn.Ticks
                            } else if chn.Ticks != ticks {
                                fmt.Printf("Warning: Mismatch in length between channels in song %d\n", song.Num)
                                break
                            }
                        }
                    }
                }
    
                comp.CurrSong.Target.Output(0)
                
                return
            }
        }
    }

    showHelp("")
}

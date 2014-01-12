/*
 * Package compiler
 *
 * Part of XPMC.
 * Contains functions for handling meta-commands (those commands
 * that begin with a #, e.g. #IFDEF, #SONG, #BASE, etc).
 *
 * /Mic, 2013-2014
 */
 
package compiler

import (
    "strconv"
    "strings"
    "../defs"
    "../song"
    "../targets"
    "../timing"
    "../utils"
)

import . "../utils"

const (
    POLARITY_POSITIVE = 0
    POLARITY_NEGATIVE = 1
    
    ELSIFDEF_TAKEN = 2
)

/* Parses an expression on the form  SYM op SYM op ...
 * Where op is either | or &
 * The polarity specifies if this expression is for an IF or IFN, and determines
 * how the ops should be interpreted.
 * Returns 1 if the expression is true, otherwise 0
 */
func evalIfdefExpr(polarity int) int {
    s := Parser.GetStringUntil("&|\r\n")
    expr := IsDefined(s)
    for {
        s = Parser.GetStringInRange("&|")
        if s == "&" {
            s = Parser.GetStringUntil("&|\r\n")
            if polarity == POLARITY_POSITIVE {
                expr = expr & IsDefined(s)
            } else {
                expr = expr | IsDefined(s)
            }
        } else if s == "|" {
            s = Parser.GetStringUntil("&|\r\n")
            if polarity == POLARITY_POSITIVE {
                expr = expr | IsDefined(s)
            } else {
                expr = expr & IsDefined(s)
            }
        } else {
            break
        }
    }
    
    return expr
}


/* Handles commands starting with '#', i.e. a meta command.
 */
func (comp *Compiler) handleMetaCommand() {
    if comp.dontCompile.PeekInt() != 0 {
        s := Parser.GetString()
        switch s {
        case "IFDEF":
            expr := evalIfdefExpr(POLARITY_POSITIVE)
            comp.dontCompile.Push((expr ^ 1) | comp.dontCompile.PeekInt())
            comp.hasElse.Push(false)

        case "IFNDEF":
            expr := evalIfdefExpr(POLARITY_NEGATIVE)
            comp.dontCompile.Push(expr | comp.dontCompile.PeekInt())
            comp.hasElse.Push(false)             

        case "ELSIFDEF":
            expr := evalIfdefExpr(POLARITY_POSITIVE)
            if comp.dontCompile.Len() > 1 {
                if !comp.hasElse.PeekBool() {
                    if (comp.dontCompile.PeekInt() & ELSIFDEF_TAKEN) != ELSIFDEF_TAKEN {
                        _ = comp.dontCompile.PopInt()
                        comp.dontCompile.Push((expr ^ 1) | comp.dontCompile.PeekInt())
                    }
                } else {
                    ERROR("ELSIFDEF found after ELSE")
                }
            } else {
                ERROR("ELSIFDEF with no matching IFDEF")
            }                   


        case "ELSE":
            if comp.dontCompile.Len() > 1 {
                if !comp.hasElse.PeekBool() {
                    if (comp.dontCompile.PeekInt() & ELSIFDEF_TAKEN) != ELSIFDEF_TAKEN {
                        x := comp.dontCompile.PopInt()
                        comp.dontCompile.Push((x ^ 1) | comp.dontCompile.PeekInt())
                    }
                    _ = comp.hasElse.PopBool()
                    comp.hasElse.Push(true)
                } else {
                    ERROR("Only one ELSE allowed per IFDEF")
                }
            } else {
                ERROR("ELSE with no matching IFDEF")
            }

        case "ENDIF":
            if comp.dontCompile.Len() > 1 {
                _ = comp.dontCompile.PopInt()
                _ = comp.hasElse.PopBool()
            } else {
                ERROR("ENDIF with no matching IFDEF")
            }
        }
    } else {
        for _, chn := range comp.CurrSong.Channels {
            chn.WriteNote(true)
        }

        cmd := Parser.GetString()

        switch cmd {
        case "IFDEF":
            expr := evalIfdefExpr(POLARITY_POSITIVE)
            comp.dontCompile.Push((expr ^ 1) | comp.dontCompile.PeekInt())
            comp.hasElse.Push(false)

        case "IFNDEF":
            expr := evalIfdefExpr(POLARITY_NEGATIVE)
            comp.dontCompile.Push(expr | comp.dontCompile.PeekInt())
            comp.hasElse.Push(false)

        case "ELSIFDEF":
            if comp.dontCompile.Len() > 1 {
                if !comp.hasElse.PeekBool() {
                    _ = Parser.GetStringUntil("\r\n")
                    _ = comp.dontCompile.PopInt()
                    // Getting here means that the current IFDEF/ELSIFDEF was true,
                    // so whatever is in subsequent ELSIFDEF/ELSE clauses should not
                    // be compiled.
                    comp.dontCompile.Push(ELSIFDEF_TAKEN)
                } else {
                    ERROR("ELSIFDEF found after ELSE")
                }
            } else {
                ERROR("ELSIFDEF with no matching IFDEF")
            }

        case "ELSE":
            if comp.dontCompile.Len() > 1 {
                if !comp.hasElse.PeekBool() {
                    _ = comp.dontCompile.PopInt()
                    comp.dontCompile.Push(1)
                    _ = comp.hasElse.PopBool()
                    comp.hasElse.Push(true)
                } else {
                    ERROR("Only one ELSE allowed per IFDEF")
                }
            } else {
                ERROR("ELSE with no matching IFDEF")
            }

        case "ENDIF":
            if comp.dontCompile.Len() > 1 {
                _ = comp.dontCompile.PopInt()
                _ = comp.hasElse.PopBool()
            } else {
                ERROR("ENDIF with no matching IFDEF")
            }
                    
        case "TITLE":
            comp.CurrSong.Title = Parser.GetStringUntil("\r\n")

        case "TUNE":
            comp.CurrSong.TuneSmsPitch = true

        case "COMPOSER":
            comp.CurrSong.Composer = Parser.GetStringUntil("\r\n")

        case "PROGRAMER","PROGRAMMER":
            comp.CurrSong.Programmer = Parser.GetStringUntil("\r\n")

        case "GAME":
            comp.CurrSong.Game = Parser.GetStringUntil("\r\n")

        case "ALBUM":
            comp.CurrSong.Album = Parser.GetStringUntil("\r\n")

        case "ERROR":
            Parser.SkipWhitespace()
            if Parser.Getch() == '"' {
                s := Parser.GetStringUntil("\"")
                if Parser.Getch() == '"' {
                    ERROR(s)
                } else {
                    ERROR("Malformed #ERROR, missing ending \"")
                }
            } else {
                ERROR("Malformed #ERROR, missing starting \"")
            }

        case "WARNING":
            Parser.SkipWhitespace()
            if Parser.Getch() == '"' {
                s := Parser.GetStringUntil("\"")
                if Parser.Getch() == '"' {
                    WARNING(s)
                } else {
                    ERROR("Malformed #WARNING, missing ending \"")
                }
            } else {
                ERROR("Malformed #WARNING, missing starting \"")
            }
                    
        case "INCLUDE":
            Parser.SkipWhitespace()
            if Parser.Getch() == '"' {
                s := Parser.GetStringUntil("\"")
                if Parser.Getch() == '"' {
                    if len(s) > 0 {
                        if !strings.ContainsRune(s, ':') && s[0] != '\\' {
                            s = Parser.WorkDir + s
                        }
                        comp.CompileFile(s)
                    }
                } else {
                    ERROR("Malformed #INCLUDE, missing ending \"")
                }
            } else {
                ERROR("Malformed #INCLUDE, missing starting \"")
            }

        case "PAL":
            if comp.CurrSong.Target.SupportsPAL() {
                timing.UpdateFreq = 50.0
            }

        case "NTSC":
            timing.UpdateFreq = 60.0

        case "SONG":
            s := Parser.GetString()
            num, err := strconv.ParseInt(s, Parser.UserDefinedBase, 0)
            if err == nil {
                if num > 1 && num < 100 {
                    // A song with the given number must not already exist
                    if _, songExists := comp.Songs[int(num)]; !songExists {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.IsVirtual() {
                                continue
                            }
                            // Verify that all []-loops have been closed
                            if chn.Loops.Len() > 0 {
                                utils.ERROR("Open [ loop on channel %s", chn.Name)
                            }
                            // Add END markers or jumps at the end of all channels
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
                        
                        if comp.keepChannelsActive || (len(comp.patName) != 0) {
                            utils.ERROR("Missing }")
                        }
                        
                        /*songNum = o[2]
                        songs[songNum] = repeat({}, length(supportedChannels))
                        songLen[songNum] = repeat(0, length(supportedChannels))
                        hasAnyNote = repeat(0, length(supportedChannels))
                        songLoopLen[songNum] = songLen[songNum]*/                   
   
                        // Create a new song and make it the current one
                        comp.CurrSong = song.NewSong(int(num), comp.CurrSong.Target.GetID(), comp)
                        comp.Songs[int(num)] = comp.CurrSong
    
                    } else {
                        ERROR("Song " + s + " already defined")
                    }
                } else {
                    ERROR("Bad song number: " + s)
                }
            } else {
                ERROR("Bad song number: " + s)
            }

        case "BASE":
            s := Parser.GetString()
            newBase, err := strconv.ParseInt(s, 10, 0) //Parser.UserDefinedBase, 0)
            if err == nil {
                if newBase == 10 || newBase == 16 {
                    Parser.UserDefinedBase = int(newBase)
                } else {
                    WARNING(cmd +": Expected 10 or 16, got: " + s)
                }
            } else {
                ERROR(cmd +": Expected 10 or 16, got: " + s)
            }

        case "UNIFORM-VOLUME":
            s := Parser.GetString()
            vol, err := strconv.Atoi(s)
            if err == nil {
                if vol > 1 {
                    for _, chn := range comp.CurrSong.Channels {
                        chn.SetMaxVolume(vol)
                    }
                    if vol > 255 {
                        WARNING("Very large max volume specified: " + s)
                    }
                } else {
                    ERROR("Volume must be >= 1: " + s)
                }
            } else {
                ERROR(cmd + ": Expected a positive integer: " + s)
            }

        case "GB-VOLUME-CONTROL":
            s := Parser.GetString()
            ctl, err := strconv.Atoi(s)
            if err == nil {
                if ctl == 1 {
                    comp.gbVolCtrl = 1
                } else if ctl == 0 {
                    comp.gbVolCtrl = 0
                } else {
                    WARNING(cmd + ": Expected 0 or 1, got: " + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got: " + s)
            }

        case "GB-NOISE":
            s := Parser.GetString()
            val, err := strconv.Atoi(s)
            if err == nil {
                if val == 1 {
                    comp.gbNoise = 1
                    if comp.CurrSong.Target.GetID() == targets.TARGET_GBC {
                        comp.CurrSong.Channels[3].SetMaxOctave(5)
                    }
                } else if val == 0 {
                    comp.gbNoise = 0
                    if comp.CurrSong.Target.GetID() == targets.TARGET_GBC {
                        comp.CurrSong.Channels[3].SetMaxOctave(11)
                    }
                } else {
                    WARNING(cmd + ": Expected 0 or 1, defaulting to 0: " + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got: " + s)
            }

        case "EN-REV":
            s := Parser.GetString()
            rev, err := strconv.Atoi(s)
            if err == nil {
                if rev == 1 {
                    comp.enRev = 1
                    if comp.CurrSong.Target.GetID() == targets.TARGET_C64 ||
                       comp.CurrSong.Target.GetID() == targets.TARGET_AT8 {
                        ERROR("#EN-REV 1 is not supported for this target")
                    }
                } else if rev == 0 {
                    comp.enRev = 0
                } else {
                    WARNING(cmd + ": Expected 0 or 1, defaulting to 0: " + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got: " + s)
            }

        case "OCTAVE-REV":
            s := Parser.GetString()
            rev, err := strconv.Atoi(s)
            if err == nil {
                if rev == 1 {
                    comp.octaveRev = -1
                } else if rev == 0 {
                } else {
                    WARNING(cmd + ": Expected 0 or 1, defaulting to 0:" + s)
                }
            } else {
                ERROR(cmd + ": Expected 0 or 1, got:" + s)
            }

        case "AUTO-BANKSWITCH":
            WARNING("Unsupported command: AUTO-BANKSWITCH")
            _ = Parser.GetString() 

        default:
            ERROR("Unknown command: " + cmd)
        }
    }
}
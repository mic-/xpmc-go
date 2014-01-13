/*
 * Package compiler
 *
 * Part of XPMC.
 * Contains functions for handling effect definitions found in
 * MML code (e.g. EP0 = { 1 2 | -2}).
 *
 * /Mic, 2013-2014
 */
 
package compiler

import (
    "os"
    "strconv"
    "strings"
    "../channel"
    "../defs"
    "../effects"
    "../targets"
    "../utils"
    "../wav"
)

import . "../utils"


func (comp *Compiler) getEffectFrequency() int {
    var n, retVal int
    
    retVal = defs.EFFECT_STEP_EVERY_FRAME
    
    Parser.SkipWhitespace()
    n = Parser.Getch()
    if n == '(' {
        t := Parser.GetStringUntil(")\t\r\n ")
        if t == "EVERY-FRAME" {
            retVal = defs.EFFECT_STEP_EVERY_FRAME
        } else if t == "EVERY-NOTE" {
            retVal = defs.EFFECT_STEP_EVERY_NOTE
        } else {
            ERROR("Unsupported effect frequency: " + t)
        }
        
        if Parser.Getch() != ')' {
            ERROR("Syntax error: expected )")
        }
    } else {
        Parser.Ungetch()
    }
    
    return retVal
}


func (comp *Compiler) handleAdsrEnvelopeDef(cmd string, isInlined bool) int {
    num := -1
    if isInlined {              
        Parser.SetListDelimiters("()")
        lst, err := Parser.GetList()
        key := -1
        if err == nil {
            if len(lst.LoopedPart) == 0 {
                if len(lst.MainPart) == comp.CurrSong.Target.GetAdsrLen() {
                    if inRange(lst.MainPart, 0, comp.CurrSong.Target.GetAdsrMax()) {
                        key = effects.ADSRs.GetKeyFor(lst)
                        if key == -1 {
                            effects.ADSRs.Append(effects.ADSRs.InlinedDefinitionId, lst)
                            num = effects.ADSRs.InlinedDefinitionId
                            effects.ADSRs.InlinedDefinitionId++
                        } else {
                            num = key
                        }
                    } else {
                        ERROR("ADSR parameters out of range: " + lst.Format())
                    }
                } else {
                    ERROR("Bad number of ADSR parameters: " + lst.Format())
                }
            } else {
                ERROR("Bad ADSR: " + lst.Format())
            }
        } else {
            ERROR("Bad ADSR: Unable to parse parameter list")
        }
    } else {
        // Normal definition
        num = comp.handleEffectDefinition("ADSR", cmd, effects.ADSRs, func(parm *ParamList) bool {
                if !parm.IsEmpty() {
                    if len(parm.LoopedPart) == 0 {
                        if len(parm.MainPart) == comp.CurrSong.Target.GetAdsrLen() {
                            if inRange(parm.MainPart, 0, comp.CurrSong.Target.GetAdsrMax()) {
                                return true
                            } else {
                                ERROR("ADSR parameters out of range: " + parm.Format())
                            }
                        } else {
                            ERROR("Bad number of ADSR parameters: " + parm.Format())
                        }
                    } else {
                        ERROR("| loops are not allowed in ADSR envelopes: " + parm.Format())
                    }
                    return true
                } else {
                    ERROR("Empty list for ADSR")
                }
                return false
            })
    }
    
    return num
}


/* Handles definitions of duty macros (@<n> = { ... })
 */
func (comp *Compiler) handleDutyMacDef(num int) {
    idx := effects.DutyMacros.FindKey(num) 
    if comp.CurrSong.GetNumActiveChannels() == 0 {
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    if len(lst.MainPart) != 0 || len(lst.LoopedPart) != 0 {
                        effects.DutyMacros.Append(num, lst)
                        effects.DutyMacros.PutExtraInt(num, effects.EXTRA_EFFECT_FREQ, comp.getEffectFrequency())
                    } else {
                        ERROR("Empty list for @")
                    }
                } else {
                    ERROR("Syntax error, unable to parse list")
                }
            } else {
                ERROR("Expected '=', got: " + t)
            }
        } else {
            ERROR("Redefinition of @" + strconv.FormatInt(int64(num), 10))
        }
    } else {
        numChannels := 0
        for _, chn := range comp.CurrSong.Channels {
            if chn.Active {
                numChannels++
                if num >= 0 && num <= chn.SupportsDutyChange() {
                    if !chn.Tuple.Active {
                        chn.AddCmd([]int{defs.CMD_DUTY | num})
                    } else {
                        chn.Tuple.Cmds = append(chn.Tuple.Cmds, channel.Note{Num: 0xFFFF, Frames: float64(defs.CMD_DUTY | num), HasData: true})
                    }
                } else {
                    if chn.SupportsDutyChange() == -1 {
                        WARNING("Unsupported command for channel " + chn.GetName() + ": @")
                    } else {
                        ERROR("@ out of range: " + strconv.FormatInt(int64(num), 10))
                    }
                }
            }
        }
        if numChannels == 0 {
            WARNING("Use of @ with no channels active")
        }
    }
                            
}

/* Handles definitions of panning macros ("@CS<xy> = {...}")
 */
func (comp *Compiler) handlePanMacDef(cmd string) {
    num, err := strconv.Atoi(cmd[2:])
    if err == nil {
        // Already defined?
        idx := effects.PanMacros.FindKey(num)
        if idx < 0 {
            // ..no. Get the '=' sign and then the list of values
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    // The list must contain at least one value
                    if !lst.IsEmpty() {
                        // PC-Engine supports any values, Super Famicom supports -127..+127, all other
                        // targets support -63..+63
                        if (inRange(lst.MainPart, -63, 63) && inRange(lst.LoopedPart, -63, 63)) ||
                           (inRange(lst.MainPart, -127, 127) &&
                              inRange(lst.LoopedPart, -127, 127) &&
                              comp.CurrSong.Target.GetID() == targets.TARGET_SFC) ||
                           comp.CurrSong.Target.GetID() == targets.TARGET_PCE {
                            // Store this effect among the pan macros
                            effects.PanMacros.Append(num, lst)
                            // And store the effect frequency for this particular effect (whether it should
                            // be applied once per frame or once per note).
                            effects.PanMacros.PutExtraInt(num, effects.EXTRA_EFFECT_FREQ, comp.getEffectFrequency())
                        } else {
                            ERROR("Value of out range (allowed: -63-63): " + lst.Format())
                        }
                    } else {
                        ERROR("Empty list for CS")
                    }
                } else {
                    ERROR("Bad CS: " + t)
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }
}


/* Handle definitions of filters ("@FT<xy> = {...}")
 */
func (comp *Compiler) handleFilterDef(cmd string) {
    _ = comp.handleEffectDefinition("FT", cmd, effects.Filters, func(parm *ParamList) bool {
            if !parm.IsEmpty() {
                if comp.CurrSong.Target.GetID() == targets.TARGET_C64 {
                    if len(parm.MainPart) == 3 && len(parm.LoopedPart) == 0 {
                        if inRange(parm.MainPart, []int{0, 0, 0}, []int{3, 2047, 15}) {
                            return true
                        }
                    }
                } else {
                    return true
                }
                return false
            } else {
                ERROR("Empty list for FT")
            }
            return false
        })  
}


/* Handle definitions of modulation macros ("@MOD<xy> = {...}")
 */
func (comp *Compiler) handleModMacDef(cmd string) {
    num, err := strconv.Atoi(cmd[3:])
    if err == nil {
        idx := effects.MODs.FindKey(num)
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    switch comp.CurrSong.Target.GetID() {
                    case targets.TARGET_SMD:
                        // 3 parameter version: Genesis/Megadrive (YM2612)
                        if len(lst.MainPart) == 3 && len(lst.LoopedPart) == 0 {
                            if inRange(lst.MainPart, 0, []int{7, 7, 3}) {
                                effects.MODs.Append(num, lst)
                            } else {
                                ERROR("Value of out range: " + lst.Format())
                            }
                        } else {
                            ERROR("Bad MOD, expected 3 parameters: " + lst.Format())
                        }
                    case targets.TARGET_CPS, targets.TARGET_X68:
                        // 6 parameter version: CPS-1, X68000 (YM2151)
                        if len(lst.MainPart) == 6 && len(lst.LoopedPart) == 0 {
                            if inRange(lst.MainPart, 0, []int{255, 127, 127, 3, 7, 3}) {
                                effects.MODs.Append(num, lst)
                            } else {
                                ERROR("Value of out range: " + lst.Format())
                            }
                        } else {
                            ERROR("Bad MOD, expected 6 parameters: " + lst.Format())
                        }
                    case targets.TARGET_PCE:
                        // 2 parameter version: PC-Engine (HuC6280)
                        if len(lst.MainPart) == 2 && len(lst.LoopedPart) == 0 {
                            if inRange(lst.MainPart, 0, []int{255, 3}) {
                                effects.MODs.Append(num, lst)
                            } else {
                                ERROR("Value of out range: " + lst.Format())
                            }
                        } else {
                            ERROR("Bad MOD, expected 2 parameters: " + lst.Format())
                        }
                    case targets.TARGET_KSS:
                        // ToDo: allow both 3-parameter and 6-parameter versions
                    default:
                        effects.MODs.Append(num, &utils.ParamList{})
                    }
                } else {
                    ERROR("Bad MOD: " + t)
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }   
}


/* Handles definitions of vibratos ("@MP<xy> = { ... }")
 */
func (comp *Compiler) handleVibratoDef(cmd string) {
    _ = comp.handleEffectDefinition("MP", cmd, effects.Vibratos, func(parm *ParamList) bool {
            if len(parm.MainPart) == 3 && len(parm.LoopedPart) == 0 {
                if inRange(parm.MainPart, []int{0, 1, 0}, []int{127, 127, 63}) {
                    return true
                } else {
                    ERROR("Value of out range: " + parm.Format())
                }
            } else {
                ERROR("Bad MP: " + parm.Format())
            }
            return false
        })
}


func (comp *Compiler) handleXpcmDef(cmd string) {
    num, err := strconv.Atoi(cmd[4:])
    if err == nil {
        idx := effects.PCMs.FindKey(num)
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if comp.CurrSong.Target.SupportsPCM() {
                    if err == nil {
                        if len(lst.LoopedPart) == 0 {
                            if len(lst.MainPart) > 0 {
                                if pcmFileName, ok := lst.MainPart[0].(string); ok {
                                    if !strings.ContainsRune(pcmFileName, rune(':')) && pcmFileName[0] != os.PathSeparator {
                                        pcmFileName = Parser.WorkDir + pcmFileName
                                        lst.MainPart[0] = pcmFileName
                                    }
                                    if len(lst.MainPart) > 2 {
                                        // {"filename" samplerate volume}
                                        lst.LoopedPart = append(lst.LoopedPart, wav.ConvertWav(pcmFileName, lst.MainPart[1].(int), lst.MainPart[2].(int)))
                                    } else {
                                        // {"filename" samlerate}
                                        lst.LoopedPart = append(lst.LoopedPart, wav.ConvertWav(pcmFileName, lst.MainPart[1].(int), 100))
                                    }
                                    effects.PCMs.Append(num, lst)
                                } else {
                                    ERROR("Bad XPCM: " + lst.Format())
                                }
                            } else {
                                ERROR("Bad XPCM: " + lst.Format())
                            }
                        } else {
                            ERROR("Loops not supported in XPCM: " + lst.Format())
                        }
                    } else {
                        ERROR("Bad XPCM: " + t)
                    }
                } else {
                    WARNING("Unsupported command for this target: @XPCM")
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + cmd)
        }
    } else {
        ERROR("Syntax error: @" + cmd)
    }       
}


func (comp *Compiler) handleEffectDefinition(effName string, mmlString string, effMap *effects.EffectMap, pred func(*ParamList) bool) int {
    num, err := strconv.Atoi(mmlString[len(effName):])
    if err == nil {
        idx := effMap.FindKey(num)
        if idx < 0 {
            t := Parser.GetString()
            if t == "=" {
                lst, err := Parser.GetList()
                if err == nil {
                    if pred(lst) {
                        freq := comp.getEffectFrequency()
                        key := effMap.GetKeyFor(lst)
                        if key == -1 {
                           effMap.Append(num, lst)
                           effMap.PutExtraInt(num, effects.EXTRA_EFFECT_FREQ, freq)
                        } else { /*if freq != effMap.GetInt(key) {*/
                            // ToDo: handle the case when we've got an existing identical effect. The references to the new
                            // effects needs to be converted to refer to the old effect.
                           effMap.Append(num, lst)
                           effMap.PutExtraInt(num, effects.EXTRA_EFFECT_FREQ, freq)
                        }
                    }
                } else {
                    ERROR("Bad " + effName +": " + lst.Format())
                }
            } else {
                ERROR("Expected '='")
            }
        } else {
            ERROR("Redefinition of @" + mmlString)
        }
    } else {
        ERROR("Syntax error: @" + mmlString)
    }
    
    return num
}


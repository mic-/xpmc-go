package compiler

import (
    //"container/list"
    "fmt"
    "sort"
    "strconv"
    "strings"
    "sync"
    "../channel"
    "../defs"
    "../effects"
    "../song"
    "../specs"
    "../targets"
    "../timing"
    "../utils"
    //"../wav"
)

import . "../utils"


const (
    POLARITY_POSITIVE = 0
    POLARITY_NEGATIVE = 1
    
    ELSIFDEF_TAKEN = 2
)

const ALPHANUM = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrtsuvwxyz"


type MmlPattern struct {
    Name string
    Cmds []int
    HasAnyNote bool
    NumTicks int
}

type MmlPatternMap struct {
    keys [] string
    data []*MmlPattern
}

func (m *MmlPatternMap) FindKey(key string) int {
    return utils.PositionOfString(m.keys, key)
}

func (m *MmlPatternMap) Append(key string, pat *MmlPattern) {
    m.keys = append(m.keys, key)
    m.data = append(m.data, pat)
}

func (m *MmlPatternMap) HasAnyNote(key string) bool {
    pos := m.FindKey(key)
    if pos >= 0 {
        return m.data[pos].HasAnyNote
    }
    return false
}

                    
type Compiler struct {
    CurrSong *song.Song
    Songs map[int]*song.Song
    SongNum uint
    ShortFileName string

    //var userDefinedBase int
    implicitAdsrId int
    gbVolCtrl int
    gbNoise int
    enRev int
    octaveRev int
    patName string
    slur bool
    tie bool
    lastWasChannelSelect bool
    //var workDir string 

    dontCompile *GenericStack
    hasElse *GenericStack
    pattern *MmlPattern
    patterns *MmlPatternMap
    keepChannelsActive bool
    callbacks []string
}

func (comp *Compiler) GetShortFileName() string {
    return comp.ShortFileName
}

func (comp *Compiler) GetGbVolCtrlType() int {
    return comp.gbVolCtrl
}

func (comp *Compiler) GetGbNoiseType() int {
    return comp.gbNoise
}

func (comp *Compiler) GetNumSongs() int {
    return len(comp.Songs)
}

func (comp *Compiler) GetSong(num int) defs.ISong {
    return comp.Songs[num]
}

func (comp *Compiler) GetSongs() []defs.ISong {
    songs := make([]defs.ISong, len(comp.Songs))
    keys := []int{}
    for key := range comp.Songs {
        keys = append(keys, key)
    }
    sort.Ints(keys)
    for i, key := range keys {
        songs[i] = comp.Songs[key]
    }
    return songs
}

func (comp *Compiler) GetCallbacks() []string {
    return comp.callbacks
}

func (comp *Compiler) Init(target int) {
    comp.patterns = &MmlPatternMap{}
    
    comp.Songs = map[int]*song.Song{}
    comp.CurrSong = song.NewSong(1, target, comp)
    comp.Songs[1] = comp.CurrSong
    
    comp.dontCompile = NewGenericStack()
    comp.hasElse = NewGenericStack()
    
    comp.dontCompile.Push(0)
    comp.hasElse.Push(false)
    
    comp.lastWasChannelSelect = false
    
    effects.Init()
}


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


/* Check if o is within the range of min and max

 Examples:
   inRange(1, 0, 5) -> 1
   inRange([]int{1,2,3}, 1, 10) -> 1
   inRange([]int{1,2}, []int{0,0}, 5) -> 1
   inRange([]int{1,2}, []int{2,0}, 5) -> 0     (min[0]>o[0])
   inRange([]int{1,2}, []int{0,0,0}, 5) -> 0   (too many elements in min)
*/
func inRange(o interface{}, minimum interface{}, maximum interface{}) bool {
    i, isValScalar := o.(int)
    lo, isMinScalar := minimum.(int)
    hi, isMaxScalar := maximum.(int)
    values, isValSlice := o.([]int)
    los, isMinSlice := minimum.([]int)
    his, isMaxSlice := maximum.([]int)
    
    if isValScalar {
        if isMinScalar && isMaxScalar {
            return (i >= lo && i <= hi)
        } else {
            return false
        }
    } else {
        if isMinScalar && isMaxScalar && isValSlice {
            for _, val := range values {
                if val < lo || val > hi {
                    return false
                }
            }
        } else if isMinScalar && isMaxSlice && isValSlice {
            if len(his) == len(values) {
                for j, _ := range values {
                    if values[j] < lo || values[j] > his[j] {
                        return false
                    }
                }
            } else {
                return false
            }
        } else if isMinSlice && isMaxScalar && isValSlice {
            if len(los) == len(values) {
                for j, _ := range values {
                    if values[j] < los[j] || values[j] > hi {
                        return false
                    }
                }
            } else {
                return false
            }
        } else if isMinSlice && isMaxSlice && isValSlice {
            if len(los) == len(his) && len(his) == len(values) {
                for j, _ := range values {
                    if values[j] < los[j] || values[j] > his[j] {
                        return false
                    }
                }
            } else {
                return false
            }
        } else {
            // Unsupported combination of types
            return false
        }
    }
    
    return true
}


func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}


func (comp *Compiler) writeAllPendingNotes(forceOctChange bool) {
    w := &sync.WaitGroup{}
    for _, chn := range comp.CurrSong.Channels {
        chnCopy := chn
        w.Add(1)
        go func() {
            chnCopy.WriteNote(forceOctChange)
            w.Done()
        }()
    }
    w.Wait()
}


// Parse an expression on the form  SYM op SYM op ...
// Where op is either | or &
// The polarity specifies if this expression is for an IF or IFN, and determines
// how the ops should be interpreted.
// Returns 1 if the expression is true, otherwise 0
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


// Handle commands starting with '#', i.e. a meta command 
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
                        comp.dontCompile.Push(x | comp.dontCompile.PeekInt())
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
                    // ToDo: fix
                    if _, songExists := comp.Songs[int(num)]; !songExists {
                        //m = -1
                        //l = 0

                        for _, chn := range comp.CurrSong.Channels {
                            if chn.IsVirtual() {
                                continue
                            }
                            if chn.Loops.Len() > 0 {
                                utils.ERROR("Open [ loop on channel %s", chn.Name)
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
                        
                        if comp.keepChannelsActive || (len(comp.patName) != 0) {
                            utils.ERROR("Missing }")
                        }
                        
                        /*songNum = o[2]
                        songs[songNum] = repeat({}, length(supportedChannels))
                        songLen[songNum] = repeat(0, length(supportedChannels))
                        hasAnyNote = repeat(0, length(supportedChannels))
                        songLoopLen[songNum] = songLen[songNum]*/                   
   
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
                        effects.DutyMacros.PutInt(num, comp.getEffectFrequency())
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

// Handle definitions of panning macros ("@CS<xy> = {...}")
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
                            effects.PanMacros.PutInt(num, comp.getEffectFrequency())
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


// Handle definitions of modulation macros ("@MOD<xy> = {...}")
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


    
func (comp *Compiler) handleEffectDefinition(effName string, mmlString string, effMap *effects.EffectMap, pred func(*ParamList) bool) {
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
                           effMap.PutInt(num, freq)
                        } else { /*if freq != effMap.GetInt(key) {*/
                            // ToDo: handle the case when we've got an existing identical effect. The references to the new
                            // effects needs to be converted to refer to the old effect.
                           effMap.Append(num, lst)
                           effMap.PutInt(num, freq)
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
}


func (comp *Compiler) applyCmdOnAllActive(cmdName string, cmd []int) {
    w := &sync.WaitGroup{}
    for _, chn := range comp.CurrSong.Channels {
        chnCopy := chn
        w.Add(1)
        go func() {
            if chnCopy.Active {
                chnCopy.AddCmd(cmd)
            }
            w.Done()
        }()
    }
    w.Wait()
}


func (comp *Compiler) applyCmdOnAllActiveFM(cmdName string, cmd []int) {
    for _, chn := range comp.CurrSong.Channels {
        if chn.Active {
            if chn.SupportsFM() {
                chn.AddCmd(cmd)
            } else {
                WARNING(cmdName + " is not supported for channel " + chn.GetName())
            }
        }
    }
}


func (comp *Compiler) applyEffectOnAllActiveSupported(cmdName string, cmd []int, pred func(*channel.Channel) bool,
                                     effMap *effects.EffectMap, num int) {
    for _, chn := range comp.CurrSong.Channels {
        if chn.Active {
            if pred(chn) {
                chn.AddCmd(cmd)
                go effMap.AddRef(num)
            } else {
                WARNING(cmdName + " is not supported for channel " + chn.GetName())
            }
        }
    }
}
                                                        
                                                                        
/*
 * Generates an error if the rune in c isn't a valid channel name.
 */
func (comp *Compiler) assertIsChannelName(c int) {
    if !strings.ContainsRune(comp.CurrSong.Target.GetChannelNames(), rune(c)) {
        ERROR("Unexpected character: " + string(byte(c)))
    }
}


func (comp *Compiler) assertEffectIdExistsAndChannelsActive(name string, eff *effects.EffectMap) (int, int, error) {
    s := Parser.GetNumericString()
    num, err := strconv.Atoi(s)
    idx := -1
    if err == nil {
        idx = eff.FindKey(num)
        if comp.CurrSong.GetNumActiveChannels() == 0 {
            ERROR(name + " requires at least one active channel")
        } else if idx < 0 {
            ERROR("Undefined macro: " + name + s)
        }
    }
    return num, idx, err
}


func (comp *Compiler) assertDisablingEffect(name string, cmd int) {
    c := Parser.Getch()
    s := string(byte(c)) + string(byte(Parser.Getch()))
    if s == "OF" {
        comp.applyCmdOnAllActive(name, []int{cmd, 0})
    } else {
        ERROR("Syntax error: " + name + s)
    }
}


func (comp *Compiler) CompileFile(fileName string) {
    var prevLine int
    var dotOff, tieOff, slurOff bool
    var parserCreationError error
    
    OldParsers.Push(Parser)
    
    Parser,parserCreationError = NewParserState(fileName)
    if parserCreationError != nil {
        ERROR("Failed to read file: " + fileName);
    }
    
    for {
        characterHandled := false

        c := Parser.Getch()
        if c == -1 {
            break
        }
        
        c2 := c
        
        if c == 10 {
            Parser.LineNum++
        }
        
        if Parser.LineNum > prevLine {
            if !comp.keepChannelsActive {
                for i, _ := range comp.CurrSong.Channels {
                    if i < len(comp.CurrSong.Channels)-1 {
                        comp.CurrSong.Channels[i].Active = false
                    }
                }
            }
            prevLine = Parser.LineNum
        }

        /* Meta-commands */
        
        if comp.dontCompile.PeekInt() != 0 {
            if c == '#' {
                comp.handleMetaCommand()
            } 
        } else {
            if c == '#' {
                comp.handleMetaCommand()

            } else if c == '@' {
                for _, chn := range comp.CurrSong.Channels {
                    chn.WriteNote(true)
                }
                m := Parser.Getch()
                s := ""
                if m == '@' {
                    s = "@" + Parser.GetNumericString()
                } else if IsNumeric(m) {
                    Parser.Ungetch()
                    s = Parser.GetNumericString()
                } else {
                    Parser.Ungetch()
                    s = Parser.GetAlphaString()
                    s += Parser.GetNumericString()
                }
                
                if len(s) > 0 {
                    if strings.HasPrefix(s, "@") {
                        num, err := strconv.Atoi(s[1:])
                        if err == nil {
                            idx := effects.DutyMacros.FindKey(num) 
                            numChannels := 0
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    numChannels++
                                    if chn.SupportsDutyChange() > 0 {
                                        idx |= effects.DutyMacros.GetInt(num) * 0x80    // Effect frequency
                                        chn.AddCmd([]int{defs.CMD_DUTMAC, idx})
                                        effects.DutyMacros.AddRef(num)
                                        chn.UsesEffect["DM"] = true
                                    } else {
                                        WARNING("Unsupported command for channel " +
                                                      chn.GetName() + ": @@")
                                    }
                                }
                            }
                            if numChannels == 0 {
                                WARNING("Use of @@ with no channels active")
                            }
                        } else {
                            ERROR("Expected a number: " + s)
                        }
                    } else {
                        num, err := strconv.Atoi(s)
                        // Duty cycle macro definition
                        if err == nil {
                            comp.handleDutyMacDef(num)

                        // Pan macro definition
                        } else if strings.HasPrefix(s, "CS") {
                            comp.handlePanMacDef(s)
                            
                        // Arpeggio definition
                        } else if strings.HasPrefix(s, "EN") {
                            comp.handleEffectDefinition("EN", s, effects.Arpeggios, func(parm *ParamList) bool {
                                    if !parm.IsEmpty() {
                                        if inRange(parm.MainPart, -63, 63) && inRange(parm.LoopedPart, -63, 63) {
                                            return true;
                                        } else {
                                            ERROR("Value of out range (allowed: -63-63): " + parm.Format())
                                        }
                                    } else {
                                        ERROR("Empty list for EN")
                                    }
                                    return false
                                })
                    
                        // Feedback macro definition
                        } else if strings.HasPrefix(s, "FBM") {
                            comp.handleEffectDefinition("FBM", s, effects.FeedbackMacros, func(parm *ParamList) bool {
                                    if !parm.IsEmpty() {
                                        if inRange(parm.MainPart, 0, 7) && inRange(parm.LoopedPart, 0, 7) {
                                            return true;
                                        } else {
                                            ERROR("Value of out range (allowed: 0-7): " + parm.Format())
                                        }
                                    } else {
                                        ERROR("Empty list for FBM")
                                    }
                                    return false
                                })
                                                      
                        // Vibrato definition
                        } else if strings.HasPrefix(s, "MP") {
                            comp.handleEffectDefinition("MP", s, effects.Vibratos, func(parm *ParamList) bool {
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
                                
                        // Amplitude/frequency modulator
                        } else if strings.HasPrefix(s, "MOD") {
                            comp.handleModMacDef(s)                  

                        // Filter definition
                        } else if strings.HasPrefix(s, "FT") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                idx := effects.Filters.FindKey(num)
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        lst, err := Parser.GetList()
                                        if err == nil {
                                            if comp.CurrSong.Target.GetID() == targets.TARGET_C64 {
                                                if len(lst.MainPart) == 3 && len(lst.LoopedPart) == 0 {
                                                    if inRange(lst.MainPart, []int{0, 0, 0}, []int{3, 2047, 15}) {
                                                        effects.Filters.Append(num, lst)
                                                    }
                                                }               
                                            } else {
                                                effects.Filters.Append(num, &utils.ParamList{})
                                            }
                                        } else {
                                            ERROR("Bad FT: " + t)
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }    
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       
                            
                        // Pitch macro definition
                        } else if strings.HasPrefix(s, "EP") {
                            //handlePitchMacDef(s)    
                            comp.handleEffectDefinition("EP", s, effects.PitchMacros, func(parm *ParamList) bool {
                                    if !parm.IsEmpty() {
                                        return true
                                    } else {
                                        ERROR("Empty list for EP")
                                    }
                                    return false
                                })
                                
                        // Portamento definition
                        } else if strings.HasPrefix(s, "PT") {
                            comp.handleEffectDefinition("PT", s, effects.Portamentos, func(parm *ParamList) bool {
                                    if len(parm.MainPart) == 2 && len(parm.LoopedPart) == 0 {
                                        if inRange(parm.MainPart, []int{0, 1}, []int{127, 127}) {
                                            return true
                                        } else {
                                            ERROR("Value of out range: " + parm.Format())
                                        }
                                    } else {
                                        ERROR("Bad PT: " + parm.Format())
                                    }
                                    return false
                                })
                                
                        // Waveform definition (@WT / @WTM)
                        } else if strings.HasPrefix(s, "WT") {
                            num := 0
                            err := error(nil)
                            isWTM := false
                            if strings.HasPrefix(s, "WTM") {
                                isWTM = true
                                num, err = strconv.Atoi(s[3:])
                            } else {
                                num, err = strconv.Atoi(s[2:])
                            }
                            if err == nil {
                                idx := 0
                                if !isWTM {
                                    idx = effects.Waveforms.FindKey(num)
                                } else {
                                    idx = effects.WaveformMacros.FindKey(num)
                                }
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        if isWTM {
                                            //ToDo: fix:  allow_wt_list()
                                        }
                                        lst, err := Parser.GetList()
                                        if comp.CurrSong.Target.SupportsWaveTable() {
                                            if err == nil {
                                                if !isWTM { // Regular @WT
                                                    if len(lst.LoopedPart) == 0 {
                                                        if inRange(lst.MainPart,
                                                                   comp.CurrSong.Target.GetMinWavSample(),
                                                                   comp.CurrSong.Target.GetMaxWavSample()) {
                                                            if len(lst.MainPart) < comp.CurrSong.Target.GetMinWavLength() {
                                                                WARNING(fmt.Sprintf("Padding waveform with zeroes (current length: %d, needs to be at least %d)",
                                                                    len(lst.MainPart), comp.CurrSong.Target.GetMinWavLength()))
                                                                lst.MainPart = append(lst.MainPart, make([]int, comp.CurrSong.Target.GetMinWavLength() - len(lst.MainPart))...)
                                                            
                                                            } else if len(lst.MainPart) > comp.CurrSong.Target.GetMaxWavLength() {
                                                                WARNING("Truncating waveform")
                                                                lst.MainPart = lst.MainPart[0:comp.CurrSong.Target.GetMaxWavLength()]
                                                            }
                                                            //waveforms[2] = append(waveforms[2], t[2])
                                                            effects.Waveforms.Append(num, lst)
                                                        } else {
                                                            ERROR("Waveform data out of range: " + lst.Format())
                                                        }
                                                    } else {
                                                        ERROR("Loops not supported in waveform: " + lst.Format())
                                                    }
                                                } else {            // @WTM
                                                    // ToDo: fix
                                                    /*u = {0,{},{}}
                                                    for j = 2 to 3 do
                                                        for i = 1 to length(t[j]) do
                                                            if and_bits(i, 1) then
                                                                if sequence(t[j][i]) then
                                                                    if length(t[j][i]) = 4 and t[j][i][1] = -1 then
                                                                        u[j] &= t[j][i][4]
                                                                        if assoc_find_key(waveforms, t[j][i][4]) >= 0 then
                                                                            waveforms = assoc_reference(waveforms, t[j][i][4])
                                                                            t[j][i] = assoc_find_key(waveforms, t[j][i][4])
                                                                        } else {
                                                                            ERROR(sprintf("WT%d has not been declared", t[j][i][4]), lineNum)
                                                                        }
                                                                    } else {
                                                                        compiler.ERROR("Expected WT<num>, got " & t[j][i])
                                                                    }
                                                                } else {
                                                                    ERROR("Expected WT<num>, got " & t[j][i])
                                                                }
                                                            } else {
                                                                u[j] &= t[j][i]
                                                                if !inRange(t[j][i], 1, 127) {
                                                                    ERROR(sprintf("Expected a number in the range 1..127, got %d", t[j][i]), lineNum)
                                                                }
                                                            }
                                                        }
                                                    }
                                                    effects.WaveformMacros.Append(num, lst)
                                                    waveformMacros = assoc_insert_extra_data(waveformMacros, o[2], u)*/
                                                }
                                             }else {
                                                ERROR("Bad waveform: " + t)
                                            }
                                        } else {
                                            WARNING("Unsupported command for this target: @WT")
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       

                        } else if strings.HasPrefix(s, "XPCM") {
                            num, err := strconv.Atoi(s[4:])
                            if err == nil {
                                idx := effects.PCMs.FindKey(num)
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        lst, err := Parser.GetList()
                                        if comp.CurrSong.Target.SupportsPCM() {
                                            if err == nil {
                                                if len(lst.LoopedPart) == 0 {
                                                    // ToDo: fix
                                                    /*if len(lst.MainPart) > 1 and sequence(t[2][1]) then
                                                        if not find(':', t[2][1]) and t[2][1][1] != '\\' then
                                                            t[2][1] = workDir & t[2][1]
                                                        }
                                                        if len(lst.MainPart) > 2 {
                                                            t[3] = convert_wav(t[2][1], t[2][2], t[2][3])
                                                        } else {
                                                            t[3] = convert_wav(t[2][1], t[2][2], 100)
                                                        }
                                                        pcms = assoc_append(pcms, o[2], t)
                                                        t = {}
                                                    } else {
                                                        ERROR("Bad XPCM: " & sprint_list(t))
                                                    }*/
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
                                    ERROR("Redefinition of @" + s)
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       
                        
                        } else if strings.HasPrefix(s, "ADSR") {
                            num, err := strconv.Atoi(s[4:])
                            if err == nil {
                                idx := effects.ADSRs.FindKey(num)
                                if idx < 0 {
                                    t := Parser.GetString()
                                    if t == "=" {
                                        lst, err := Parser.GetList()
                                        if err == nil {
                                            if len(lst.LoopedPart) == 0 {
                                                if len(lst.MainPart) == comp.CurrSong.Target.GetAdsrLen() {
                                                    if inRange(lst.MainPart, 0, comp.CurrSong.Target.GetAdsrMax()) {
                                                        effects.ADSRs.Append(num, lst)
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
                                            ERROR("Bad ADSR: " + t)
                                        }
                                    } else {
                                        ERROR("Expected '='")
                                    }
                                } else {
                                    ERROR("Redefinition of @" + s)
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }
                            
    
                        } else if strings.HasPrefix(s, "te") {
                            //t := s[2:]
                            m = Parser.Getch()
                            num1, err1 := strconv.Atoi(s[2:])
                            num2 := 7
                            err2 := error(nil)
                            if m == ',' {
                                num2, err2 = strconv.Atoi(Parser.GetNumericString()) 
                            } else {
                                Parser.Ungetch()
                            }
                                
                            if err1 == nil && err2 == nil {
                                for _, chn := range comp.CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set tone envelope with no active channels")
                                } else {
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsHwToneEnv() != 0 {
                                                if inRange(num1, -chn.SupportsHwToneEnv(), chn.SupportsHwToneEnv()) {
                                                    switch chn.GetChipID() {
                                                    case specs.CHIP_GBAPU: 
                                                        if num1 < 0 {
                                                            num1 = ((abs(num1) - 1) ^ 7) * 16 + 8 + (num2 & 7)
                                                        } else if num1 > 0 {
                                                            num1 = ((num1 - 1) ^ 7) * 16 + (num2 & 7)
                                                        } else {
                                                            num1 = 8
                                                        }
                                                    case specs.CHIP_YM2413: //specs.CHIP_AY_3_8910
                                                        num1 = abs(num1)
                                                    }
                                                    chn.AddCmd([]int{defs.CMD_HWTE, num1})
                                                } else {
                                                    ERROR("Tone envelope number out of range: " + s)
                                                }
                                            } else {
                                                WARNING("Unsupported command for channel " + chn.GetName() + ": @" + s)
                                            }
                                        }
                                    }
                                } 
                            } else {
                                ERROR("Bad tone envelope: " + s)
                            }

                        } else if strings.HasPrefix(s, "es") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                for _, chn := range comp.CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set envelope speed with no active channels")
                                } else if inRange(num, 0, 65535) {
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active {
                                            switch chn.GetChipID() {
                                            case specs.CHIP_AY_3_8910:
                                                num ^= 0xFFFF
                                                chn.AddCmd([]int{defs.CMD_HWES, (num & 0xFF), (num / 0x100)})
                                            default:
                                                WARNING("Unsupported command for this target: es")
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("Bad envelope speed: " + s)
                                }
                            } else {
                                ERROR("Bad tone envelope: " + s)
                            }
                            
                        } else if strings.HasPrefix(s, "ve") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                for _, chn := range comp.CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                m = 0
                                if comp.CurrSong.Target.GetID() == targets.TARGET_GBC {
                                    if num < 0 {
                                        num = (abs(num) - 1) ^ 7
                                    } else if num > 0 {
                                        num = (num - 1) ^ 7
                                        m = 8
                                    } else {
                                        num = 0
                                    }
                                }
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set volume envelope with no active channels")
                                } else if err == nil {
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsHwVolEnv() != 0 {
                                                if inRange(num, -chn.SupportsHwVolEnv(), chn.SupportsHwVolEnv()) {
                                                    switch chn.GetChipID() {
                                                    case specs.CHIP_YM2413:
                                                        num = abs(num) ^ chn.SupportsHwVolEnv()
                                                    case specs.CHIP_SPC:
                                                        if num == -2 {
                                                            num =  0xA0
                                                        } else if num == -1 {
                                                            num = 0x80
                                                        } else if num == 1 {
                                                            num = 0xC0
                                                        } else if num == 2 {
                                                            num = 0xE0
                                                        }
                                                    case specs.CHIP_AY_3_8910:
                                                        num = abs(num)
                                                    }
                                                    chn.AddCmd([]int{defs.CMD_HWVE, num | m})
                                                } else {
                                                    ERROR("Volume envelope value out of range: " + s[2:])
                                                }
                                            } else {
                                                WARNING("Unsupported command for channel " + chn.GetName() + ": @" + s)
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("Bad volume envelope: " + s)
                                }
                            } else {
                                ERROR("Bad volume envelope: " + s)
                            }

                        // Volume macro definition
                        } else if strings.HasPrefix(s, "v") {
                            num, err := strconv.Atoi(s[1:])
                            if err == nil {
                                idx := effects.VolumeMacros.FindKey(num)
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    if len(comp.patName) > 0 {
                                    } else {
                                        if idx < 0 {
                                            t := Parser.GetString()
                                            if t == "=" {
                                                lst, err := Parser.GetList()
                                                if err == nil {
                                                    if inRange(lst.MainPart,
                                                               comp.CurrSong.Target.GetMinVolume(),
                                                               comp.CurrSong.Target.GetMaxVolume()) &&
                                                       inRange(lst.LoopedPart,
                                                               comp.CurrSong.Target.GetMinVolume(),
                                                               comp.CurrSong.Target.GetMaxVolume()) {
                                                        effects.VolumeMacros.Append(num, lst)
                                                        effects.VolumeMacros.PutInt(num, comp.getEffectFrequency())
                                                    } else {
                                                        ERROR("Value out of range: " + lst.Format())
                                                    }
                                                } else {
                                                    ERROR("Bad volume macro: " + t)
                                                }
                                            } else {
                                                ERROR("Expected '='")
                                            }
                                        } else {
                                            ERROR("Redefinition of @" + s)
                                        }
                                    }
                                } else {
                                    if idx < 0 {
                                        ERROR("Undefined macro: @" + s)
                                    } else {
                                        for _, chn := range comp.CurrSong.Channels {
                                            if chn.Active {
                                                if !effects.VolumeMacros.IsEmpty(num) {
                                                    if effects.VolumeMacros.GetInt(num) == defs.EFFECT_STEP_EVERY_FRAME {
                                                        chn.AddCmd([]int{defs.CMD_VOLMAC, idx})
                                                    } else {
                                                        chn.AddCmd([]int{defs.CMD_VOLMAC, idx | 0x80})
                                                    }
                                                    effects.VolumeMacros.AddRef(num)
                                                }
                                            }
                                        }
                                    }
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       
                            

                        // Pulse width macro
                        } else if strings.HasPrefix(s, "pw") {
                            num, err := strconv.Atoi(s[2:])
                            if err == nil {
                                idx := effects.PulseMacros.FindKey(num)
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    if len(comp.patName) > 0 {
                                    } else {
                                        if idx < 0 {
                                            t := Parser.GetString()
                                            if t == "=" {
                                                lst, err := Parser.GetList()
                                                if err == nil {
                                                    if inRange(lst.MainPart, 0, 15) &&
                                                       inRange(lst.LoopedPart, 0, 15) {
                                                        effects.PulseMacros.Append(num, lst)
                                                        effects.PulseMacros.PutInt(num, comp.getEffectFrequency())
                                                    } else {
                                                        ERROR("Value out of range: " + lst.Format())
                                                    }
                                                } else {
                                                    ERROR("Bad pulse width macro: " + t)
                                                }
                                            } else {
                                                ERROR("Expected '='")
                                            }
                                        } else {
                                            ERROR("Redefinition of @" + s)
                                        }
                                    }
                                } else {
                                    if idx < 0 {
                                        ERROR("Undefined macro: @" + s)
                                    } else {
                                        for _, chn := range comp.CurrSong.Channels {
                                            if chn.Active {
                                                if !effects.PulseMacros.IsEmpty(num) {
                                                    chn.AddCmd([]int{defs.CMD_PULMAC, idx})
                                                    effects.PulseMacros.AddRef(num)
                                                    chn.UsesEffect["pw"] = true
                                                }
                                            }
                                        }
                                    }
                                }
                            } else {
                                ERROR("Syntax error: @" + s)
                            }       

                        } else if strings.HasPrefix(s, "q") {
                            num, err := strconv.Atoi(s[1:])
                            if err == nil {
                                for _, chn := range comp.CurrSong.Channels {
                                    chn.WriteNote(true)
                                }
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    WARNING("Trying to set cutoff with no active channels")
                                } else if num >= 0 && num <= 15 {
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active {
                                            chn.CurrentCutoff.Val = num
                                            chn.CurrentCutoff.Typ = defs.CT_FRAMES
                                            /*s = note_length(i, currentLength[i])
                                            if s[1] != currentNoteFrames[i][1] then
                                                currentNoteFrames[i] = s
                                                chn.AddCmd([]int{defs.CMD_LEN})
                                                chn.WriteLength()
                                            end if*/                                            
                                        }
                                    }
                                } else if num < 0 && num >= -15 {
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active {
                                            chn.CurrentCutoff.Val = -num
                                            chn.CurrentCutoff.Typ = defs.CT_NEG_FRAMES
                                            /*s = note_length(i, currentLength[i])
                                            if s[1] != currentNoteFrames[i][1] then
                                                currentNoteFrames[i] = s
                                                chn.AddCmd([]int{defs.CMD_LEN})
                                                chn.WriteLength()
                                            }*/                                         
                                        }
                                    }
                                } else {
                                    ERROR("Bad cutoff: " + s)
                                }
                            } else {
                                ERROR("Bad cutoff: " + s)
                            }
                
                        } else {
                            ERROR("Syntax error: @" + s)
                        }
                    }
                } else {
                    ERROR("Unexpected end of file")
                }


            // Pattern definition/invokation
            } else if c == '\\' {
                for _, chn := range comp.CurrSong.Channels {
                    chn.WriteNote(true)
                }
                s := Parser.GetStringInRange(ALPHANUM)
                Parser.SkipWhitespace()
                m := Parser.Getch()
                if len(s) > 0 {
                    comp.pattern = &MmlPattern{}
                    if m == '(' {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Pattern invokation with no active channels")
                        } else {
                            if comp.CurrSong.Channels[ len(comp.CurrSong.Channels) - 1 ].Active {
                                ERROR("Pattern invokation found inside pattern")
                            } else {
                                t := Parser.GetStringUntil(")")
                                if len(t) > 0 {
                                }
                                Parser.SkipWhitespace()
                                n := Parser.Getch()
                                if n != ')' {
                                    ERROR("Expected ), got " + string(byte(n)))
                                }

                                // Pattern invokation
                                idx := comp.patterns.FindKey(s)
                                if idx >= 0 {
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active {
                                            chn.AddCmd([]int{defs.CMD_JSR, idx - 1})
                                            chn.HasAnyNote = chn.HasAnyNote || comp.patterns.HasAnyNote(s)
                                            /*usesEffect[i] = or_bits(usesEffect[i], {1,1,1,1,1,1})
                                            songLen[songNum][i] += patterns[4][idx]*/
                                        }
                                    }
                                } else {
                                    ERROR("Undefined pattern: " + s)
                                }
                            }
                        }
                    } else if m == '{' {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            comp.patName = s
                            comp.CurrSong.Channels[ len(comp.CurrSong.Channels) - 1 ].Active = true
                            //songs[songNum][length(songs[songNum])] = {}
                        } else {
                            ERROR("Pattern definitions are not allowed while channels are active")
                        }
                    } else {
                        ERROR("Expected ( or {, got " + string(byte(m)))
                    }
                } else {
                    ERROR("Syntax error")
                }
            
            } else if c == '(' {
                if len(comp.patName) > 0 {
                    ERROR("Found ( inside pattern")
                } else if !comp.keepChannelsActive {
                    if comp.CurrSong.GetNumActiveChannels() > 0 && !comp.CurrSong.Channels[ len(comp.CurrSong.Channels) - 1 ].Active {
                        comp.keepChannelsActive = true
                    } else {
                        ERROR("( requires at least one active channel")
                    }
                } else {
                    ERROR("Unexpected (")
                }
                
            } else if c == ')' {
                if comp.keepChannelsActive {
                    comp.keepChannelsActive = false
                } else {
                    ERROR("Unexpected )")
                }
            
            // Macro definition/invokation
            } else if c == '$' {
                comp.writeAllPendingNotes(true)
                s := Parser.GetStringInRange(ALPHANUM)
                Parser.SkipWhitespace()
                m := Parser.Getch()
                t := ""
                if len(s) > 0 {
                    // ToDo: macro := []int{}
                    n := 1
                    if m == '(' {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            l := 1
                            // Read default parameters
                            for {
                                t = Parser.GetStringUntil(",)\t\r\n ")
                                // ToDo: fix
                                //macro = append(macro, {2, t})
                                Parser.SkipWhitespace()
                                n = Parser.Getch()
                                if n == -1 || n == ')' {
                                    break
                                }
                                l += 1
                            }
                            n = 2
                        } else {
                            // Macro invokation
                            for {
                                t = Parser.GetStringUntil(",)\t\r\n ")
                                if len(t) > 0 {
                                    // ToDo: fix
                                    //macro = append(macro, t)
                                }
                                Parser.SkipWhitespace()
                                n = Parser.Getch()
                                if n == -1 || n == ')' {
                                    break
                                }
                            }
                            //ToDo: u := ""
                            //ToDo: defaultParm := []int{}
                            idx := 0 // ToDo: fix:  macros.FindKey(s)
                            if idx >= 0 {
                                // ToDo: fix
                                // Expand the macro
                                /*for i = 1 to length(macros[2][idx]) do
                                    if macros[2][idx][i][1] = 0 then
                                        // This character should be appended as-is
                                        u &= macros[2][idx][i][2]
                                    } else if macros[2][idx][i][1] = 1 then
                                        // This is an argument reference
                                        if macros[2][idx][i][2] > 0 and
                                           macros[2][idx][i][2] <= length(macro) then
                                            u &= macro[macros[2][idx][i][2]]
                                        } else if macros[2][idx][i][2] > 0 and
                                              macros[2][idx][i][2] <= length(defaultParm) then
                                            u &= defaultParm[macros[2][idx][i][2]]
                                        end if
                                    } else if macros[2][idx][i][1] = 2 then
                                        defaultParm = append(defaultParm, macros[2][idx][i][2])
                                    }
                                }
                                if fileDataPos <= len(fileData) then
                                    fileData = fileData[1..fileDataPos-1] & u & fileData[fileDataPos..length(fileData)]
                                } else {
                                    fileData &= u
                                }*/

                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("Trying to invoke a macro with no channels active")
                                }
                            } else {
                                ERROR("Undefined macro: " + s)
                            }
                            n = 0
                        }
                    }

                    if n == 2 {                     
                        Parser.SkipWhitespace()
                        m = Parser.Getch()
                    }
                    
                    // Macro definition
                    if m == '{' && n > 0 {
                        idx := 0  // ToDo: fix:  macros.FindKey(s)
                        if idx < 0 {
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                for n != '}' {
                                    n = Parser.Getch()
                                    if n == '%' {
                                        t = Parser.GetNumericString()
                                        _, err := strconv.Atoi(t)
                                        if err == nil {
                                            // ToDo: fix
                                            //macro = append(macro, {1, num})
                                        } else {
                                            ERROR("Syntax error: " + t)
                                        }
                                        n = Parser.Getch()
                                        if n != '%' {
                                            ERROR("Missing %%")
                                        }
                                    } else if n == '}' || n == -1 {
                                        break
                                    } else if n == 10 {
                                        Parser.LineNum++
                                    } else if n != '\r' {
                                        // ToDo: fix
                                        //macro = append(macro, {0, n})
                                    }
                                }
                            } else {
                                ERROR("Macro definitions are not allowed while channels are active")
                            }
                            // ToDo: fix
                            //macros[1] = append(macros[1], s)
                            //macros[2] = append(macros[2], macro)
                        } else {
                            ERROR("Macro already defined: " + s)
                        }

                    } else if n != 0 {
                        ERROR("Syntax error: " + string(byte(m)))
                    }
                } else {
                    ERROR("Expected an identifier following $")
                }

            // Callback
            } else if c == '!' {
                comp.writeAllPendingNotes(true)
                s := Parser.GetStringUntil("(\t\r\n ")
                Parser.SkipWhitespace()
                n := Parser.Getch()
                if n == '(' {
                    idx := -1
                    for i, cb := range comp.callbacks {
                        if cb == s {
                            idx = i
                            break
                        }
                    }
                    if idx < 0 {
                        comp.callbacks = append(comp.callbacks, s)
                        idx = len(comp.callbacks) - 1
                    }

                    t := Parser.GetStringUntil(")\t\r\n ")
                    num, err := strconv.Atoi(t)
                    if err == nil {
                        if num == 0 {
                            // Never (Off)
                            numChannels := 0
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    numChannels++
                                    chn.AddCmd([]int{defs.CMD_CBOFF})
                                }
                            }
                            if numChannels == 0 {
                                ERROR("Trying to set a callback with no channels active")
                            }
                        } else if num == 1 {
                            // Once
                            numChannels := 0
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    numChannels++
                                    chn.AddCmd([]int{defs.CMD_CBONCE, idx})
                                }
                            }
                            if numChannels == 0 {
                                ERROR("Trying to set a callback with no channels active")
                            }
                        } else {
                            ERROR("Bad callback frequency: " + s)
                        }
                    } else if t == "EVERY-NOTE" {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            comp.applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVNT, idx})
                        }

                    } else if t == "EVERY-VOL-CHANGE" {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            comp.applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVVC, idx})
                        }

                    } else if t == "EVERY-VOL-MIN" {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            comp.applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVVM, idx})
                        }

                    } else if t == "EVERY-OCT-CHANGE" {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("Trying to set a callback with no channels active")
                        } else {
                            comp.applyCmdOnAllActive("Callback", []int{defs.CMD_CBEVOC, idx})
                        }

                    } else {
                        ERROR("Bad callback frequency: " + s)
                    }
                    n = Parser.Getch()
                    if n != ')' {
                        ERROR("Expected ): " + string(byte(n)))
                    }
                } else {
                    ERROR("Expected (: " + string(byte(n)))
                }

            // Multiline comment
            } else if c == '/' {
                comp.writeAllPendingNotes(true)
                m := Parser.Getch()
                if m == '*' {
                    n := 0
                    for n != -1 {
                        n = Parser.Getch()
                        if n == '*' {
                            m = Parser.Getch()
                            if m == '/' {
                                break
                            } else {
                                Parser.Ungetch()
                            }
                        } else if n == 10 {
                            Parser.LineNum++
                        }
                    }
                } else {
                    ERROR("Syntax error: " + string(byte(c)))
                }

            // Beginning of a [..|..]<num> loop
            } else if c == '[' {
                comp.writeAllPendingNotes(true)
                if comp.CurrSong.GetNumActiveChannels() == 0 {
                    // ToDo: treat as error?
                } else {
                    for _, chn := range comp.CurrSong.Channels {
                        if chn.Active {
                            if chn.Tuple.Active {
                                ERROR("Loops not allowed inside {}")
                            }
                            chn.Loops.Push(channel.LoopStackElem{
                                StartPos:   len(chn.Cmds) + 2,
                                StartTicks: chn.Ticks,
                                Unknown:    -1,
                                Skip1Pos:   -1,
                                Skip1Ticks: -1,
                                OrigOctave: chn.CurrentOctave,
                                OctChange:  0,
                                HasOctCmd:  -1,
                                Skip1OctChg:0,
                                Skip1OctCmd:-1,
                            })
                            
                            if chn.Loops.Len() > comp.CurrSong.Target.GetMaxLoopDepth() {
                                ERROR("Maximum nesting level for loops is " +
                                            strconv.FormatInt(int64(comp.CurrSong.Target.GetMaxLoopDepth()), 10))
                            }
                            chn.AddCmd([]int{defs.CMD_LOPCNT, 0})
                        }
                    }
                }
            
            } else if c == '|' {
                comp.writeAllPendingNotes(true)
                if comp.CurrSong.GetNumActiveChannels() == 0 {
                    // ToDo: treat as error?
                } else {
                    for _, chn := range comp.CurrSong.Channels {
                        if chn.Active {
                            if chn.Loops.Len() > 0 {
                                pElem := chn.Loops.Peek()
                                if pElem.Skip1Pos != -1 {
                                    ERROR("Only one | allowed per repeat loop")
                                }
                                pElem.Skip1Pos = len(chn.Cmds) + 2
                                pElem.Skip1Ticks = chn.Ticks
                                chn.AddCmd([]int{defs.CMD_J1, 0, 0})
                            } else {
                                ERROR("Unexpected character: |")
                            }
                        }
                    }
                }
            
            // End of a [..|..]<num> loop
            } else if c == ']' {
                comp.writeAllPendingNotes(true)
                t := Parser.GetNumericString()
                loopCount, err := strconv.Atoi(t)
                for _, chn := range comp.CurrSong.Channels {
                    if chn.Active {
                        elem := chn.Loops.Pop()
                        if elem.StartPos != 0 {
                            if err == nil {
                                if loopCount > 0 {
                                    // Set the value for CMD_LOPCNT
                                    chn.Cmds[elem.StartPos - 1] = loopCount
                                    
                                    chn.AddCmd([]int{defs.CMD_DJNZ, (elem.StartPos & 0xFF), (elem.StartPos / 0x100)})
                                    if elem.Skip1Ticks == -1 {
                                        chn.Ticks += (chn.Ticks - elem.StartTicks) * (loopCount - 1)
                                        if elem.HasOctCmd == -1 {
                                            chn.CurrentOctave = elem.OrigOctave + elem.OctChange * loopCount
                                        } else {
                                            chn.CurrentOctave = elem.HasOctCmd + elem.OctChange
                                        }
                                    } else {
                                        if loopCount < 2 {
                                            ERROR("Loop count must be >= 2 when | is used")
                                        }
                                        chn.Ticks += (chn.Ticks - elem.StartTicks) * (loopCount - 2) +
                                                     (elem.Skip1Ticks - elem.StartTicks)
                                        chn.Cmds[elem.Skip1Pos - 1] = len(chn.Cmds) & 0xFF
                                        chn.Cmds[elem.Skip1Pos] = len(chn.Cmds) / 0x100

                                        if elem.Skip1OctCmd == -1 {
                                            if elem.HasOctCmd == -1 {
                                                chn.CurrentOctave = elem.OrigOctave + elem.OctChange * loopCount +
                                                                                      elem.Skip1OctChg * (loopCount - 1)
                                            } else {
                                                chn.CurrentOctave = elem.HasOctCmd + elem.OctChange
                                            }
                                        } else {
                                            if elem.HasOctCmd == -1 {
                                                chn.CurrentOctave = elem.Skip1OctCmd + elem.OctChange + elem.Skip1OctChg
                                            } else {
                                                chn.CurrentOctave = elem.HasOctCmd + elem.OctChange
                                            }
                                        }
                                    }
                                } else {
                                    ERROR("Bad loop count: " + t)
                                }
                            } else {
                                ERROR("Expected a loop count: " + t)
                            }
                        } else {
                            ERROR("Use of ] with no matching [ on channel " + chn.GetName())
                        }
                    }
                }

            // Octave up/down       
            } else if c == '>' || c == '<' {
                for _, chn := range comp.CurrSong.Channels {
                    chn.WriteNote(true)
                }
                delta := 1
                if c == '<' {
                    delta = -1
                }
                if comp.CurrSong.GetNumActiveChannels() == 0 {
                    WARNING("Trying to change octave with no active channels: " + string(byte(c)))
                } else {
                    for _, chn := range comp.CurrSong.Channels {
                        if chn.Active {
                            if chn.CurrentOctave + delta >= chn.GetMinOctave() &&
                               chn.CurrentOctave + delta <= chn.GetMaxOctave() {
                                chn.CurrentOctave += delta
                                chn.PendingOctChange = delta

                                if chn.Loops.Len() > 0 {
                                    pElem := chn.Loops.Peek()
                                    if pElem.Skip1Pos == -1 {
                                        pElem.OctChange += delta
                                    } else {
                                        pElem.Skip1OctChg += delta
                                    }
                                }
                            
                            } else {
                                ERROR("Octave out of range: " + string(byte(c)) + " (" + 
                                            strconv.FormatInt(int64(chn.CurrentOctave + delta), 10) + ")")
                            }
                        }
                    }
                }
                
            } else if c == 'k' {
                comp.writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set length with no active channels")
                    } else if inRange(num, 1, 256) {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentLength = float64(num)
                                chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff, _ = chn.NoteLength(chn.CurrentLength)
                                chn.AddCmd([]int{defs.CMD_LEN})
                                chn.WriteLength()
                            }
                        }
                    } else {
                        ERROR("Bad length: " + s)
                    }
                } else {
                    ERROR("Bad length: " + s)
                }
                        
            } else if c == 'l' {
                comp.writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set length with no active channels")
                    } else if utils.PositionOfInt(timing.SupportedLengths, num) >= 0 {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentLength = 32.0 / float64(num)
                                chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff, _ = chn.NoteLength(chn.CurrentLength)
                                //fmt.Printf("The new length is %d. %f active frames, %f cutoff frames\n", num, chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff)
                                chn.AddCmd([]int{defs.CMD_LEN})
                                chn.WriteLength()
                            }
                        }
                    } else {
                        ERROR("Bad length: " + s)
                    }
                }else {
                    ERROR("Bad length: " + s)
                }

            // Set octave
            } else if c == 'o' {
                comp.writeAllPendingNotes(false)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set octave with no active channels")
                    } else {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                if inRange(num, chn.GetMinOctave(), chn.GetMaxOctave()) {
                                    chn.CurrentOctave = num
                                    if !chn.Tuple.Active {
                                        chn.AddCmd([]int{defs.CMD_OCTAVE | chn.CurrentOctave})
                                    } else {
                                        chn.Tuple.Cmds = append(chn.Tuple.Cmds, channel.Note{Num: defs.NON_NOTE_TUPLE_CMD,
                                                                                             Frames: float64(defs.CMD_OCTAVE | chn.CurrentOctave),
                                                                                             HasData: true})
                                    }
                                    if chn.Loops.Len() > 0 {
                                        pElem := chn.Loops.Peek()
                                        if pElem.Skip1Ticks == -1 {
                                            pElem.HasOctCmd = chn.CurrentOctave
                                            pElem.OctChange = 0
                                        } else {
                                            pElem.Skip1OctCmd = chn.CurrentOctave
                                            pElem.Skip1OctChg = 0
                                        }
                                    }
                                
                                } else {
                                    ERROR("Octave out of range (" +
                                                fmt.Sprintf(" (%d vs [%d,%d])", num, chn.GetMinOctave(), chn.GetMaxOctave()))
                                }
                            }
                        }
                    } /*else {
                        ERROR("Bad octave:", s)
                    }*/
                } else {
                    ERROR("Bad octave: " + s)
                }

            // Set cutoff
            } else if c == 'q' {
                comp.writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set cutoff with no active channels")
                    } else if num >= 0 && num <= 8 {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentCutoff.Val = num
                                chn.CurrentCutoff.Typ = defs.CT_NORMAL
                                active, cutoff, _ := chn.NoteLength(chn.CurrentLength)
                                if active != chn.CurrentNoteFrames.Active {
                                    chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff = active, cutoff
                                    chn.AddCmd([]int{defs.CMD_LEN})
                                    chn.WriteLength()
                                }
                            }
                        }
                    } else if num < 0 && num >= -8 {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentCutoff.Val = -num
                                chn.CurrentCutoff.Typ = defs.CT_NEG
                                active, cutoff, _ := chn.NoteLength(chn.CurrentLength)
                                if active != chn.CurrentNoteFrames.Active {
                                    chn.CurrentNoteFrames.Active, chn.CurrentNoteFrames.Cutoff = active, cutoff
                                    chn.AddCmd([]int{defs.CMD_LEN})
                                    chn.WriteLength()
                                }
                            }
                        }
                    } else {
                        ERROR("Bad cutoff: " + s)
                    }
                } else {
                    ERROR("Bad cutoff: " + s)
                }

            // Set tempo
            } else if c == 't' {
                comp.writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                num, err := strconv.Atoi(s)
                if err == nil {
                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                        WARNING("Trying to set tempo with no active channels")
                    } else if inRange(num, 0, comp.CurrSong.Target.GetMaxTempo()) {
                        //fmt.Printf("New tempo: %d\n", num)
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                chn.CurrentTempo = num
                            }
                        }
                    } else {
                        ERROR("Bad tempo: " + s)
                    }
                } else {
                    ERROR("Bad tempo: " + s)
                }

            // Set volume
            } else if c == 'v' {
                comp.writeAllPendingNotes(true)
                
                volType := defs.CMD_VOL2
                volDelta := 0
                num := 0
                err := error(nil)
                s := ""
                
                m := Parser.Getch()
                if m == '+' {
                    volType = defs.CMD_VOLUP
                    if Parser.Getch() == '+' {
                        volType = defs.CMD_VOLUPC
                    } else {
                        Parser.Ungetch()
                    }
                    volDelta = 1
                    s = Parser.GetNumericString()
                    if len(s) > 0 {
                        num, err = strconv.Atoi(s)
                        if err == nil {
                            volDelta = num
                        }
                    }
                } else if m == '-' {
                    volType = defs.CMD_VOLDN
                    if Parser.Getch() == '-' {
                        volType = defs.CMD_VOLDNC
                    } else {
                        Parser.Ungetch()
                    }
                    //l = 1   Needed?
                    s = Parser.GetNumericString()
                    if len(s) > 0 {
                        num, err = strconv.Atoi(s)
                        if err == nil {
                            volDelta = num
                        }
                    }
                } else {
                    Parser.Ungetch()
                    s = Parser.GetNumericString()
                    if len(s) > 0 {
                        num, err = strconv.Atoi(s)
                    } else {
                        ERROR("Expected +, - or a number: " + string(byte(m)))
                    }
                }

                if err == nil {
                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                        if len(comp.patName) > 0 {
                            if num >= comp.CurrSong.Target.GetMinVolume() {
                            } else {
                                ERROR("Bad volume: " + strconv.FormatInt(int64(num), 10))
                            }
                        } else {
                            WARNING("Trying to set volume with no active channels")
                        }
                    } else if num >= comp.CurrSong.Target.GetMinVolume() {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                if volType == defs.CMD_VOL2 {
                                    if num <= chn.GetMaxVolume() {
                                        chn.CurrentVolume = num
                                        if chn.SupportsVolumeChange() > 0 {
                                            vol := int((chn.CurrentVolume / chn.GetMaxVolume()) * chn.MachineVolLimit())
                                            if vol <= 0x0F {
                                                chn.AddCmd([]int{defs.CMD_VOL2 | vol})
                                            } else {
                                                chn.AddCmd([]int{defs.CMD_VOLSET, vol})
                                            }
                                        } else {
                                            // TODO: Handle this case (e.g. NES triangle channel)
                                            WARNING("Setting volume on channel " + chn.GetName() + " is not supported")
                                        }
                                    } else {
                                        ERROR("Bad volume: " + strconv.FormatInt(int64(num), 10))
                                    }
                                } else {
                                    if volType == defs.CMD_VOLDN {
                                        volDelta = (-volDelta & 0xFF)
                                        volType = defs.CMD_VOLUP
                                    }
                                    if chn.SupportsVolumeChange() > 0 {
                                        chn.AddCmd([]int{volType, volDelta})
                                    }
                                }
                            }
                        }
                    } else {
                        ERROR("Bad volume: " + strconv.FormatInt(int64(num), 10))
                    }
                } else {
                    ERROR("Bad volume: " + s)
                }

            // Single line comment
            } else if c == ';' {
                comp.writeAllPendingNotes(true)
                n := 0
                for n != -1 {
                    n = Parser.Getch()
                    if n == 10 {
                        Parser.LineNum++
                        break
                    } else if n == '\r' {
                        break
                    } else {
                    }
                }


            // Reads notes (e.g. c d+ a+4 g-1.. e1^8^32 f&f16&f32)
            } else if defs.NoteIndex(c) >= 0 {
                if !comp.slur {
                    comp.writeAllPendingNotes(false)
                }

                comp.tie     = false
                comp.slur    = false
                hasTie      := false
                hasSlur     := false
                hasDot      := false
                firstNote   := -1
                ticks       := -1
                flatSharp   := 0
                
                extraChars  := 0
                n           := c
                note        := 0
                
                for n != -1 {
                    if defs.NoteIndex(n) >= 0 {
                        if extraChars > 0 && !comp.slur {
                            Parser.Ungetch()
                            break
                        }
                        note = defs.NoteIndex(n)
                    } else if n == '&' {
                        if hasTie {
                            ERROR("Trying to use & and ^ in same expression")
                        }
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active && chn.Tuple.Active {
                                WARNING("Command ignored inside tuple: &")
                            }
                        }
                        hasSlur = true
                        comp.slur = true
                        note = -1
                    } else if n == '^' {
                        if hasSlur {
                            ERROR("Trying to use & and ^ in same expression")
                        }
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active && chn.Tuple.Active {
                                WARNING("Command ignored inside tuple: ^")
                            }
                        }
                        hasTie = true
                        comp.tie = true
                    } else if n == '.' {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active && chn.Tuple.Active {
                                WARNING("Command ignored inside tuple: .")
                            }
                        }
                        hasDot = true
                    } else {
                        Parser.Ungetch()
                        break
                    }

                    if (defs.NoteIndex(n) >= 0 && n != 'r' && n != 's') || (extraChars > 0 && note != -1 && comp.slur) {
                        flatSharp = 0
                        ticks = -1 //frames = -1
                        m := Parser.Getch()
                        if m == '+' {
                            flatSharp = 1
                            if defs.NoteVal(note + flatSharp) != -2 {
                                ERROR("Bad note: " + string(byte(n)) + "+")
                            }
                        } else if m == '-' {
                            flatSharp = -1
                            if (note + flatSharp < 1) || defs.NoteVal(note + flatSharp) != -2 {
                                ERROR("Bad note: " + string(byte(n)) + "-")
                            }
                        } else {
                            Parser.Ungetch()
                        }
                    } else if n == 'r' || n == 's' {
                        flatSharp = 0
                    }

                    if firstNote == -1 {
                        firstNote = note + flatSharp
                    } else {
                        if (firstNote != note + flatSharp) && note != -1 {
                            ERROR("Trying to concatenate different notes")
                        }
                    }

                    if defs.NoteIndex(n) >= 0 || comp.tie {
                        s := Parser.GetNumericString()
                        if len(s) > 0 {
                            noteLen, err := strconv.Atoi(s)
                            if err == nil {
                                if utils.PositionOfInt(timing.SupportedLengths, noteLen) >= 0 { 
                                    ticks = 32 / noteLen //frames = 32.0 / float64(noteLen) 
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active && chn.Tuple.Active {
                                            WARNING("Note length ignored inside tuple")
                                        }
                                    }
                                } else {
                                    ERROR("Unsupported length: " + s)
                                }
                            } else {
                                ERROR("Bad length: " + s)
                            }
                        }
                    }
                    if ticks == -1 && comp.tie {
                        ERROR("Expected a length")
                    }

                    tieOff  = false
                    slurOff = false
                    dotOff  = false
                    
                    numChannels := 0
                    for _, chn := range comp.CurrSong.Channels {
                        if chn.Active {
                            if chn.CurrentOctave <= chn.GetMinOctave() &&
                               note > 0 &&
                               note < chn.GetMinNote() {
                                WARNING("The frequency of the note can not be reproduced for this octave: " +
                                              string(byte(defs.NoteVal(note))))
                            }
                            numChannels++
                            if ticks == -1 {
                                ticks = int(chn.CurrentLength)
                            }
                            if extraChars == 0 {
                                if n != 'r' && n != 's' {
                                    chn.CurrentNote = channel.Note{Num: (chn.CurrentOctave - chn.GetMinOctave()) * 12 + note + flatSharp - 1,
                                                                   Frames: float64(ticks),
                                                                   HasData: true}
                                } else if n == 'r' {
                                    chn.CurrentNote = channel.Note{Num: defs.Rest, Frames: float64(ticks), HasData: true}
                                } else {
                                    chn.CurrentNote = channel.Note{Num: defs.Rest2, Frames: float64(ticks), HasData: true}
                                }
                                chn.LastSetLength = float64(ticks)
                            } else {
                                if hasDot {
                                    ticks = int(chn.LastSetLength / 2)
                                    chn.LastSetLength = float64(ticks)
                                    if ticks >= 1 {
                                        chn.CurrentNote.Frames += float64(ticks)
                                    } else {
                                        ERROR("Note length out of range due to dot command")
                                    }
                                    dotOff = true
                                } else if comp.tie {
                                    chn.CurrentNote.Frames += float64(ticks)
                                    tieOff = true
                                } else if comp.slur {
                                    if note != -1 {
                                        if ((chn.CurrentOctave - chn.GetMinOctave()) * 12 + note + flatSharp - 1 == chn.CurrentNote.Num) ||
                                            (chn.CurrentNote.Num == defs.Rest && n == 'r') ||
                                            (chn.CurrentNote.Num == defs.Rest2 && n == 's') {
                                            chn.CurrentNote.Frames += float64(ticks)
                                            chn.LastSetLength = float64(ticks)
                                            slurOff = true
                                        } else {
                                            ERROR("Bad note: " + string(byte(n)))
                                        }
                                    }
                                } else {
                                    ERROR("Bad note: " + string(byte(n)))
                                }
                            }
                        }
                    }
                    
                    if dotOff {
                        hasDot = false
                    }
                    if slurOff {
                        comp.slur = false
                    }
                    if tieOff {
                        comp.tie = false
                    }
                    
                    if numChannels == 0 {
                        ERROR("Trying to play a note with no active channels")
                    }

                    n = Parser.Getch()
                    extraChars++
                }
                comp.tie = false
                comp.slur = false
                hasDot = false
                for _, chn := range comp.CurrSong.Channels {
                    chn.LastSetLength = 0
                }

            } else if c == '{' {
                comp.writeAllPendingNotes(true)
                if comp.CurrSong.GetNumActiveChannels() == 0 {
                    ERROR("{ requires at least one active channel")
                } else {
                    for _, chn := range comp.CurrSong.Channels {
                        if chn.Active {
                            if !chn.Tuple.Active {
                                chn.Tuple.Active = true
                            } else {
                                //  ERROR("Nested tuples are not allowed", lineNum)
                            }
                        }
                    }
                }
                
            } else if c == '}' {
                comp.writeAllPendingNotes(true)
                s := Parser.GetNumericString()
                tupleLen := -1.0
                if len(s) > 0 {
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        if utils.PositionOfInt(timing.SupportedLengths, num) >= 0 { 
                            tupleLen = 32.0 / float64(num) 
                        } else {
                            ERROR("Unsupported length: " + s)
                        }
                    } else {
                        ERROR("Bad length: " + s)
                    }
                }
                if tupleLen == -1 {
                    if len(comp.patName) > 0 {
                        comp.CurrSong.Channels[ len(comp.CurrSong.Channels)-1 ].AddCmd([]int{defs.CMD_RTS})
                        comp.patterns.Append(comp.patName, comp.pattern)
                        /*patterns[1] = append(patterns[1], patName)
                        patterns[2] = append(patterns[2], songs[songNum][length(songs[songNum])])
                        patterns[3] &= hasAnyNote[length(supportedChannels)]
                        patterns[4] &= songLen[songNum][length(supportedChannels)]*/
                        comp.patName = ""
                        //songs[songNum][length(songs[songNum])] = {}
                        comp.CurrSong.Channels[len(comp.CurrSong.Channels)-1].Active = false
                    } else {
                        ERROR("Syntax error: }")
                    }
                } else {
                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                        ERROR("Trying to close a tuple with no active channels")
                    } else {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                if chn.Tuple.Active {
                                    chn.WriteTuple(tupleLen)
                                    chn.Tuple.Active = false
                                }
                            }
                        }
                    }
                }   
                
            } else if strings.ContainsRune(comp.CurrSong.Target.GetChannelNames(), rune(c)) ||
                      strings.ContainsRune("ACDEFKLMOPRSWnpw", rune(c)) {
                comp.writeAllPendingNotes(true)
                
                Parser.Ungetch()    // To make PeekString work
                                  
                if Parser.PeekString(4) == "ADSR" {
                    Parser.SkipN(4)
                    s := ""
                    characterHandled = true
                    //Parser.Ungetch()    // why? is this correct?
                    num := 0
                    err := error(nil)
                    m := Parser.Getch()
                    if m == '(' {
                        // Implicit ADSR declaration
                        Parser.SetListDelimiters("()")
                        lst, err := Parser.GetList()
                        key := -1
                        if err == nil {
                            if len(lst.LoopedPart) == 0 {
                                if len(lst.MainPart) == comp.CurrSong.Target.GetAdsrLen() {
                                    if inRange(lst.MainPart, 0, comp.CurrSong.Target.GetAdsrMax()) {
                                        key = effects.ADSRs.GetKeyFor(lst)
                                        if key == -1 {
                                            effects.ADSRs.Append(comp.implicitAdsrId, lst)
                                            num = comp.implicitAdsrId
                                            comp.implicitAdsrId++
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
                        // Use a previously declared ADSR
                        s = Parser.GetNumericString()
                        num, err = strconv.Atoi(s)
                    }

                    if err == nil {
                        idx := effects.ADSRs.FindKey(num)
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("ADSR requires at least one active channel")
                        } else if idx >= 0 {
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    if chn.SupportsADSR() {
                                        chn.AddCmd([]int{defs.CMD_ADSR, idx})
                                        effects.ADSRs.AddRef(num)
                                    } else {
                                        WARNING("Unsupported command for this channel: ADSR")
                                    }
                                }
                            }
                        } else {
                            ERROR("Undefined macro: ADSR" + s)
                        }
                    } else {
                        ERROR("Expected a number: " + s)
                    }
                } else if Parser.PeekString(2) == "AM" {
                    Parser.SkipN(2)
                    s := Parser.GetNumericString()
                    if len(s) > 0 {
                        characterHandled = true
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if inRange(num, 0, 1) {
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("AM requires at least one active channel")
                                } else {
                                    comp.applyCmdOnAllActiveFM("AM", []int{defs.CMD_HWAM, num})
                                }
                            } else {
                                ERROR("AM out of range")
                            }
                        } else {
                            ERROR("Bad AM: " + s)
                        }
                    } else {
                        //Parser.Ungetch()
                        comp.assertIsChannelName(c)
                    }                       
  

                } else if Parser.PeekString(2) == "CS" {
                    Parser.SkipN(2)
                    characterHandled = true
                    num, idx, err := comp.assertEffectIdExistsAndChannelsActive("CS", effects.PanMacros)

                    if err == nil {
                        if comp.CurrSong.Target.SupportsPan() {
                            idx |= effects.PanMacros.GetInt(num) * 0x80
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    chn.AddCmd([]int{defs.CMD_PANMAC, idx + 1})
                                    effects.PanMacros.AddRef(num)
                                }
                            }
                        } else {
                            WARNING("Unsupported command for this target: CS")
                        }
                    } else {
                        m := Parser.Getch()
                        t := string(byte(m)) + string(byte(Parser.Getch()))
                        if t == "OF" {
                            if comp.CurrSong.Target.SupportsPan() {
                                comp.applyCmdOnAllActive("CS", []int{defs.CMD_PANMAC, 0})
                            }
                        } else {
                            ERROR("Syntax error: CS" + t)
                        }
                    }
                
                } else if c == 'D' {
                    Parser.SkipN(1)
                    m := Parser.Getch()
                    if m == '-' || IsNumeric(m) {
                        characterHandled = true
                        Parser.Ungetch()
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("Detune requires at least one active channel")
                            } else {
                                for _, chn := range comp.CurrSong.Channels {
                                    if chn.Active {
                                        if chn.SupportsDetune() {
                                            if chn.SupportsFM() &&
                                            (chn.GetChipID() == specs.CHIP_YM2612 || chn.GetChipID() == specs.CHIP_YM2151) {
                                                if inRange(num, -3, 3) {
                                                    amount := num
                                                    if amount < 0 {
                                                        amount = (-amount) + 4
                                                    }
                                                    chn.AddCmd([]int{defs.CMD_DETUNE, amount})
                                                } else {
                                                    ERROR("Detune value out of range: " + s)
                                                }
                                            } else {
                                                if inRange(num, -127, 127) {
                                                    chn.AddCmd([]int{defs.CMD_DETUNE, num})
                                                } else {
                                                    ERROR("Detune value out of range: " + s)
                                                }
                                            }
                                        } else {
                                            WARNING("Unsupported command for this channel: D")
                                        }
                                    }
                                }
                            }
                        } else {
                            ERROR("Expected a number: " + s)
                        }
                    } else {
                        Parser.Ungetch()
                        Parser.Ungetch()
                        comp.assertIsChannelName(c)
                    }                       

                } else if c == 'K' {
                    Parser.SkipN(1)
                    m := Parser.Getch()
                    if m == '-' || IsNumeric(m) {
                        characterHandled = true
                        Parser.Ungetch()
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("Transpose requires at least one active channel")
                            } else {
                                if inRange(num, -127, 127) {
                                    comp.applyCmdOnAllActive("K", []int{defs.CMD_TRANSP, num})
                                } else {
                                    ERROR("Transpose value out of range: " + s)
                                }
                            }
                        } else {
                            ERROR("Expected a number: " + s)
                        }
                    } else {
                        Parser.Ungetch()
                        Parser.Ungetch()
                        comp.assertIsChannelName(c)
                    }       
                
                // Arpeggio select ("EN<num>")
                } else if Parser.PeekString(2) == "EN" {
                    Parser.SkipN(2)
                    characterHandled = true
                    num, idx, err := comp.assertEffectIdExistsAndChannelsActive("EN", effects.Arpeggios)

                    if err == nil {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                if !effects.Arpeggios.IsEmpty(num) {
                                    effects.Arpeggios.AddRef(num)
                                    idx |= effects.Arpeggios.GetInt(num) * 0x80
                                    if comp.enRev == 0 {
                                        chn.AddCmd([]int{defs.CMD_ARPMAC, idx})
                                        //ToDo fix: usesEN[1] += 1
                                        chn.UsesEffect["EN"] = true
                                    } else {
                                        chn.AddCmd([]int{defs.CMD_APMAC2, idx})
                                        //ToDo: fix: usesEN[2] += 1
                                        chn.UsesEffect["EN2"] = true
                                    }
                                }
                            }
                        }
                    } else {
                        comp.assertDisablingEffect("EN", defs.CMD_ARPOFF)
                    }

                // Pitch macro select ("PT<num>")
                } else if Parser.PeekString(2) == "EP" {
                    Parser.SkipN(2)
                    characterHandled = true
                    num, idx, err := comp.assertEffectIdExistsAndChannelsActive("EP", effects.PitchMacros)

                    if err == nil {
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                if !effects.PitchMacros.IsEmpty(num) {
                                    idx |= effects.PitchMacros.GetInt(num) * 0x80
                                    chn.AddCmd([]int{defs.CMD_SWPMAC, idx})
                                    effects.PitchMacros.AddRef(num)
                                    chn.UsesEffect["EP"] = true
                                }
                            }
                        }
                    } else {
                        comp.assertDisablingEffect("EP", defs.CMD_SWPMAC)
                    }

                // FM feedback select
                } else if Parser.PeekString(2) == "FB" {
                    Parser.SkipN(2)
                    s := Parser.GetNumericString()
                    if len(s) > 0 {
                        characterHandled = true
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("FB requires at least one active channel")
                            } else {
                                if inRange(num, 0, 7) {
                                    comp.applyCmdOnAllActiveFM("FB", []int{defs.CMD_FEEDBK | num})
                                } else {
                                    ERROR(fmt.Sprintf("Feedback value out of range: %d", num))
                                }
                            }
                        } else {
                            ERROR("Bad feedback value: " +  s)
                        }
                    } else {
                        m := Parser.Getch()
                        if m == 'M' {
                            s = Parser.GetNumericString()
                            if len(s) > 0 {
                                characterHandled = true
                                num, err := strconv.Atoi(s)
                                if err == nil {
                                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                                        ERROR("FBM requires at least one active channel")
                                    } else {
                                        idx := effects.FeedbackMacros.FindKey(num)
                                        if idx >= 0 {
                                            idx |= effects.FeedbackMacros.GetInt(num) * 0x80                
                                            for _, chn := range comp.CurrSong.Channels {
                                                if chn.Active {
                                                    if chn.SupportsFM() {
                                                        //if o[2] >=0 and o[2] <= 7 then
                                                            chn.AddCmd([]int{defs.CMD_FBKMAC, idx})
                                                            effects.FeedbackMacros.AddRef(num)

                                                        //else
                                                        //  ERROR(sprintf("Feedback value out of range: %d", o[2]))
                                                        //end if
                                                    } else {
                                                        WARNING("FBM commands on non-FM channels are ignored")
                                                    }
                                                }
                                            }
                                        } else {
                                            ERROR("Undefined macro: FBM" + s)
                                        }
                                    }
                                } else {
                                    ERROR("Bad feedback value:" + s)
                                }
                            } else {
                                Parser.Ungetch()
                                Parser.Ungetch()
                                comp.assertIsChannelName(c)
                            }
                        } else {
                            Parser.Ungetch()
                            Parser.Ungetch()
                            comp.assertIsChannelName(c)
                        }
                    }

                // Filter select ("FT<num>")
                } else if Parser.PeekString(2) == "FT" {
                    Parser.SkipN(2)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        idx := -1
                        if comp.CurrSong.Target.GetID() != targets.TARGET_AT8 {
                            idx = effects.Filters.FindKey(num)
                        }
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("FT requires at least one active channel")
                        } else if idx >= 0 {
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    if chn.SupportsFilter() {
                                        chn.AddCmd([]int{defs.CMD_FILTER, idx + 1})
                                        effects.Filters.AddRef(num)
                                    } else {
                                        WARNING("Unsupported command for this channel: FT")
                                    }
                                }
                            }
                        } else if comp.CurrSong.Target.GetID() == targets.TARGET_AT8 && num == 0 {
                            comp.applyCmdOnAllActive("FT", []int{defs.CMD_FILTER, 1})
                        } else {
                            ERROR("Undefined macro: FT" + s)
                        }
                    } else if Parser.PeekString(2) == "OF" {
                        Parser.SkipN(2)
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                if chn.SupportsFilter() {
                                    chn.AddCmd([]int{defs.CMD_FILTER, 0})
                                }
                            }
                        }
                    } else {
                        ERROR("Syntax error: expected FT<num> or FTOF")
                    }
         
                // Loop
                } else if c == 'L' {
                    Parser.SkipN(1)
                    if comp.CurrSong.GetNumActiveChannels() == 0 || comp.lastWasChannelSelect {
                        if !strings.ContainsRune(comp.CurrSong.Target.GetChannelNames(), rune(c)) {
                            comp.writeAllPendingNotes(true)
                            WARNING(fmt.Sprintf("Trying to set a loop point with no active channels (last=%d)", comp.lastWasChannelSelect))
                            characterHandled = true
                        }
                    } else {
                        for _, chn := range comp.CurrSong.Channels {
                            chn.WriteNote(true)
                        }
                        characterHandled = true
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                if chn.LoopPoint == -1 {
                                    chn.LoopPoint = len(chn.Cmds)
                                    chn.LoopFrames = chn.Frames
                                    chn.LoopTicks = chn.Ticks
                                } else {
                                    ERROR("Loop point already defined for channel " + chn.GetName())
                                }
                            }
                        }
                    }

        
                // Vibrato select                       
                } else if Parser.PeekString(2) == "MP" {
                    Parser.SkipN(2)
                    characterHandled = true
                    num, idx, err := comp.assertEffectIdExistsAndChannelsActive("MP", effects.Vibratos)

                    if err == nil {
                        idx |= effects.Vibratos.GetInt(num) * 0x80
                        for _, chn := range comp.CurrSong.Channels {
                            if chn.Active {
                                chn.AddCmd([]int{defs.CMD_VIBMAC, idx + 1})
                                effects.Vibratos.AddRef(num)
                                chn.UsesEffect["MP"] = true
                            }
                        }
                    } else {
                        comp.assertDisablingEffect("MP", defs.CMD_VIBMAC);
                    }

                } else if Parser.PeekString(2) == "MF" {
                    Parser.SkipN(2)
                    s := Parser.GetNumericString()
                    if len(s) > 0 {
                        characterHandled = true
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if inRange(num, 0, 15) {
                                if comp.CurrSong.GetNumActiveChannels() == 0 {
                                    ERROR("MF requires at least one active channel")
                                } else {
                                    for _, chn := range comp.CurrSong.Channels {
                                        if chn.Active {
                                            if chn.SupportsFM() || chn.GetChipID() == specs.CHIP_POKEY {
                                                if chn.GetChipID() == specs.CHIP_POKEY {
                                                    if num == 0 {           // 15 kHz
                                                        num = 1
                                                    } else if num == 1 {    // 64 kHz
                                                        num = 2
                                                    } else if num == 2 {
                                                        num = 3             // CPU clock    
                                                    }
                                                }
                                                chn.AddCmd([]int{defs.CMD_MULT, num})
                                            } else {
                                                WARNING("MF ignored for non-FM channel")
                                            }
                                        }
                                    }
                                }
                            } else {
                                ERROR("MF out of range: " + s)
                            } 
                        } else {
                            ERROR("Bad MF: " + s)
                        }
                     } else {
                        Parser.Ungetch()
                        comp.assertIsChannelName(c)
                    }

                // Modulator select ("MOD<num>")
                } else if Parser.PeekString(3) == "MOD" {
                    Parser.SkipN(3)
                    s := Parser.GetNumericString()
                    if len(s) > 0 {
                        characterHandled = true
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            idx := effects.MODs.FindKey(num)
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("MOD requires at least one active channel")
                            } else if idx >= 0 {
                                for _, chn := range comp.CurrSong.Channels {
                                    if chn.Active {
                                        if chn.SupportsFM() || chn.GetChipID() == specs.CHIP_HUC6280 {
                                            chn.AddCmd([]int{defs.CMD_MODMAC, idx + 1})
                                            effects.MODs.AddRef(num)
                                        }
                                    }
                                }
                            } else {
                                ERROR("Undefined macro: MOD" + s)
                            }
                        } else {
                            ERROR("Bad MOD: " + s)
                        }
                    } else {
                        m := Parser.Getch()
                        t := string(byte(m)) + string(byte(Parser.Getch()))
                        if t == "OF" {
                            // Modulator off ("MODOF")
                            characterHandled = true
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    if chn.SupportsFM() || chn.GetChipID() == specs.CHIP_HUC6280 {
                                        chn.AddCmd([]int{defs.CMD_MODMAC, 0})
                                    }
                                }
                            }
                        } else {
                            Parser.Ungetch()
                            Parser.Ungetch()
                            Parser.Ungetch()
                            Parser.Ungetch()
                            comp.assertIsChannelName(c)
                        }
                    }

                } else if c == 'M' {
                    Parser.SkipN(1)
                    m := Parser.Getch()
                    if IsNumeric(m) {
                        // Mode change ("M<num>")
                        characterHandled = true
                        Parser.Ungetch()
                        s := Parser.GetNumericString()
                        num, err := strconv.Atoi(s)
                        if err == nil {
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("M requires at least one active channel")
                            } else if inRange(num, 0, 15) {
                                comp.applyCmdOnAllActive("M", []int{defs.CMD_MODE | num})
                            } else {
                                ERROR("Bad mode: " + s)
                            }
                        } else {
                            ERROR("Bad mode:" + s)
                        }
                    } else {
                        Parser.Ungetch()
                        comp.assertIsChannelName(c)
                    }

                // FM operator select
                } else if Parser.PeekString(2) == "OP" {
                    Parser.SkipN(2)
                    if !comp.lastWasChannelSelect {
                        s := Parser.GetNumericString()
                        if len(s) > 0 {
                            characterHandled = true
                            num, err := strconv.Atoi(s)
                            if err == nil {
                                if inRange(num, 0, 4) {
                                    if comp.CurrSong.GetNumActiveChannels() == 0 {
                                        ERROR("OP requires at least one active channel")
                                    } else {
                                        comp.applyCmdOnAllActiveFM("OP", []int{defs.CMD_OPER | num})
                                    }
                                } else {
                                    ERROR("OP out of range: " + s)
                                }
                            } else {
                                ERROR("Bad OP: " + s)
                            }
                        } else {
                            Parser.Ungetch()
                            comp.assertIsChannelName(c)
                        }
                    } else {
                        Parser.Ungetch()
                        comp.assertIsChannelName(c)
                    }

                // Portamento select
                } else if Parser.PeekString(2) == "PT" {
                    Parser.SkipN(2)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        idx := effects.Portamentos.FindKey(num)
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("PT requires at least one active channel")
                        } else if idx >= 0 {
                            // TODO: set portamento
                            WARNING("PT command not yet implemented")
                        } else {
                            ERROR("Undefined macro: PT" + s)
                        }
                    } else {
                        m := Parser.Getch()
                        t := string(byte(m)) + string(byte(Parser.Getch()))
                        if t == "OF" {
                            // TODO: deactivate portamento
                        } else {
                            ERROR("Syntax error: PT" + s)
                        }
                    }


                // Ring modulation select ("RING<num>")
                } else if Parser.PeekString(4) == "RING" {
                    Parser.SkipN(4)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("RING requires at least one active channel")
                        } else if inRange(num, 0, 1) {
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    if chn.SupportsRingMod() {
                                        chn.AddCmd([]int{defs.CMD_HWRM, num})
                                    } else {
                                        WARNING("Unsupported command for this channel: RING")
                                    }
                                }
                            }
                        } else {
                            ERROR("RING out of range: " + s)
                        }
                    } else {
                        ERROR("Syntax error: RING" + s)
                    }
                        
                // Rate scaling ("RS<num>")
                } else if Parser.PeekString(2) == "RS" && !comp.lastWasChannelSelect {
                    Parser.SkipN(2)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("RS requires at least one active channel")
                        } else if inRange(num, 0, 3) {
                            comp.applyCmdOnAllActiveFM("RS", []int{defs.CMD_RSCALE, num})
                        } else {
                            ERROR("RS out of range: " + s)
                        }
                    } else {
                        ERROR("Syntax error: RS" + s)
                    }

                // Hard sync ("SYNC<num>")
                } else if Parser.PeekString(4) == "SYNC" {
                    Parser.SkipN(4)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("SYNC requires at least one active channel")
                        } else if inRange(num, 0, 1) {
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    if chn.SupportsRingMod() {
                                        chn.AddCmd([]int{defs.CMD_SYNC, num})
                                    } else {
                                        WARNING("Unsupported command for this channel: SYNC")
                                    }
                                }
                            }
                        } else {
                            ERROR("SYNC out of range: " + s)
                        }
                    } else {
                        ERROR("Syntax error: SYNC" + s)
                    }
                
                // SSG Envelope Generator mode ("SSG<num>")
                } else if Parser.PeekString(3) == "SSG" {
                    Parser.SkipN(3)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("SSG requires at least one active channel")
                        } else if inRange(num, 0, 7) {
                            comp.applyCmdOnAllActiveFM("SSG", []int{defs.CMD_SSG, num + 1})
                        } else {
                            ERROR("SSG out of range: " + s)
                        }
                    } else {
                        m := Parser.Getch()
                        t := string(byte(m)) + string(byte(Parser.Getch()))
                        if t == "OF" {
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    if chn.SupportsFM() {
                                        chn.AddCmd([]int{defs.CMD_SSG, 0})
                                    } else {
                                        WARNING("SSG commands not supported for channel " + chn.GetName())
                                    }
                                }
                            }
                        } else {
                            ERROR("Syntax error: SSG" + s)
                        }
                    }

                } else if Parser.PeekString(2) == "WT" {
                    // WT / WTM
                    Parser.SkipN(2)
                    characterHandled = true
                    isWTM := false
                    m := Parser.Getch()
                    if m == 'M' {
                        isWTM = true
                    } else {
                        Parser.Ungetch()
                    }
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil {
                        if !isWTM {
                            idx := effects.Waveforms.FindKey(num)
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("WT requires at least one active channel")
                            } else if idx >= 0 {
                                if comp.CurrSong.GetNumActiveChannels() > 0 {
                                    comp.applyEffectOnAllActiveSupported("WT", []int{defs.CMD_LDWAVE, idx + 1},
                                                                    func(c *channel.Channel) bool { return c.SupportsWaveTable() },
                                                                    effects.Waveforms, num)
                                } else {
                                    WARNING("Trying to use WT with no channels active")
                                }
                            } else if comp.CurrSong.Target.SupportsWaveTable() {
                                ERROR("Undefined macro: WT" + s)
                            }
                        } else {
                            idx := effects.WaveformMacros.FindKey(num)
                            if comp.CurrSong.GetNumActiveChannels() == 0 {
                                ERROR("WTM requires at least one active channel")
                            } else if idx >= 0 {
                                if comp.CurrSong.GetNumActiveChannels() > 0 {
                                    comp.applyEffectOnAllActiveSupported("WTM", []int{defs.CMD_WAVMAC, idx + 1},
                                                                    func(c *channel.Channel) bool { return c.SupportsWaveTable() },
                                                                    effects.WaveformMacros, num)
                                    /*for i = 1 to length(supportedChannels) do
                                        if activeChannels[i] then
                                            if supportsWave[i] then
                                                chn.AddCmd([]int{defs.CMD_WAVMAC, idx})
                                                waveformMacros = assoc_reference(waveformMacros, o[2])
                                            else
                                                WARNING("Unsupported command for channel " & supportedChannels[i] & ": WTM")
                                            end if
                                        end if
                                    end for*/
                                } else {
                                    WARNING("Trying to use WTM with no channels active")
                                }
                            } else if comp.CurrSong.Target.SupportsWaveTable() {
                                ERROR("Undefined macro: WTM" + s)
                            }
                        }
                    } else {
                        m = Parser.Getch()
                        t := string(byte(m)) + string(byte(Parser.Getch()))
                        if t == "OF" {
                            for _, chn := range comp.CurrSong.Channels {
                                if chn.Active {
                                    if chn.SupportsWaveTable() {
                                        chn.AddCmd([]int{defs.CMD_LDWAVE, 0})
                                    } else {
                                        WARNING("WT commands not supported for channel " + chn.GetName())
                                    }
                                }
                            }
                        } else {
                            ERROR("Syntax error: WT" + s)
                        }
                    }

                // Noise speed
                } else if c == 'n' {
                    Parser.SkipN(1)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil && inRange(num, 0, 63) {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("n requires at least one active channel")
                        } else {
                            if comp.CurrSong.GetNumActiveChannels() > 0 {
                                for _, chn := range comp.CurrSong.Channels {
                                    if chn.Active {
                                        switch comp.CurrSong.Target.GetID() {
                                        case targets.TARGET_AST, targets.TARGET_KSS, targets.TARGET_CPC:
                                            chn.AddCmd([]int{defs.CMD_HWNS, num ^ 0x3F})
                                        default:
                                            WARNING("Unsupported command for this channel: n")
                                        }
                                    }
                                }
                            } else {
                                WARNING("Trying to use n with no channels active")
                            }
                        }
                    } else {
                        ERROR("Bad n: " + s)
                    }
                    
                } else if Parser.PeekString(2) == "pw" {
                    Parser.SkipN(2)
                    characterHandled = true
                    s := Parser.GetNumericString()
                    num, err := strconv.Atoi(s)
                    if err == nil && inRange(num, 0, 15) {
                        if comp.CurrSong.GetNumActiveChannels() == 0 {
                            ERROR("pw requires at least one active channel")
                        } else {
                            if comp.CurrSong.GetNumActiveChannels() > 0 {
                                comp.applyCmdOnAllActive("pw", []int{defs.CMD_PULSE | num})
                            } else {
                                WARNING("Trying to use pw with no channels active")
                            }
                        }
                    } else {
                        ERROR("Bad pw: " + s)
                    }
                
                } else if c == 'w' {
                    Parser.SkipN(1)
                    wrType := defs.CMD_WRMEM
                    m := Parser.Getch()
                    if m == '(' {
                        wrType = defs.CMD_WRPORT
                    } else {
                        Parser.Ungetch()
                    }
                    s := Parser.GetNumericString()
                    if len(s) > 0 {
                        addr, err := strconv.Atoi(s)
                        if err == nil {
                            if inRange(addr, 0, 0xFFFF) {
                                m = Parser.Getch()
                                if wrType == defs.CMD_WRPORT {
                                    if m != ')' {
                                        ERROR("Expected ')'")
                                    }
                                    m = Parser.Getch()
                                }
                                if m != ',' {
                                    ERROR("Expected ','")
                                }
                                s = Parser.GetNumericString()
                                if len(s) == 0 {
                                    ERROR("Missing second argument for w")
                                }
                                val, err := strconv.Atoi(s)
                                if err != nil {
                                    ERROR("Bad second argument for w")
                                }
                                for _, chn := range comp.CurrSong.Channels {
                                    if chn.Active {
                                        go chn.AddCmd([]int{wrType,
                                                         (addr & 0xFF),
                                                         (addr / 0x100),
                                                         (val & 0xFF)})
                                    }
                                }
                                characterHandled = true
                            } else {
                                ERROR("Memory address out of range: " + s)
                            }
                        } else {
                            ERROR("Bad first argument for w: " + s)
                        }
                    } else {
                        ERROR("Bad first argument for w")
                    }
                }
                
                if !characterHandled {
                    Parser.SkipN(1)
                    comp.assertIsChannelName(c)
                    //fmt.Printf("Char not handled: '%c' on line %d (c3='%c')\n", c, Parser.LineNum, c3)
                    for _, chn := range comp.CurrSong.Channels {
                        if !comp.lastWasChannelSelect {
                            chn.Active = false
                        }
                        if chn.GetName() == string(byte(c)) {
                            chn.Active = true
                        }
                    }
                }
            } else if strings.ContainsRune("\t\r\n ", rune(c)) {
                //fmt.Printf("Whitespace on line %d\n", Parser.LineNum)
                for _, chn := range comp.CurrSong.Channels {
                    chn.WriteNote(c == 10)
                }
            } else {
                if c == '%' {
                    ERROR("Unexpected character: %%")
                } else {
                    ERROR("Unexpected character: " + string(byte(c)))
                }
            }
            
            comp.lastWasChannelSelect = strings.ContainsRune(comp.CurrSong.Target.GetChannelNames(), rune(c2))
            /*if comp.lastWasChannelSelect {
                fmt.Printf("%c on line %d\n", c2, Parser.LineNum)
            } else {
                fmt.Printf("'%c' on line %d\n", c2, Parser.LineNum)
            }*/          
        }               
                
    }
    
    comp.writeAllPendingNotes(true)
    Parser = OldParsers.PopParserState()
}

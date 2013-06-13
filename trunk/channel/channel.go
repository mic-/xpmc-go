/*
 * Package channel
 *
 * Part of XPMC.
 * Contains data/functions representing individual channels
 * of a song.
 *
 * /Mic, 2013
 */
 
package channel

import (
    "container/list"
    "math"
    "../defs"
    "../specs"
    "../timing"
    "../utils"
)


type Note struct {
    Num int
    Frames float64
    HasData bool
}


type LoopStackElem struct {
    StartPos int    /* The index within the channel's command sequence where the first command
                       following the '[' occurs. */
    StartTicks int  /* The number of ticks (32nd notes) that the channel contains at the point where
                       the loop starts. */
    Unknown int
    Skip1Pos int    /* The index within the channel's command sequence where the first command
                       following the '|' occurs. */
    Skip1Ticks int  /* The number of ticks that the channel contains at the point where the part of
                       loop following the '|' starts. */
    OrigOctave int  /* The current octave at the point where the loop starts. */
    OctChange int   /* The relative octave change within the first part of the loop. */
    HasOctCmd int   /* Set if there's an absolute octave command ('o') within the first part of the loop. */
    Skip1OctChg int /* The relative octave change within the part of the loop following the '|'. */
    Skip1OctCmd int /* Set if there's an absolute octave command ('o') within the part of the loop
                       following the '|'. */
}

type LoopStack struct {
    data *list.List
}

func (s *LoopStack) Push(e LoopStackElem) {
    _ = s.data.PushBack(e)
}

func (s *LoopStack) Pop() LoopStackElem {
    e := s.data.Back()
    return s.data.Remove(e).(LoopStackElem)
}

func (s *LoopStack) Peek() LoopStackElem {
    e := s.data.Back()
    return e.Value.(LoopStackElem)
}

func (s *LoopStack) Len() int {
    return s.data.Len()
}

func NewLoopStack() *LoopStack {
    return &LoopStack{list.New()}
}


type Channel struct {
    Name string             // The channel's name ("A", "B", "C", etc)
    Num int                 // The channel's number (0, 1, 2, etc)
    Frames float64          // The total number of frames used by this channel given the current refresh rate
    LoopFrames float64
    LoopPoint int
    Ticks int
    LastSetLength float64
    CurrentTempo int        // The channel's tempo, in BPM
    CurrentNote Note
    CurrentOctave int       // The currently set octave for this channel
    CurrentVolume int       // The currently set volume for this channel
    Tuple struct {
        Cmds []Note
        HasData bool
        Active bool
    }
    CurrentLength float64   // Length in 32nd notes
    CurrentNoteFrames struct {
        Active, Cutoff float64
    }
    CurrentCutoff struct {
        Typ int
        Val float64
    }
    PendingOctChange int
    HasAnyNote bool         // Wheter any notes have been added on this channel
    Active bool             // Is this channel currently active?
    Cmds []int
    UsesEffect map[string]bool
    Loops *LoopStack
    ChannelSpecs *specs.Specs
}


func NewChannel() *Channel {
    return &Channel{}
}

func (chn *Channel) SupportsADSR() bool {
    if len(chn.ChannelSpecs.ADSR) > 0 {
        if chn.ChannelSpecs.ADSR[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsDetune() bool {
    if len(chn.ChannelSpecs.Detune) > 0 {
        if chn.ChannelSpecs.Detune[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsDutyChange() int {
    if len(chn.ChannelSpecs.Duty) > 0 {
        return chn.ChannelSpecs.Duty[chn.Num];
    }
    return -1
}

func (chn *Channel) SupportsFilter() bool {
    if len(chn.ChannelSpecs.Filter) > 0 {
        if chn.ChannelSpecs.Filter[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsFM() bool {
    if len(chn.ChannelSpecs.FM) > 0 {
        if chn.ChannelSpecs.FM[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsRingMod() bool {
    if len(chn.ChannelSpecs.RingMod) > 0 {
        if chn.ChannelSpecs.RingMod[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsWaveTable() bool {
    if len(chn.ChannelSpecs.WaveTable) > 0 {
        if chn.ChannelSpecs.WaveTable[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsHwToneEnv() int {
    // ToDo: implement
    return 0
}

func (chn *Channel) SupportsHwVolEnv() int {
    // ToDo: implement
    return 0
}

func (chn *Channel) SupportsVolumeChange() int {
    if len(chn.ChannelSpecs.VolChange) > 0 {
        return chn.ChannelSpecs.VolChange[chn.Num]
    }
    return 0
}


func (chn *Channel) GetName() string {
    return chn.Name
}

func (chn *Channel) GetMaxOctave() int {
    if len(chn.ChannelSpecs.MaxOct) > 0 {
        return chn.ChannelSpecs.MaxOct[chn.Num]
    }
    return 0
}

func (chn *Channel) GetMinOctave() int {
    if len(chn.ChannelSpecs.MinOct) > 0 {
        return chn.ChannelSpecs.MinOct[chn.Num]
    }
    return 0
}

func (chn *Channel) GetMinNote() int {
    if len(chn.ChannelSpecs.MinNote) > 0 {
        return chn.ChannelSpecs.MinNote[chn.Num]
    }
    return 0
}

func (chn *Channel) GetMaxVolume() int {
    if len(chn.ChannelSpecs.MaxVol) > 0 {
        return chn.ChannelSpecs.MaxVol[chn.Num]
    }
    return 0
}

func (chn *Channel) MachineVolLimit() int {
    // ToDo: implement
    return 0
}

func (chn *Channel) SetMaxVolume(maxVol int) {
    if len(chn.ChannelSpecs.MaxVol) > 0 {
        chn.ChannelSpecs.MaxVol[chn.Num] = maxVol
    }
}

func (chn *Channel) SetMaxOctave(maxOct int) {
    if len(chn.ChannelSpecs.MaxOct) > 0 {
        chn.ChannelSpecs.MaxOct[chn.Num] = maxOct
    }
}

func (chn *Channel) AddCmd(cmds []int) {
    chn.Cmds = append(chn.Cmds, cmds...)
}


// Calculate the length of a note in frames based on its length in 32nd notes, the current tempo,
// the current playback speed and the current note cutoff setting.
func (chn *Channel) NoteLength(len float64) (frames, cutoffFrames, scaling float64) {
    var length32 int

    frames = (timing.UpdateFreq * 60.0) / float64(chn.CurrentTempo) // frames per quarternote
    
    if timing.UseFractionalDelays {
        scaling = 256.0
        length32 = int((frames / 8) * scaling)  // frames per 32nd note, scaled by 256
        frames = math.Floor(float64(length32) * len)
        
        if (chn.CurrentCutoff.Typ == defs.CT_FRAMES ||
            chn.CurrentCutoff.Typ == defs.CT_NEG_FRAMES) {
            cutoffFrames = math.Min(chn.CurrentCutoff.Val * scaling, frames)
        } else {
            cutoffFrames = (frames * (8 - chn.CurrentCutoff.Val)) / 8
        }
        frames = math.Floor(frames) - math.Floor(cutoffFrames)
        if (frames < 0) {
            utils.ERROR("Note has negative length")
        }
    } else {
        scaling = 1.0
        length32 = int(frames / 8)  // frames per 1/32 note
        frames = float64(length32) * len
        if (chn.CurrentCutoff.Typ == defs.CT_FRAMES ||
            chn.CurrentCutoff.Typ == defs.CT_NEG_FRAMES) {
            cutoffFrames = math.Min(chn.CurrentCutoff.Val, frames)
        } else {
            cutoffFrames = math.Floor((frames * (8 - chn.CurrentCutoff.Val)) / 8)
        }
        frames -= cutoffFrames
    }
    
    if (frames < 1.0 * scaling) {
        utils.WARNING("Note is too short, will not be heard")
    } else if (int(frames) > int(0x3FFF * scaling + (scaling - 1))) {
        utils.WARNING("Note is too long, cutting at 16383 frames")
        frames = 0x3FFF * scaling + (scaling - 1)
    }
    
    return
}



func (chn *Channel) WriteLength() {
    
    if timing.UseFractionalDelays {
        len1 := int(math.Floor(chn.CurrentNoteFrames.Active))
        len2 := int(math.Floor(chn.CurrentNoteFrames.Cutoff))
        scaling := 256.0
        
        timing.UpdateDelayMinMax(len1)
        timing.UpdateDelayMinMax(len2)

        if math.Floor(chn.CurrentNoteFrames.Active) > 127.0 * scaling {
            chn.AddCmd([]int{(len1 / 0x8000) | 0x80,
                             (len1 / 0x100) & 0x7F,
                             (len1 & 0xFF)})
        } else {
            chn.AddCmd([]int{(len1 / 0x100),
                             (len1 & 0xFF)})
        }
        
    } else {
        //w1 := int(math.Floor(chn.CurrentNoteFrames.Active))
        //w2 := int(math.Floor(chn.CurrentNoteFrames.Active))

        /*if w1 < shortestDelay[1] then
            shortestDelay[1] = w1
        end if
        if w2 >= 1 and w2 < shortestDelay[1] then
            shortestDelay[1] = w2
        end if
        if w1 > longestDelay then
            longestDelay = w1
        end if
        if w2 > longestDelay then
            longestDelay = w2
        end if

        if w1 > 127 then
            ADD_CMD(chn, {or_bits(floor(w1 / 128), #80),
                    and_bits(w1, #7F)})
        elsif w1 >= 1 then
            ADD_CMD(chn, {w1})
        end if*/
    }
}


func (chn *Channel) WriteNoteAndLength(note int, noteLen int, minLen int, scaling float64) {
    if noteLen > int(127.0 * scaling) {
        chn.AddCmd([]int{note,
                         (noteLen / 0x8000) | 0x80,
                         (noteLen / 0x100) & 0x7F,
                         (noteLen & 0xFF)})
    } else if noteLen >= minLen {
        chn.AddCmd([]int{note, 
                         (noteLen / 0x100),
                         (noteLen & 0xFF)})
    }
}

                                         
// Output all pending notes on this channel to the current song     
func (chn *Channel) WriteNote(forceOctChange bool) {
    var frames, cutoffFrames, scaling float64
    var len1, len2 int
    
    if chn.CurrentNote.HasData {
        if !chn.Tuple.HasData {
            chn.Frames += chn.CurrentNote.Frames

            frames, cutoffFrames, scaling = chn.NoteLength(chn.CurrentNote.Frames)
                                                    
            if chn.CurrentNote.Num == defs.Rest {
                chn.CurrentNote.Num = defs.CMD_REST
            } else if chn.CurrentNote.Num == defs.Rest2 {
                chn.CurrentNote.Num = defs.CMD_REST2
            } else {
                chn.CurrentNote.Num = chn.CurrentNote.Num % 12
            }

            if chn.PendingOctChange == 1 {
                chn.CurrentNote.Num |= defs.CMD_OCTUP
            } else if chn.PendingOctChange == -1 {
                chn.CurrentNote.Num |= defs.CMD_OCTDN
            }
            
            if timing.UseFractionalDelays {
                len1 = int(frames) 
                len2 = int(cutoffFrames)
                timing.UpdateDelayMinMax(len1)
                timing.UpdateDelayMinMax(len2)

                if chn.CurrentCutoff.Typ == defs.CT_NORMAL ||
                   chn.CurrentCutoff.Typ == defs.CT_FRAMES {
                    if math.Floor(frames) == chn.CurrentNoteFrames.Active {
                        if chn.PendingOctChange == 0 {
                            chn.AddCmd([]int{chn.CurrentNote.Num | defs.CMD_NOTE2})
                        } else {
                            chn.AddCmd([]int{chn.CurrentNote.Num})
                        }
                    } else {
                        if chn.PendingOctChange != 0 {
                            chn.AddCmd([]int{defs.CMD_OCTAVE | chn.CurrentOctave})
                            chn.CurrentNote.Num &= 0x0F
                            chn.PendingOctChange = 0
                        }
                        chn.WriteNoteAndLength(chn.CurrentNote.Num, len1, 0, scaling)
                    }

                    chn.WriteNoteAndLength(defs.CMD_REST, len2, 1, scaling)
                } else {
                    chn.WriteNoteAndLength(defs.CMD_REST, len2, 1, scaling)
                
                    if math.Floor(frames) == chn.CurrentNoteFrames.Active {
                        if chn.PendingOctChange == 0 {
                            chn.AddCmd([]int{chn.CurrentNote.Num | defs.CMD_NOTE2})
                        } else {
                            chn.AddCmd([]int{chn.CurrentNote.Num})
                        }
                    } else {
                        if chn.PendingOctChange != 0 {
                            chn.AddCmd([]int{defs.CMD_OCTAVE | chn.CurrentOctave})
                            chn.CurrentNote.Num &= 0x0F
                            chn.PendingOctChange = 0
                        }
                        chn.WriteNoteAndLength(chn.CurrentNote.Num, len1, 0, scaling)
                    }
                }
            } else {
                len1 = int(frames)
                len2 = int(cutoffFrames)
                if len1 < timing.ShortestDelay.Lo {
                    timing.ShortestDelay.Lo = len1
                }
                if len2 >= 1 && len2 < timing.ShortestDelay.Lo {
                    timing.ShortestDelay.Lo = len2
                }
                if len1 > timing.LongestDelay {
                    timing.LongestDelay = len1
                }
                if len2 > timing.LongestDelay {
                    timing.LongestDelay = len2
                }
                if len1 > 127 {
                    chn.AddCmd([]int{chn.CurrentNote.Num,
                                     (len1 / 128) | 0x80,
                                     (len1 & 0x7F)})
                } else if len1 >= 1 {
                    chn.AddCmd([]int{chn.CurrentNote.Num, len1})
                }
                if len2 > 127 {
                    chn.AddCmd([]int{defs.CMD_REST,
                                     (len2 / 128) | 0x80,
                                     (len2 & 0x7F)})
                } else if len2 >= 1 {
                    chn.AddCmd([]int{defs.CMD_REST, len2})
                }
            }  // if timing.UseFractionalDelays
        } else {
            if chn.PendingOctChange == 1 {
                chn.CurrentNote.Num |= defs.CMD_OCTUP
            } else if chn.PendingOctChange == -1 {
                chn.CurrentNote.Num |= defs.CMD_OCTDN
            }
            chn.Tuple.Cmds = append(chn.Tuple.Cmds, chn.CurrentNote)
        }

        chn.CurrentNote.HasData = false
        chn.PendingOctChange = 0
        chn.HasAnyNote = true
            
    } else if forceOctChange && chn.PendingOctChange != 0 {
        if !chn.Tuple.HasData {
            chn.AddCmd([]int{defs.CMD_OCTAVE | chn.CurrentOctave})
        } else {
            chn.Tuple.Cmds = append(chn.Tuple.Cmds,
                                    Note{defs.NON_NOTE_TUPLE_CMD, float64(defs.CMD_OCTAVE | chn.CurrentOctave), true})
        }
        chn.PendingOctChange = 0
    }
}



func (chn *Channel) WriteTuple(tupleLen float64) {
    var totalFrames, frames, cutoffFrames, scaling float64
    var w1, w2, skipBytes int

    chn.Frames += tupleLen
    
    totalFrames, cutoffFrames, scaling = chn.NoteLength(tupleLen)
    
    frames = totalFrames / float64(len(chn.Tuple.Cmds))
    w1 = int(frames) 
    for i, _ := range chn.Tuple.Cmds {
        if chn.Tuple.Cmds[i].Num != defs.NON_NOTE_TUPLE_CMD {
            chn.Tuple.Cmds[i].Frames = float64(w1)
        }
    }
    for i := 0; i < int(totalFrames) - (w1 * len(chn.Tuple.Cmds)); i++ {
        for j := 0; j < len(chn.Tuple.Cmds)-1; j++ {
            if chn.Tuple.Cmds[(i + j) % len(chn.Tuple.Cmds)].Num != defs.NON_NOTE_TUPLE_CMD {
                chn.Tuple.Cmds[(i + j) % len(chn.Tuple.Cmds)].Frames += 1
                break
            }
        }
    }
    w2 = int(cutoffFrames) 
    
    if chn.CurrentCutoff.Typ == defs.CT_NEG ||
       chn.CurrentCutoff.Typ == defs.CT_NEG_FRAMES {
        chn.WriteNoteAndLength(defs.CMD_REST, w2, 1, scaling)
    }
    
    skipBytes = 0
    
    for i, _ := range chn.Tuple.Cmds {
        if skipBytes > 0 {
            skipBytes--
        } else {
            if chn.Tuple.Cmds[i].Num == defs.NON_NOTE_TUPLE_CMD {
                chn.AddCmd([]int{int(chn.Tuple.Cmds[i].Frames)})
                skipBytes = 1           
            } else if chn.Tuple.Cmds[i].Num == defs.Rest {
                chn.Tuple.Cmds[i].Num = defs.CMD_REST
            } else if chn.Tuple.Cmds[i].Num == defs.Rest2 {
                chn.Tuple.Cmds[i].Num = defs.CMD_REST2
            } else {
                chn.Tuple.Cmds[i].Num = chn.Tuple.Cmds[i].Num % 12
            }

            if skipBytes > 0 {
                skipBytes--
            } else {
                if timing.UseFractionalDelays {
                    if int(chn.Tuple.Cmds[i].Frames / 256.0) > 127 {
                        chn.AddCmd([]int{chn.Tuple.Cmds[i].Num,
                                         int(chn.Tuple.Cmds[i].Frames / 0x8000) | 0x80,
                                         (int(chn.Tuple.Cmds[i].Frames / 0x100) & 0x7F),
                                         (int(chn.Tuple.Cmds[i].Frames) & 0xFF)})
                    } else {
                        chn.AddCmd([]int{chn.Tuple.Cmds[i].Num,
                                         int(chn.Tuple.Cmds[i].Frames / 0x100),
                                         int(chn.Tuple.Cmds[i].Frames) & 0xFF})
                    }

                    timing.UpdateDelayMinMax(int(chn.Tuple.Cmds[i].Frames))
                } else {
                }
            }
        }
    }
    

    if w2 >= 1 {
        timing.UpdateDelayMinMax(w2)
    }
    
    if chn.CurrentCutoff.Typ == defs.CT_NORMAL ||
       chn.CurrentCutoff.Typ == defs.CT_FRAMES {
        chn.WriteNoteAndLength(defs.CMD_REST, w2, 1, scaling)
    }
}


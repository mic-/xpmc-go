/*
 * Package channel
 *
 * Part of XPMC.
 * Contains data/functions representing individual channels
 * of a song.
 *
 * /Mic, 2013-2014
 */
 
package channel

import (
    "math"
    "../defs"
    "../timing"
    "../utils"
)


type Note struct {
    Num int
    Frames float64
    HasData bool
}


type Channel struct {
    Name string                 // The channel's name ("A", "B", "C", etc)
    Num int                     // The channel's number (0, 1, 2, etc)
    Ticks int                   // The total number of ticks (32nd notes) used by this channel
    Frames float64              // The total number of frames used by this channel given the current refresh rate
    LoopFrames float64
    LoopTicks int
    LoopPoint int
    LastSetLength float64
    CurrentTempo int            // The channel's tempo, in BPM
    CurrentNote Note
    CurrentOctave int           // The currently set octave for this channel
    CurrentVolume int           // The currently set volume for this channel
    Tuple struct {
        Cmds []Note
        HasData bool
        Active bool
    }
    CurrentLength float64       // Length in 32nd notes
    CurrentNoteFrames struct {
        Active, Cutoff float64
    }
    CurrentCutoff struct {
        Typ int
        Val int         
    }
    PendingOctChange int
    HasAnyNote bool             // Whether any notes have been added on this channel
    Active bool                 // Is this channel currently active?
    Cmds []int
    UsesEffect map[string]bool
    Loops *LoopStack
    ChannelSpecs defs.ISpecs
    IsVirtualChannel bool
}


func NewChannel() *Channel {
    chn := &Channel{}
    chn.LoopPoint = -1
    chn.CurrentTempo = 125
    chn.CurrentOctave = 4
    chn.CurrentLength = 8 // ToDo: correct init value?
    chn.UsesEffect = map[string]bool{}
    chn.Loops = NewLoopStack()
    chn.CurrentCutoff.Typ = defs.CT_NORMAL
    chn.CurrentCutoff.Val = 8
    // Note: some fields are initialized by song.NewSong()
    return chn
}

/*****/

func (chn *Channel) SupportsADSR() bool {
    if len(chn.ChannelSpecs.GetADSR()) > 0 {
        if chn.ChannelSpecs.GetADSR()[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsDetune() bool {
    if len(chn.ChannelSpecs.GetDetune()) > 0 {
        if chn.ChannelSpecs.GetDetune()[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsDutyChange() int {
    if len(chn.ChannelSpecs.GetDuty()) > 0 {
        return chn.ChannelSpecs.GetDuty()[chn.Num];
    }
    return -1
}

func (chn *Channel) SupportsFilter() bool {
    if len(chn.ChannelSpecs.GetFilter()) > 0 {
        if chn.ChannelSpecs.GetFilter()[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsFM() bool {
    if len(chn.ChannelSpecs.GetFM()) > 0 {
        if chn.ChannelSpecs.GetFM()[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsRingMod() bool {
    if len(chn.ChannelSpecs.GetRingMod()) > 0 {
        if chn.ChannelSpecs.GetRingMod()[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsWaveTable() bool {
    if len(chn.ChannelSpecs.GetWaveTable()) > 0 {
        if chn.ChannelSpecs.GetWaveTable()[chn.Num] != 0 {
            return true
        }
    }
    return false
}

func (chn *Channel) SupportsHwToneEnv() int {
    if len(chn.ChannelSpecs.GetToneEnv()) > 0 {
        return chn.ChannelSpecs.GetToneEnv()[chn.Num]
    }
    return 0
}

func (chn *Channel) SupportsHwVolEnv() int {
    if len(chn.ChannelSpecs.GetVolEnv()) > 0 {
        return chn.ChannelSpecs.GetVolEnv()[chn.Num]
    }
    return 0
}

func (chn *Channel) SupportsVolumeChange() int {
    if len(chn.ChannelSpecs.GetVolChange()) > 0 {
        return chn.ChannelSpecs.GetVolChange()[chn.Num]
    }
    return 0
}


func (chn *Channel) GetMaxOctave() int {
    if len(chn.ChannelSpecs.GetMaxOct()) > 0 {
        return chn.ChannelSpecs.GetMaxOct()[chn.Num]
    }
    return 0
}

func (chn *Channel) GetMinOctave() int {
    if len(chn.ChannelSpecs.GetMinOct()) > 0 {
        return chn.ChannelSpecs.GetMinOct()[chn.Num]
    }
    return 0
}

func (chn *Channel) GetMinNote() int {
    if len(chn.ChannelSpecs.GetMinNote()) > 0 {
        return chn.ChannelSpecs.GetMinNote()[chn.Num]
    }
    return 0
}

func (chn *Channel) GetMaxVolume() int {
    if len(chn.ChannelSpecs.GetMaxVol()) > 0 {
        return chn.ChannelSpecs.GetMaxVol()[chn.Num]
    }
    return 0
}

func (chn *Channel) MachineVolLimit() int {
    // ToDo: correct implementation?
    return chn.GetMaxVolume()
}

func (chn *Channel) SetMaxVolume(maxVol int) {
    if len(chn.ChannelSpecs.GetMaxVol()) > 0 {
        chn.ChannelSpecs.GetMaxVol()[chn.Num] = maxVol
    }
}

func (chn *Channel) AddCmd(cmds []int) {
    chn.Cmds = append(chn.Cmds, cmds...)
}


/* Calculate the length of a note in frames based on its length in 32nd notes, the current tempo,
 * the current playback speed and the current note cutoff setting.
 */
func (chn *Channel) NoteLength(len float64) (frames, cutoffFrames, scaling float64) {
    var length32 int

    frames = (timing.UpdateFreq * 60.0) / float64(chn.CurrentTempo) // frames per quarternote
    
    if timing.UseFractionalDelays {
        scaling = 256.0
        length32 = int((frames / 8.0) * scaling)    // frames per 32nd note, scaled by 256
        frames = math.Floor(float64(length32) * len)
        
        if (chn.CurrentCutoff.Typ == defs.CT_FRAMES ||
            chn.CurrentCutoff.Typ == defs.CT_NEG_FRAMES) {
            cutoffFrames = math.Min(float64(chn.CurrentCutoff.Val) * scaling, frames)
        } else {
            cutoffFrames = (frames * (8.0 - float64(chn.CurrentCutoff.Val))) / 8.0
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
            cutoffFrames = math.Min(float64(chn.CurrentCutoff.Val), frames)
        } else {
            cutoffFrames = math.Floor((frames * (8 - float64(chn.CurrentCutoff.Val))) / 8)
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


/* Split a length into either 2 or 3 bytes in the format that is recognized by the playback
 * libraries.
 */
func SplitLength(len int, scaling float64) []int {
    // The longest length that can be represented by a 2-byte sequence is 127.0
    if float64(len) > 127.0 * scaling {
        return []int{(len / 0x8000) | 0x80,     // 0x80 is a marker saying that this is a 3-byte sequence
                     (len / 0x100) & 0x7F,
                     (len & 0xFF)}
    } else {
        return []int{(len / 0x100),
                     (len & 0xFF)}
    }
}


/* Write the length of the current note to the channel's command
 * stream.
 */
func (chn *Channel) WriteLength() {
    
    if timing.UseFractionalDelays {
        len1 := int(math.Floor(chn.CurrentNoteFrames.Active))
        len2 := int(math.Floor(chn.CurrentNoteFrames.Cutoff))
        scaling := 256.0
        
        timing.UpdateDelayMinMax(len1)
        timing.UpdateDelayMinMax(len2)

        chn.AddCmd(SplitLength(int(math.Floor(chn.CurrentNoteFrames.Active)), scaling))      
    } else {
        // ToDo: necessary to handle this?
    }
}


func (chn *Channel) WriteNoteAndLength(note int, noteLen int, minLen int, scaling float64) {
    if noteLen >= minLen {
        chn.AddCmd([]int{note})
        chn.AddCmd(SplitLength(noteLen, scaling))
    }
}

                                         
/* Output all pending notes on this channel to the current song.
 */
func (chn *Channel) WriteNote(forceOctChange bool) {
    var frames, cutoffFrames, scaling float64
    var len1, len2 int
       
    if chn.CurrentNote.HasData {
        if !chn.Tuple.Active {
            chn.Frames += chn.CurrentNote.Frames
            chn.Ticks += int(chn.CurrentNote.Frames)
            
            frames, cutoffFrames, scaling = chn.NoteLength(chn.CurrentNote.Frames)
                                                    
            if chn.CurrentNote.Num == defs.Rest {
                chn.CurrentNote.Num = defs.CMD_REST
            } else if chn.CurrentNote.Num == defs.Rest2 {
                chn.CurrentNote.Num = defs.CMD_REST2
            } else {
                chn.CurrentNote.Num = chn.CurrentNote.Num % 12
            }

            // Handle any pending octave increase/decrease operations
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
        // There's no note data to output, but there's been an octave change that
        // we need to flush to the command array.
        if !chn.Tuple.Active {
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
    var w1, w2 int
   
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
    
    for i, _ := range chn.Tuple.Cmds {
        if chn.Tuple.Cmds[i].Num == defs.NON_NOTE_TUPLE_CMD {
            chn.AddCmd([]int{int(chn.Tuple.Cmds[i].Frames)})
        } else if chn.Tuple.Cmds[i].Num == defs.Rest {
            chn.Tuple.Cmds[i].Num = defs.CMD_REST
        } else if chn.Tuple.Cmds[i].Num == defs.Rest2 {
            chn.Tuple.Cmds[i].Num = defs.CMD_REST2
        } else {
            chn.Tuple.Cmds[i].Num = chn.Tuple.Cmds[i].Num % 12
        }

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
    

    if w2 >= 1 {
        timing.UpdateDelayMinMax(w2)
    }
    
    if chn.CurrentCutoff.Typ == defs.CT_NORMAL ||
       chn.CurrentCutoff.Typ == defs.CT_FRAMES {
        chn.WriteNoteAndLength(defs.CMD_REST, w2, 1, scaling)
    }
}


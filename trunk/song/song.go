/*
 * Package song
 *
 * Part of XPMC.
 *
 * /Mic, 2012-2014
 */
 
package song

import (
    "fmt"
    "../channel"
    "../defs"
    "../specs"
    "../targets"
)


type Song struct {
    Channels []*channel.Channel
    Target defs.ITarget
    
    TuneSmsPitch bool
    
    Num int
    Title string
    Composer string
    Programmer string
    Game string
    Album string
}


func NewSong(num int, targetId int, icomp defs.ICompiler) *Song {
    s := &Song{}
    s.Target = targets.NewTarget(targetId, icomp)
    s.Target.Init()
    s.Num = num
    chnSpecs := s.Target.GetChannelSpecs()
    for i, _ := range chnSpecs.GetDuty() {
        chn := channel.NewChannel()
        chn.Num = i
        chn.Name = fmt.Sprintf("%c", 'A'+i)
        chn.ChannelSpecs = chnSpecs
        s.Channels = append(s.Channels, chn)
    }

    // Create a "virtual" channel used for patterns    
    chn := channel.NewChannel()
    chn.Num = len(s.Channels)
    chn.Name = "Pattern"
    targetSpecs := s.Target.GetChannelSpecs().(*specs.Specs)
    comboSpecs := specs.CreateCombination(targetSpecs)
    specs.SetChannelSpecs(targetSpecs, 0, len(s.Channels), comboSpecs)
    chn.ChannelSpecs = chnSpecs
    chn.IsVirtualChannel = true
    s.Channels = append(s.Channels, chn)
    
    return s
}


/* ISong interface *
/*******************/

func (song *Song) GetChannels() []defs.IChannel {
    channels := make([]defs.IChannel, len(song.Channels))
    for i, chn := range song.Channels {
        channels[i] = chn
    }
    return channels
}

func (song *Song) GetComposer() string {
    return song.Composer
}

func (song *Song) GetNum() int {
    return song.Num
}

func (song *Song) GetProgrammer() string {
    return song.Programmer
}

func (song *Song) GetSmsTuning() bool {
    return song.TuneSmsPitch
}

func (song *Song) GetTitle() string {
    return song.Title
}

/* Returns true if the song is using any of the channels
 * of the given chip.
 */
func (song *Song) UsesChip(chipId int) bool {
    channels := song.GetChannels()
    for _, chn := range channels {
        if chn.IsUsed() && chn.GetChipID() == chipId {
            return true
        }
    }
    return false
}

/**********************/

func (song *Song) GetChannelType(chn int) int {
    return song.Channels[chn].GetChipID()
}

/* Get the number of currently active channels.
 */
func (song *Song) GetNumActiveChannels() (numActive int) {
    numActive = 0
    for _, chn := range song.Channels {
        if chn.Active {
            numActive++
        }
    }
    return
}


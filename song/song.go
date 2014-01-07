/*
 * Package song
 *
 * Part of XPMC.
 *
 * /Mic, 2012-2013
 */
 
package song

import (
    "fmt"
    "../channel"
    "../defs"
    "../targets"
)


type Song struct {
    Channels []*channel.Channel
    Target defs.ITarget
    
    TuneSmsPitch bool
    
    Title string
    Composer string
    Programmer string
    Game string
    Album string
}


func NewSong(targetId int, icomp defs.ICompiler) *Song {
    s := &Song{}
    s.Target = targets.NewTarget(targetId, icomp)
    s.Target.Init()
    chnSpecs := s.Target.GetChannelSpecs()
    for i, _ := range chnSpecs.GetDuty() {
        chn := channel.NewChannel()
        chn.Num = i
        chn.Name = fmt.Sprintf("%c", 'A'+i)
        chn.ChannelSpecs = chnSpecs
        s.Channels = append(s.Channels, chn)
    }
    return s
}

func (song *Song) GetChannels() []defs.IChannel {
    channels := make([]defs.IChannel, len(song.Channels))
    for i, chn := range song.Channels {
        channels[i] = chn
    }
    return channels
}

func (song *Song) GetNumActiveChannels() (numActive int) {
    numActive = 0
    for _, chn := range song.Channels {
        if chn.Active {
            numActive++
        }
    }
    return
}

func (song *Song) GetTitle() string {
    return song.Title
}

func (song *Song) GetComposer() string {
    return song.Composer
}

func (song *Song) GetProgrammer() string {
    return song.Programmer
}

func (song *Song) GetChannelType(chn int) int {
    return song.Channels[chn].GetChipID()
}

func (song *Song) GetSmsTuning() bool {
    return song.TuneSmsPitch
}

func (song *Song) UsesChip(chipId int) bool {
    channels := song.GetChannels()
    for _, chn := range channels {
        if chn.IsUsed() && chn.GetChipID() == chipId {
            return true
        }
    }
    return false
}

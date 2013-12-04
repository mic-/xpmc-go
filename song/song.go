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


func (song *Song) GetNumActiveChannels() (numActive int) {
    numActive = 0
    for _, chn := range song.Channels {
        if chn.Active {
            numActive++
        }
    }
    return
}

func (song *Song) GetChannelType(chn int) int {
    return song.Channels[chn].GetChipID()
}

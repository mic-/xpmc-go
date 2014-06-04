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


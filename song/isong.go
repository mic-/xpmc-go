package song

import "../defs"

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


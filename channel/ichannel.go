package channel

import "../specs"

/* IChannel interface *
/**********************/

func (chn *Channel) GetNum() int {
    return chn.Num
}

func (chn *Channel) GetName() string {
    return chn.Name
}

func (chn *Channel) GetCommands() []int {
    return chn.Cmds
}

func (chn *Channel) GetTicks() int {
    return chn.Ticks
}

func (chn *Channel) GetLoopTicks() int {
    return chn.LoopTicks
}

/* Returns the ID (one of the specs.CHIP_* constants) of the sound chip
 * that this channel is a part of.
 */
func (chn *Channel) GetChipID() int {
    if len(chn.ChannelSpecs.GetIDs()) > 0 {
        return chn.ChannelSpecs.GetIDs()[chn.Num]
    }
    return specs.CHIP_UNKNOWN
}

func (chn *Channel) IsVirtual() bool {
    return chn.IsVirtualChannel
}

func (chn *Channel) IsUsed() bool {
    return chn.HasAnyNote
}

func (chn *Channel) IsUsingEffect(effName string) bool {
    return chn.UsesEffect[effName]
}

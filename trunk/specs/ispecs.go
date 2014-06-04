package specs

/* ISpecs interface implementation *
/***********************************/

func (s *Specs) GetDuty() []int {
    return s.Duty
}

func (s *Specs) GetVolChange() []int {
    return s.VolChange
}

func (s *Specs) GetFM() []int {
    return s.FM
}

func (s *Specs) GetADSR() []int {
    return s.ADSR
}

func (s *Specs) GetFilter() []int {
    return s.Filter
}

func (s *Specs) GetRingMod() []int {
    return s.RingMod
}

func (s *Specs) GetWaveTable() []int {
    return s.WaveTable
}

func (s *Specs) GetPCM() []int {
    return s.PCM
}

func (s *Specs) GetToneEnv() []int {
    return s.ToneEnv
}

func (s *Specs) GetVolEnv() []int {
    return s.VolEnv
}

func (s *Specs) GetDetune() []int {
    return s.Detune
}

func (s *Specs) GetMinOct() []int {
    return s.MinOct
}

func (s *Specs) GetMaxOct() []int {
    return s.MaxOct
}

func (s *Specs) GetMaxVol() []int {
    return s.MaxVol
}

func (s *Specs) GetMinNote() []int {
    return s.MinNote
}

func (s *Specs) GetID() int {
    return s.ID
}

func (s *Specs) GetIDs() []int {
    return s.IDs
}

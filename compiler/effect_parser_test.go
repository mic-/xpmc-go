package compiler_test

import (
	"testing"

	"xpmc-go/defs"
	"xpmc-go/effects"
)

// intSlice converts []interface{} to []int for test assertions against ParamList data.
func intSlice(vals []interface{}) []int {
	out := make([]int, len(vals))
	for i, v := range vals {
		out[i] = v.(int)
	}
	return out
}

// --- Arpeggio (EN / handleArpeggioDef) ---

func TestArpeggioDefinitionStored(t *testing.T) {
	doCompile(t, "sms", "@EN0 = {0 4 7}")

	if effects.Arpeggios.FindKey(0) < 0 {
		t.Fatal("EN0 not registered in effects.Arpeggios")
	}
	data := effects.Arpeggios.GetData(0)
	assertEq(t, intSlice(data.MainPart), []int{0, 4, 7})
	assertEq(t, len(data.LoopedPart), 0)
}

func TestArpeggioLoopedStored(t *testing.T) {
	// The '|' separates the intro (MainPart) from the looped section (LoopedPart).
	doCompile(t, "sms", "@EN0 = {0 | 4 7}")

	data := effects.Arpeggios.GetData(0)
	if data == nil {
		t.Fatal("EN0 not registered")
	}
	assertEq(t, intSlice(data.MainPart), []int{0})
	assertEq(t, intSlice(data.LoopedPart), []int{4, 7})
}

func TestArpeggioDefaultFrequency(t *testing.T) {
	// Without an explicit frequency suffix the default is EFFECT_STEP_EVERY_FRAME.
	doCompile(t, "sms", "@EN0 = {0 4 7}")
	assertEq(t, effects.Arpeggios.GetExtraInt(0, effects.EXTRA_EFFECT_FREQ), defs.EFFECT_STEP_EVERY_FRAME)
}

func TestArpeggioEveryNoteFrequency(t *testing.T) {
	// The (EVERY-NOTE) suffix sets EFFECT_STEP_EVERY_NOTE via getEffectFrequency().
	doCompile(t, "sms", "@EN0 = {0 4 7}(EVERY-NOTE)")
	assertEq(t, effects.Arpeggios.GetExtraInt(0, effects.EXTRA_EFFECT_FREQ), defs.EFFECT_STEP_EVERY_NOTE)
}

func TestArpeggioApplication(t *testing.T) {
	// EN0 emits {CMD_ARPMAC, idx+1} where idx=0 (first entry), so arg = 1.
	mml := "@EN0 = {0 4 7}\nA t120 o4 EN0 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_ARPMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestArpeggioEveryNoteApplication(t *testing.T) {
	// (EVERY-NOTE) ORs bit 7 into the map index: arg becomes (0+1)|0x80 = 0x81.
	mml := "@EN0 = {0 4 7}(EVERY-NOTE)\nA t120 o4 EN0 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_ARPMAC, 1 | 0x80},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestArpeggioDisable(t *testing.T) {
	// ENOF (no arpeggio number) emits the single-byte CMD_ARPOFF.
	mml := "@EN0 = {0 4 7}\nA t120 o4 EN0 c ENOF c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_ARPMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_ARPOFF},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestArpeggioInlined(t *testing.T) {
	// EN({...}) defines and applies an arpeggio without a prior @EN declaration.
	mml := "A t120 o4 EN({0 4 7}) c"
	s := doCompile(t, "sms", mml)
	cmds := chanCmds(t, s, "A")
	// Verify CMD_ARPMAC is present; the exact inline ID is an implementation detail.
	if len(cmds) < 3 || cmds[1] != defs.CMD_ARPMAC {
		t.Errorf("expected CMD_ARPMAC in output, got %#04x", cmds)
	}
}

func TestArpeggioSecondDefinitionIndex(t *testing.T) {
	// EN1 is stored at map position 1, so its arg is idx+1 = 2.
	mml := "@EN0 = {0 4 7}\n@EN1 = {0 3 7}\nA t120 o4 EN1 c"
	s := doCompile(t, "sms", mml)
	if effects.Arpeggios.FindKey(1) < 0 {
		t.Fatal("EN1 not registered")
	}
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_ARPMAC, 2},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

// --- Vibrato (MP / handleVibratoDef) ---

func TestVibratoDefinitionStored(t *testing.T) {
	// @MP0 = {delay speed depth}; ranges: delay 0..127, speed 1..127, depth 0..63.
	doCompile(t, "sms", "@MP0 = {5 3 2}")

	if effects.Vibratos.FindKey(0) < 0 {
		t.Fatal("MP0 not registered")
	}
	data := effects.Vibratos.GetData(0)
	assertEq(t, intSlice(data.MainPart), []int{5, 3, 2})
	assertEq(t, len(data.LoopedPart), 0)
}

func TestVibratoApplication(t *testing.T) {
	mml := "@MP0 = {5 3 2}\nA t120 o4 MP0 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_VIBMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestVibratoDisable(t *testing.T) {
	// MPOF emits {CMD_VIBMAC, 0}.
	mml := "@MP0 = {5 3 2}\nA t120 o4 MP0 c MPOF c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_VIBMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_VIBMAC, 0},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

// --- Pitch envelope (EP / handleEffectDefinition via EP path) ---

func TestPitchEnvelopeDefinitionStored(t *testing.T) {
	doCompile(t, "sms", "@EP0 = {-4 -3 -2 -1 0 1}")

	if effects.PitchMacros.FindKey(0) < 0 {
		t.Fatal("EP0 not registered")
	}
	data := effects.PitchMacros.GetData(0)
	assertEq(t, intSlice(data.MainPart), []int{-4, -3, -2, -1, 0, 1})
}

func TestPitchEnvelopeLoopedStored(t *testing.T) {
	doCompile(t, "sms", "@EP0 = {-4 -3 | -2 -1 0}")

	data := effects.PitchMacros.GetData(0)
	if data == nil {
		t.Fatal("EP0 not registered")
	}
	assertEq(t, intSlice(data.MainPart), []int{-4, -3})
	assertEq(t, intSlice(data.LoopedPart), []int{-2, -1, 0})
}

func TestPitchEnvelopeApplication(t *testing.T) {
	mml := "@EP0 = {-4 -3 -2 -1 0}\nA t120 o4 EP0 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_SWPMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestPitchEnvelopeDisable(t *testing.T) {
	// EPOF emits {CMD_SWPMAC, 0}.
	mml := "@EP0 = {-4 -3 -2}\nA t120 o4 EP0 c EPOF c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_SWPMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_SWPMAC, 0},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

// --- Volume macro (@v / handleAtCommand "v" branch) ---

func TestVolumeMacroDefinitionStored(t *testing.T) {
	doCompile(t, "sms", "@v0 = {15 12 10 8}")

	if effects.VolumeMacros.FindKey(0) < 0 {
		t.Fatal("@v0 not registered")
	}
	assertEq(t, intSlice(effects.VolumeMacros.GetData(0).MainPart), []int{15, 12, 10, 8})
}

func TestVolumeMacroEveryFrameApplication(t *testing.T) {
	// Default (EVERY-FRAME): {CMD_VOLMAC, idx+1} = {CMD_VOLMAC, 1}.
	mml := "@v0 = {15 12 10 8}\nA t120 o4 @v0 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_VOLMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestVolumeMacroEveryNoteApplication(t *testing.T) {
	// (EVERY-NOTE): arg becomes (idx+1)|0x80 = 0x81.
	mml := "@v0 = {15 12 10}(EVERY-NOTE)\nA t120 o4 @v0 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_VOLMAC, 1 | 0x80},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

// --- Duty cycle macro (@<n> = {...} and @@<n> / handleDutyMacDef) ---

func TestDutyMacroDefinitionStored(t *testing.T) {
	// Without active channels, @0 = {...} stores in effects.DutyMacros.
	doCompile(t, "sms", "@0 = {0 1 0 1}")

	if effects.DutyMacros.FindKey(0) < 0 {
		t.Fatal("duty macro @0 not registered")
	}
	assertEq(t, intSlice(effects.DutyMacros.GetData(0).MainPart), []int{0, 1, 0, 1})
}

func TestDutyMacroApplication(t *testing.T) {
	// Channel D on SMS has SupportsDutyChange()==1 (SN76489 noise channel).
	// @@0 emits {CMD_DUTMAC, idx+1} = {CMD_DUTMAC, 1}.
	mml := "@0 = {0 1 0 1}\nD t120 o4 @@0 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, "D"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_DUTMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

// --- Panning macro (CS / handlePanMacDef) on Genesis ---

func TestPanMacroDefinitionStored(t *testing.T) {
	// Genesis (gen) sets SupportsPanning=1; @CS0 values must be in -63..63.
	doCompile(t, "gen", "@CS0 = {-32 0 32}")

	if effects.PanMacros.FindKey(0) < 0 {
		t.Fatal("CS0 not registered")
	}
	assertEq(t, intSlice(effects.PanMacros.GetData(0).MainPart), []int{-32, 0, 32})
}

func TestPanMacroApplication(t *testing.T) {
	mml := "@CS0 = {-32 0 32}\nA t120 o4 CS0 c"
	s := doCompile(t, "gen", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_PANMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestPanMacroDisable(t *testing.T) {
	// CSOF emits {CMD_PANMAC, 0}.
	mml := "@CS0 = {-32 0 32}\nA t120 o4 CS0 c CSOF c"
	s := doCompile(t, "gen", mml)
	check(t, chanCmds(t, s, "A"), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_PANMAC, 1},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_PANMAC, 0},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_END},
	))
}

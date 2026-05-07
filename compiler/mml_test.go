package compiler_test

import (
	"slices"
	"testing"

	"xpmc-go/compiler"
	"xpmc-go/defs"
	"xpmc-go/song"
	"xpmc-go/targets"
	"xpmc-go/timing"
	"xpmc-go/utils"
)

// Note IDs (semitone within octave, matching CMD_NOTE encoding).
const (
	N_C  = 0
	N_Cs = 1
	N_D  = 2
	N_Ds = 3
	N_E  = 4
	N_F  = 5
	N_Fs = 6
	N_G  = 7
	N_Gs = 8
	N_A  = 9
	N_As = 10
	N_B  = 11
)

// Fractional frame lengths at t120 BPM, 60 Hz (floor(frames) * 256).
var (
	LEN_16 = [2]int{0x07, 0x80} // 16th:  floor(7.5 * 256)
	LEN_E  = [2]int{0x0F, 0x00} // 8th:   15 * 256
	LEN_Q  = [2]int{0x1E, 0x00} // 4th:   30 * 256
	LEN_H  = [2]int{0x3C, 0x00} // half:  60 * 256
	LEN_W  = [2]int{0x78, 0x00} // whole: 120 * 256
)

func noteSeq(id int, l [2]int) []int  { return []int{id, l[0], l[1]} }
func restSeq(l [2]int) []int          { return []int{defs.CMD_REST, l[0], l[1]} }
func lenCmd(l [2]int) []int           { return []int{defs.CMD_LEN, l[0], l[1]} }
func cat(seqs ...[]int) []int {
	var out []int
	for _, s := range seqs {
		out = append(out, s...)
	}
	return out
}

// doCompile sets up the global timing state, compiles the MML snippet under
// the named target, appends CMD_END / CMD_JMP markers, and returns song 1.
// The MML string should start with the channel name (e.g. "A t120 o4 c").
func doCompile(t *testing.T, targetName, mml string) *song.Song {
	t.Helper()

	timing.UpdateFreq = 60.0
	timing.UseFractionalDelays = true
	timing.SupportedLengths = defs.EXTENDED_LENGTHS()
	utils.OldParsers = utils.NewGenericStack()
	utils.Verbose(false)
	utils.DebugMode(false)
	utils.WarningsAreErrors(false)

	targetID := targets.NameToID(targetName)
	if targetID == targets.TARGET_UNKNOWN {
		t.Fatalf("unknown target %q", targetName)
	}

	comp := &compiler.Compiler{}
	comp.Init(targetID)

	comp.CompileCode("#TITLE Unit test\n#COMPOSER Test\n\n" + mml + "\n")

	for _, s := range comp.Songs {
		for _, chn := range s.Channels {
			if chn.IsVirtual() {
				continue
			}
			chn.LoopTicks = chn.Ticks - chn.LoopTicks
			if chn.LoopPoint == -1 {
				chn.AddCmd([]int{defs.CMD_END})
			} else if !chn.HasAnyNote {
				chn.AddCmd([]int{defs.CMD_END})
			} else {
				chn.AddCmd([]int{defs.CMD_JMP, chn.LoopPoint & 0xFF, chn.LoopPoint / 0x100})
			}
		}
	}

	return comp.Songs[1]
}

// chanCmds returns the Cmds slice for the named channel in song s.
func chanCmds(t *testing.T, s *song.Song, name string) []int {
	t.Helper()
	for _, chn := range s.Channels {
		if chn.Name == name && !chn.IsVirtual() {
			return chn.Cmds
		}
	}
	t.Fatalf("channel %q not found", name)
	return nil
}

// check is a shorthand for comparing got vs want and failing with a diff.
func check(t *testing.T, got, want []int) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("\ngot:  %#04x\nwant: %#04x", got, want)
	}
}

// mmlChan extracts the leading channel letter from an MML snippet ("A t120 ..." → "A").
func mmlChan(mml string) string { return string(mml[0]) }

func TestBasicNotes(t *testing.T) {
	// All seven diatonic notes plus rest at t120 v10 o4.
	mml := "A t120 v10 o4 c d e f g a b r"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, mmlChan(mml)), cat(
		[]int{defs.CMD_VOL2 | 10, defs.CMD_OCTAVE | 4},
		noteSeq(N_C, LEN_Q), noteSeq(N_D, LEN_Q), noteSeq(N_E, LEN_Q),
		noteSeq(N_F, LEN_Q), noteSeq(N_G, LEN_Q), noteSeq(N_A, LEN_Q),
		noteSeq(N_B, LEN_Q), restSeq(LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestSharpNotes(t *testing.T) {
	mml := "A t120 o4 c+ d+ f+ g+ a+"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, mmlChan(mml)), cat(
		[]int{defs.CMD_OCTAVE | 4},
		noteSeq(N_Cs, LEN_Q), noteSeq(N_Ds, LEN_Q), noteSeq(N_Fs, LEN_Q),
		noteSeq(N_Gs, LEN_Q), noteSeq(N_As, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestFlatNotes(t *testing.T) {
	// Flats are enharmonic equivalents of sharps.
	mml := "A t120 o4 d- e- g- a- b-"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, mmlChan(mml)), cat(
		[]int{defs.CMD_OCTAVE | 4},
		noteSeq(N_Cs, LEN_Q), noteSeq(N_Ds, LEN_Q), noteSeq(N_Fs, LEN_Q),
		noteSeq(N_Gs, LEN_Q), noteSeq(N_As, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestExplicitNoteLengths(t *testing.T) {
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{"c4  = 30 frames ($1E00)", "A t120 o4 c4", cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END})},
		{"c8  = 15 frames ($0F00)", "A t120 o4 c8", cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_E), []int{defs.CMD_END})},
		{"c16 = 7.5 frames ($0780)", "A t120 o4 c16", cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_16), []int{defs.CMD_END})},
		{"c2  = 60 frames ($3C00)", "A t120 o4 c2", cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_H), []int{defs.CMD_END})},
		{"c1  = 120 frames ($7800)", "A t120 o4 c1", cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_W), []int{defs.CMD_END})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestDefaultLength(t *testing.T) {
	// l command emits CMD_LEN and switches subsequent notes to CMD_NOTE2 short form.
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"l4 c: CMD_LEN then CMD_NOTE2",
			"A t120 o4 l4 c",
			cat([]int{defs.CMD_OCTAVE | 4}, lenCmd(LEN_Q), []int{defs.CMD_NOTE2 | N_C, defs.CMD_END}),
		},
		{
			"l8 c: CMD_LEN then CMD_NOTE2",
			"A t120 o4 l8 c",
			cat([]int{defs.CMD_OCTAVE | 4}, lenCmd(LEN_E), []int{defs.CMD_NOTE2 | N_C, defs.CMD_END}),
		},
		{
			"l4 c d e: CMD_LEN then three CMD_NOTE2",
			"A t120 o4 l4 c d e",
			cat(
				[]int{defs.CMD_OCTAVE | 4},
				lenCmd(LEN_Q),
				[]int{defs.CMD_NOTE2 | N_C, defs.CMD_NOTE2 | N_D, defs.CMD_NOTE2 | N_E, defs.CMD_END},
			),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestDottedNotes(t *testing.T) {
	// c4.  = 1.5  * 30 = 45 frames   → 45 * 256 = $2D00
	// c4.. = 1.75 * 30 = 52.5 frames → floor(52.5 * 256) = $3480
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"c4.  = 1.5 quarter ($2D00)",
			"A t120 o4 c4.",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, [2]int{0x2D, 0x00}), []int{defs.CMD_END}),
		},
		{
			"c4.. = 1.75 quarter ($3480)",
			"A t120 o4 c4..",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, [2]int{0x34, 0x80}), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestTempo(t *testing.T) {
	// t60: 60 f/quarter → $3C00 (same as LEN_H); t240: 15 f/quarter → $0F00 (LEN_E).
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"t60  quarter = 60 frames ($3C00)",
			"A t60  o4 c",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_H), []int{defs.CMD_END}),
		},
		{
			"t240 quarter = 15 frames ($0F00)",
			"A t240 o4 c",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_E), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestVolume(t *testing.T) {
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"v0  → $30",
			"A t120 v0  o4 c",
			cat([]int{defs.CMD_VOL2 | 0, defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"v8  → $38",
			"A t120 v8  o4 c",
			cat([]int{defs.CMD_VOL2 | 8, defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"v15 → $3F",
			"A t120 v15 o4 c",
			cat([]int{defs.CMD_VOL2 | 15, defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestVolumeUpDown(t *testing.T) {
	// v+ and v- emit {CMD_VOLUP, step}; v- negates the step as a signed byte.
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"v+ → {CMD_VOLUP, 1}",
			"A t120 o4 v10 v+ c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_VOL2 | 10, defs.CMD_VOLUP, 1}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"v- → {CMD_VOLUP, $FF}",
			"A t120 o4 v10 v- c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_VOL2 | 10, defs.CMD_VOLUP, 0xFF}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"v+3 → {CMD_VOLUP, 3}",
			"A t120 o4 v0 v+3 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_VOL2 | 0, defs.CMD_VOLUP, 3}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"v-2 → {CMD_VOLUP, $FE}",
			"A t120 o4 v15 v-2 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_VOL2 | 15, defs.CMD_VOLUP, 0xFE}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestNoiseType(t *testing.T) {
	// @ on channel D selects SN76489 noise type; emits CMD_DUTY | value.
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"@0: white noise (CMD_DUTY|0 = $20)",
			"D t120 o4 @0 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_DUTY | 0}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"@1: periodic noise (CMD_DUTY|1 = $21)",
			"D t120 o4 @1 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_DUTY | 1}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestOctaveSet(t *testing.T) {
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"o3 c → CMD_OCTAVE|3 = $13",
			"A t120 o3 c",
			cat([]int{defs.CMD_OCTAVE | 3}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"o5 c → CMD_OCTAVE|5 = $15",
			"A t120 o5 c",
			cat([]int{defs.CMD_OCTAVE | 5}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestOctaveUpDown(t *testing.T) {
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"o3 > c: up from 3 to 4",
			"A t120 o3 > c",
			cat([]int{defs.CMD_OCTAVE | 3, defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"o5 < c: down from 5 to 4",
			"A t120 o5 < c",
			cat([]int{defs.CMD_OCTAVE | 5, defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestRest(t *testing.T) {
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"r4 → CMD_REST + $1E00",
			"A t120 o4 r4",
			cat([]int{defs.CMD_OCTAVE | 4}, restSeq(LEN_Q), []int{defs.CMD_END}),
		},
		{
			"r8 → CMD_REST + $0F00",
			"A t120 o4 r8",
			cat([]int{defs.CMD_OCTAVE | 4}, restSeq(LEN_E), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestCutoff(t *testing.T) {
	// q4 = 4/8 = 50% active: each half uses LEN_E.
	// Emits: CMD_LEN+LEN_E, CMD_NOTE2 for C, CMD_REST+LEN_E.
	mml := "B t120 o4 q4 c"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, mmlChan(mml)), cat(
		[]int{defs.CMD_OCTAVE | 4},
		lenCmd(LEN_E),
		[]int{defs.CMD_NOTE2 | N_C},
		restSeq(LEN_E),
		[]int{defs.CMD_END},
	))
}

func TestTie(t *testing.T) {
	// c4^4 = quarter + quarter = half note = $3C00.
	mml := "A t120 o4 c4^4"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, mmlChan(mml)), cat(
		[]int{defs.CMD_OCTAVE | 4},
		noteSeq(N_C, LEN_H),
		[]int{defs.CMD_END},
	))
}

func TestRepeatLoop(t *testing.T) {
	// [c]3: CMD_LOPCNT + count, note, CMD_DJNZ + offset bytes.
	mml := "A t120 o4 [c]3"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, mmlChan(mml)), cat(
		[]int{defs.CMD_OCTAVE | 4, defs.CMD_LOPCNT, 3},
		noteSeq(N_C, LEN_Q),
		[]int{defs.CMD_DJNZ, 3, 0, defs.CMD_END},
	))
}

func TestInfiniteLoop(t *testing.T) {
	// L records the loop point; end marker becomes CMD_JMP to that position.
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			// OCT4 (1 byte) emitted before L → loop point = 1.
			"L before first note: CMD_JMP target = 1",
			"A t120 o4 L c",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), []int{defs.CMD_JMP, 1, 0}),
		},
		{
			// OCT4 (1 byte) + noteSeq(N_C, LEN_Q) (3 bytes) = 4 bytes before L → loop point = 4.
			"L after intro note: CMD_JMP skips the intro",
			"A t120 o4 c L d",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_Q), noteSeq(N_D, LEN_Q), []int{defs.CMD_JMP, 4, 0}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestExplicitLengthsAlwaysLongForm(t *testing.T) {
	// Per-note explicit lengths always use the 3-byte long form; CMD_NOTE2
	// is only emitted when the l command sets a cached length.
	mml := "A t120 o4 c4 d4 e4"
	s := doCompile(t, "sms", mml)
	check(t, chanCmds(t, s, mmlChan(mml)), cat(
		[]int{defs.CMD_OCTAVE | 4},
		noteSeq(N_C, LEN_Q), noteSeq(N_D, LEN_Q), noteSeq(N_E, LEN_Q),
		[]int{defs.CMD_END},
	))
}

func TestSlur(t *testing.T) {
	// & fuses same-pitch notes by summing lengths at compile time.
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"c4&c4 = half note (two quarters fused)",
			"A t120 o4 c4&c4",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, LEN_H), []int{defs.CMD_END}),
		},
		{
			// c4 (30 f) + c8 (15 f) = 45 f → 45 * 256 = $2D00
			"c4&c8 = dotted quarter ($2D00)",
			"A t120 o4 c4&c8",
			cat([]int{defs.CMD_OCTAVE | 4}, noteSeq(N_C, [2]int{0x2D, 0x00}), []int{defs.CMD_END}),
		},
		{
			// After l4, the fused length (60 f) no longer matches the cache → long form.
			"l4 c&c: slur breaks CMD_NOTE2, emits long-form half note",
			"A t120 o4 l4 c&c",
			cat([]int{defs.CMD_OCTAVE | 4}, lenCmd(LEN_Q), noteSeq(N_C, LEN_H), []int{defs.CMD_END}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestDetune(t *testing.T) {
	// D emits {CMD_DETUNE, value} with the raw signed integer.
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"D5 c: positive detune",
			"A t120 o4 D5 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_DETUNE, 5}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"D-5 c: negative detune",
			"A t120 o4 D-5 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_DETUNE, -5}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"D5 c D0 d: detune reset between notes",
			"A t120 o4 D5 c D0 d",
			cat(
				[]int{defs.CMD_OCTAVE | 4, defs.CMD_DETUNE, 5},
				noteSeq(N_C, LEN_Q),
				[]int{defs.CMD_DETUNE, 0},
				noteSeq(N_D, LEN_Q),
				[]int{defs.CMD_END},
			),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}

func TestTranspose(t *testing.T) {
	// K emits {CMD_TRANSP, value} with the raw signed integer.
	cases := []struct {
		name string
		mml  string
		want []int
	}{
		{
			"K5 c: positive transpose",
			"A t120 o4 K5 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_TRANSP, 5}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"K-3 c: negative transpose",
			"A t120 o4 K-3 c",
			cat([]int{defs.CMD_OCTAVE | 4, defs.CMD_TRANSP, -3}, noteSeq(N_C, LEN_Q), []int{defs.CMD_END}),
		},
		{
			"K5 c K0 d: transpose reset between notes",
			"A t120 o4 K5 c K0 d",
			cat(
				[]int{defs.CMD_OCTAVE | 4, defs.CMD_TRANSP, 5},
				noteSeq(N_C, LEN_Q),
				[]int{defs.CMD_TRANSP, 0},
				noteSeq(N_D, LEN_Q),
				[]int{defs.CMD_END},
			),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := doCompile(t, "sms", tc.mml)
			check(t, chanCmds(t, s, mmlChan(tc.mml)), tc.want)
		})
	}
}


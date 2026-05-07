/*
 * Package vgm
 *
 * Minimal VGM writer for xpmc-go.
 *
 * /Mic, 2026
 */

package vgm

import (
    "bytes"
    "compress/gzip"
    "encoding/binary"
    "fmt"
    "io"
    "os"
    "xpmc-go/defs"
    "xpmc-go/specs"
)

const (
    VGM_HEADER_SIZE = 0x40

    VGM_MAGIC_OFFSET            = 0x00
    VGM_EOF_OFFSET              = 0x04
    VGM_VERSION_OFFSET          = 0x08
    VGM_SN76489_CLOCK_OFFSET    = 0x0C
    VGM_YM2413_CLOCK_OFFSET     = 0x10
    VGM_GD3_OFFSET              = 0x14
    VGM_TOTAL_SAMPLES_OFFSET    = 0x18
    VGM_LOOP_OFFSET             = 0x1C
    VGM_LOOP_SAMPLES_OFFSET     = 0x20
    VGM_RATE_OFFSET             = 0x24
    VGM_SN76489_FEEDBACK_OFFSET = 0x28
    VGM_SN76489_SHIFT_OFFSET    = 0x2A
    VGM_YM2612_CLOCK_OFFSET     = 0x2C
    VGM_YM2151_CLOCK_OFFSET     = 0x30
    VGM_DATA_OFFSET_OFFSET      = 0x34

    VGM_VERSION_150             = 0x00000150
    VGM_SN76489_CLOCK           = 3579545
    VGM_YM2413_CLOCK            = 3579545
    VGM_YM2612_CLOCK            = 7670454
    VGM_YM2151_CLOCK            = 3579545

    VGM_END_COMMAND_0           = 0x66
    VGM_END_COMMAND_1           = 0x00
)

func WriteVGM(filename string, song defs.ISong, compress bool, rate int) error {
    if song == nil {
        return fmt.Errorf("song is nil")
    }

    // TODO: replace this placeholder header-only writer with a generic
    // player-based playback pipeline that steps through a Song and emits
    // VGM commands via a dedicated VGM playback writer.

    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := io.Writer(file)
    var gzipWriter *gzip.Writer
    if compress {
        gzipWriter = gzip.NewWriter(file)
        writer = gzipWriter
    }

    header := make([]byte, VGM_HEADER_SIZE)
    copy(header[VGM_MAGIC_OFFSET:VGM_MAGIC_OFFSET+4], []byte("Vgm "))
    binary.LittleEndian.PutUint32(header[VGM_VERSION_OFFSET:VGM_VERSION_OFFSET+4], VGM_VERSION_150)

    if song.UsesChip(specs.CHIP_SN76489) || song.UsesChip(specs.CHIP_T6W28) {
        binary.LittleEndian.PutUint32(header[VGM_SN76489_CLOCK_OFFSET:VGM_SN76489_CLOCK_OFFSET+4], VGM_SN76489_CLOCK)
        binary.LittleEndian.PutUint16(header[VGM_SN76489_FEEDBACK_OFFSET:VGM_SN76489_FEEDBACK_OFFSET+2], 0x0009)
        header[VGM_SN76489_SHIFT_OFFSET] = 16
    }
    if song.UsesChip(specs.CHIP_YM2413) {
        binary.LittleEndian.PutUint32(header[VGM_YM2413_CLOCK_OFFSET:VGM_YM2413_CLOCK_OFFSET+4], VGM_YM2413_CLOCK)
    }
    if song.UsesChip(specs.CHIP_YM2612) {
        binary.LittleEndian.PutUint32(header[VGM_YM2612_CLOCK_OFFSET:VGM_YM2612_CLOCK_OFFSET+4], VGM_YM2612_CLOCK)
    }
    if song.UsesChip(specs.CHIP_YM2151) {
        binary.LittleEndian.PutUint32(header[VGM_YM2151_CLOCK_OFFSET:VGM_YM2151_CLOCK_OFFSET+4], VGM_YM2151_CLOCK)
    }
    if rate > 0 {
        binary.LittleEndian.PutUint32(header[VGM_RATE_OFFSET:VGM_RATE_OFFSET+4], uint32(rate))
    }

    binary.LittleEndian.PutUint32(header[VGM_GD3_OFFSET:VGM_GD3_OFFSET+4], 0)
    binary.LittleEndian.PutUint32(header[VGM_TOTAL_SAMPLES_OFFSET:VGM_TOTAL_SAMPLES_OFFSET+4], 0)
    binary.LittleEndian.PutUint32(header[VGM_LOOP_OFFSET:VGM_LOOP_OFFSET+4], 0)
    binary.LittleEndian.PutUint32(header[VGM_LOOP_SAMPLES_OFFSET:VGM_LOOP_SAMPLES_OFFSET+4], 0)
    binary.LittleEndian.PutUint32(header[VGM_DATA_OFFSET_OFFSET:VGM_DATA_OFFSET_OFFSET+4], VGM_HEADER_SIZE)

    data := bytes.NewBuffer(header)
    data.Write([]byte{VGM_END_COMMAND_0, VGM_END_COMMAND_1})
    buf := data.Bytes()
    binary.LittleEndian.PutUint32(buf[VGM_EOF_OFFSET:VGM_EOF_OFFSET+4], uint32(len(buf)-4))

    if _, err := writer.Write(buf); err != nil {
        if gzipWriter != nil {
            _ = gzipWriter.Close()
        }
        return err
    }

    if gzipWriter != nil {
        return gzipWriter.Close()
    }
    return nil
}

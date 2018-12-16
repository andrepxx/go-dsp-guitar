package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

/*
 * Global constants.
 */
const (
	BITS_PER_BYTE                 = 8
	MIN_CHUNK_HEADER_SIZE         = 8
	MIN_DATASIZE_CHUNK_SIZE       = 36
	LENGTH_DATASIZE_TABLE_ENTRIES = 12
)

/*
 * Constants for handling of 24-bit integers.
 */
const (
	MAX_INT24      = 0x007fffff       // int32
	MIN_INT24      = -(MAX_INT24 + 1) // int32
	SIGN_BIT_INT24 = 0x00800000       // int32
	SIZE_INT24     = 3
)

/*
 * RIFF header constants.
 */
const (
	AUDIO_PCM             = 0x0001     // uint16
	AUDIO_IEEE_FLOAT      = 0x0003     // uint16
	DEFAULT_BIT_DEPTH     = 0x0010     // uint16
	FORMAT_WAVE           = 0x45564157 // uint32
	ID_DATA               = 0x61746164 // uint32
	ID_DATASIZE           = 0x34367364 // uint32
	ID_FORMAT             = 0x20746d66 // uint32
	ID_RIFF               = 0x46464952 // uint32
	ID_RIFF64             = 0x34364652 // uint32
	MIN_CHUNK_SIZE_FORMAT = 0x00000010 // uint32
	MIN_TOTAL_HEADER_SIZE = 0x0000002c // uint32
)

/*
 * An interface type representing the channels inside a RIFF wave file.
 */
type Channel interface {
	Clear()
	Floats() []float64
	WriteFloats(samples []float64)
}

/*
 * The internal data structure representing a channel of a RIFF wave file.
 */
type channelStruct struct {
	samples []float64
}

/*
 * An interface type representing a RIFF wave file.
 */
type File interface {
	BitDepth() uint16
	Bytes() ([]byte, error)
	Channel(id uint16) (Channel, error)
	ChannelCount() uint16
	SampleFormat() uint16
	SampleRate() uint32
}

/*
 * The internal data structure representing a RIFF wave file.
 */
type fileStruct struct {
	bitDepth     uint16
	sampleFormat uint16
	sampleRate   uint32
	channels     []Channel
}

/*
 * The structure of a wave file's RIFF header.
 */
type riffHeader struct {
	ChunkID   uint32
	ChunkSize uint32
	Format    uint32
}

/*
 * The structure of a wave file's data size header.
 */
type dataSizeHeader struct {
	ChunkID     uint32
	ChunkSize   uint32
	SizeRIFF    uint64
	SizeData    uint64
	SampleCount uint64
	TableLength uint32
}

/*
 * The structure of a chunk header for pre-parsing.
 */
type chunkHeader struct {
	ChunkID   uint32
	ChunkSize uint32
}

/*
 * The structure of a wave file's format header.
 */
type formatHeader struct {
	ChunkID      uint32
	ChunkSize    uint32
	AudioFormat  uint16
	ChannelCount uint16
	SampleRate   uint32
	ByteRate     uint32
	BlockAlign   uint16
	BitDepth     uint16
}

/*
 * The structure of a wave file's data header.
 */
type dataHeader struct {
	ChunkID   uint32
	ChunkSize uint32
}

/*
 * Clears all samples from the channel.
 */
func (this *channelStruct) Clear() {
	this.samples = make([]float64, 0)
}

/*
 * Returns all samples inside this channel in floating-point representation.
 */
func (this *channelStruct) Floats() []float64 {
	size := len(this.samples)
	samples := make([]float64, size)
	copy(samples, this.samples)
	return samples
}

/*
 * Writes (appends) samples in floating-point representation to this channel.
 */
func (this *channelStruct) WriteFloats(samples []float64) {
	this.samples = append(this.samples, samples...)
}

/*
 * Utility function for creating an empty buffer.
 */
func createBuffer() *bytes.Buffer {
	buf := bytes.Buffer{}
	return &buf
}

/*
 * Converts a slice of channels into a slice of samples.
 */
func channelsToSamples(channels []Channel) []float64 {
	channelCount := len(channels)
	channelCount16 := uint16(channelCount)
	channelCount32 := uint32(channelCount)
	samplesByChannel := make([][]float64, channelCount)
	maxSampleCount := uint32(0)

	/*
	 * Iterate over all channels and extract the samples for each.
	 */
	for i, currentChannel := range channels {
		currentSamples := currentChannel.Floats()
		sampleCount := len(currentSamples)
		sampleCount32 := uint32(sampleCount)
		samplesByChannel[i] = currentSamples

		/*
		 * If we found a channel with more samples, make its sample
		 * count the new longest channel sample count.
		 */
		if sampleCount32 > maxSampleCount {
			maxSampleCount = sampleCount32
		}

	}

	totalSampleCount := channelCount32 * maxSampleCount
	data := make([]float64, totalSampleCount)

	/*
	 * Iterate over the samples to reorder them by time.
	 */
	for i := uint32(0); i < maxSampleCount; i++ {

		/*
		 * Iterate over the channels and extract the current sample.
		 */
		for j := uint16(0); j < channelCount16; j++ {
			currentChannel := samplesByChannel[j]
			currentChannelLength := len(currentChannel)
			currentChannelLength32 := uint32(currentChannelLength)
			currentSample := float64(0.0)

			/*
			 * If the channel is long enough, read the sample from it,
			 * otherwise pad with zeroes.
			 */
			if i < currentChannelLength32 {
				currentSample = currentChannel[i]
			}

			j32 := uint32(j)
			offset := (channelCount32 * i) + j32
			data[offset] = currentSample
		}

	}

	return data
}

/*
 * Converts a slice of samples into a slice of channels.
 */
func samplesToChannels(samples []float64, channelCount uint16) []Channel {
	channels := make([]Channel, channelCount)
	channelCount32 := uint32(channelCount)
	size := len(samples)
	size32 := uint32(size)
	samplesPerChannel := size32 / channelCount32

	/*
	 * Extract each channel from the sample data.
	 */
	for i := uint16(0); i < channelCount; i++ {
		currentSamples := make([]float64, samplesPerChannel)
		i32 := uint32(i)

		/*
		 * Extract each sample for this channel.
		 */
		for j := uint32(0); j < samplesPerChannel; j++ {
			idx := (j * channelCount32) + i32
			currentSamples[j] = samples[idx]
		}

		/*
		 * Data structure representing this channel.
		 */
		channel := channelStruct{
			samples: currentSamples,
		}

		channels[i] = &channel
	}

	return channels
}

/*
 * Convert samples to bytes, encoding them as 8-bit LPCM values.
 */
func samplesToBytesLPCM8(samples []float64) ([]byte, error) {
	numSamples := len(samples)
	data := make([]byte, numSamples)
	scale := float64(math.MaxInt8)

	/*
	 * Iterate over the samples and encode them as 8-bit LPCM values.
	 */
	for i, sample := range samples {

		/*
		 * Make sure that limits are not exceeded.
		 */
		if sample < -1.0 {
			sample = -1.0
		} else if sample > 1.0 {
			sample = 1.0
		}

		temp := int16(scale * sample)
		res := temp - math.MinInt8

		/*
		 * Make sure that limits are not exceeded.
		 */
		if res < 0 {
			data[i] = 0
		} else if res > math.MaxUint8 {
			data[i] = math.MaxUint8
		} else {
			data[i] = byte(res)
		}

	}

	return data, nil
}

/*
 * Convert bytes, encoded as 8-bit LPCM values, to samples.
 */
func bytesToSamplesLPCM8(data []byte) ([]float64, error) {
	numSamples := len(data)
	samples := make([]float64, numSamples)
	scale := 1.0 / float64(math.MaxInt8)

	/*
	 * Iterate over the samples and decode the 8-bit LPCM values.
	 */
	for i, byt := range data {
		temp := int16(byt) + math.MinInt8
		res := scale * float64(temp)

		/*
		 * Make sure that limits are not exceeded.
		 */
		if res < -1.0 {
			samples[i] = -1.0
		} else if res > 1.0 {
			samples[i] = 1.0
		} else {
			samples[i] = res
		}

	}

	return samples, nil
}

/*
 * Convert samples to bytes, encoding them as 16-bit LPCM values.
 */
func samplesToBytesLPCM16(samples []float64) ([]byte, error) {
	numSamples := len(samples)
	samplesInt := make([]int16, numSamples)
	const delta = math.MaxInt16 - math.MinInt16
	scale := 0.5 * float64(delta)

	/*
	 * Iterate over the samples and convert them into integer representation.
	 */
	for i, sample := range samples {

		/*
		 * Make sure that limits are not exceeded.
		 */
		if sample < -1.0 {
			sample = -1.0
		} else if sample > 1.0 {
			sample = 1.0
		}

		tmp := int32(scale * sample)

		/*
		 * Make sure that limits are not exceeded.
		 */
		if tmp > math.MaxInt16 {
			tmp = math.MaxInt16
		} else if tmp < math.MinInt16 {
			tmp = math.MinInt16
		}

		samplesInt[i] = int16(tmp)
	}

	buf := createBuffer()
	err := binary.Write(buf, binary.LittleEndian, samplesInt)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to convert samples: %s", msg)
	} else {
		data := buf.Bytes()
		return data, nil
	}

}

/*
 * Convert bytes, encoded as 16-bit LPCM values, to samples.
 */
func bytesToSamplesLPCM16(data []byte) ([]float64, error) {
	numBytes := len(data)
	numBytes64 := uint64(numBytes)
	numSamples := numBytes64 >> 1
	samplesInt := make([]int16, numSamples)
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, samplesInt)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to decode LPCM16 data: %s", msg)
	} else {
		samplesFloat := make([]float64, numSamples)
		scaling := 2.0 / (math.MaxInt16 - math.MinInt16)

		/*
		 * Convert samples to floating-point representation.
		 */
		for i, sample := range samplesInt {
			samplesFloat[i] = scaling * float64(sample)
		}

		return samplesFloat, nil
	}

}

/*
 * Convert samples to bytes, encoding them as 24-bit LPCM values.
 */
func samplesToBytesLPCM24(samples []float64) ([]byte, error) {
	const delta = MAX_INT24 - MIN_INT24
	scale := 0.5 * float64(delta)
	buf := createBuffer()

	/*
	 * Iterate over the samples and convert them into integer representation.
	 */
	for _, sample := range samples {

		/*
		 * Make sure that limits are not exceeded.
		 */
		if sample < -1.0 {
			sample = -1.0
		} else if sample > 1.0 {
			sample = 1.0
		}

		tmp := int32(scale * sample)

		/*
		 * Make sure that limits are not exceeded.
		 */
		if tmp > MAX_INT24 {
			tmp = MAX_INT24
		} else if tmp < MIN_INT24 {
			tmp = MIN_INT24
		}

		sampleUint := uint32(tmp)

		/*
		 * Write each byte to the buffer.
		 */
		for j := 0; j < SIZE_INT24; j++ {
			shift := BITS_PER_BYTE * uint32(j)
			byt := byte((sampleUint >> shift) & 0xff)
			buf.WriteByte(byt)
		}

	}

	data := buf.Bytes()
	return data, nil
}

/*
 * Convert bytes, encoded as 24-bit LPCM values, to samples.
 */
func bytesToSamplesLPCM24(data []byte) ([]float64, error) {
	numBytes := len(data)
	numBytes64 := uint64(numBytes)
	numSamples := numBytes64 / SIZE_INT24
	samplesFloat := make([]float64, numSamples)
	scaling := 2.0 / (MAX_INT24 - MIN_INT24)
	reader := bytes.NewReader(data)
	buf := make([]byte, SIZE_INT24)
	words := make([]uint32, SIZE_INT24)

	/*
	 * Read samples from input stream.
	 */
	for idx := range samplesFloat {
		reader.Read(buf)

		/*
		 * Turn the single bytes from the buffer into machine words.
		 */
		for i, byt := range buf {
			words[i] = uint32(byt)
		}

		sampleWord := uint32(0)

		/*
		 * Combine the extracted words into a single machine word.
		 */
		for i, word := range words {
			shift := BITS_PER_BYTE * uint32(i)
			sampleWord |= word << shift
		}

		sampleInt := int32(sampleWord)
		signBit := (sampleWord & SIGN_BIT_INT24) != 0

		/*
		 * Handle negative values in two's complement representation.
		 */
		if signBit {
			offset := sampleInt & MAX_INT24
			sampleInt = MIN_INT24 + offset
		}

		samplesFloat[idx] = scaling * float64(sampleInt)
	}

	return samplesFloat, nil
}

/*
 * Convert samples to bytes, encoding them as 32-bit LPCM values.
 */
func samplesToBytesLPCM32(samples []float64) ([]byte, error) {
	numSamples := len(samples)
	samplesInt := make([]int32, numSamples)
	const delta = math.MaxInt32 - math.MinInt32
	scale := 0.5 * float64(delta)

	/*
	 * Iterate over the samples and convert them into integer representation.
	 */
	for i, sample := range samples {

		/*
		 * Make sure that limits are not exceeded.
		 */
		if sample < -1.0 {
			sample = -1.0
		} else if sample > 1.0 {
			sample = 1.0
		}

		tmp := int64(scale * sample)

		/*
		 * Make sure that limits are not exceeded.
		 */
		if tmp > math.MaxInt32 {
			tmp = math.MaxInt32
		} else if tmp < math.MinInt32 {
			tmp = math.MinInt32
		}

		samplesInt[i] = int32(tmp)
	}

	buf := createBuffer()
	err := binary.Write(buf, binary.LittleEndian, samplesInt)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to convert samples: %s", msg)
	} else {
		data := buf.Bytes()
		return data, nil
	}

}

/*
 * Convert bytes, encoded as 32-bit LPCM values, to samples.
 */
func bytesToSamplesLPCM32(data []byte) ([]float64, error) {
	numBytes := len(data)
	numBytes64 := uint64(numBytes)
	numSamples := numBytes64 >> 2
	samplesInt := make([]int32, numSamples)
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, samplesInt)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to decode LPCM32 data: %s", msg)
	} else {
		samplesFloat := make([]float64, numSamples)
		scaling := 2.0 / (math.MaxInt32 - math.MinInt32)

		/*
		 * Convert samples to floating-point representation.
		 */
		for i, sample := range samplesInt {
			samplesFloat[i] = scaling * float64(sample)
		}

		return samplesFloat, nil
	}

}

/*
 * Convert samples to bytes, encoding them as 32-bit IEEE floating-point values.
 */
func samplesToBytesIEEE32(samples []float64) ([]byte, error) {
	numSamples := len(samples)
	samples32 := make([]float32, numSamples)

	/*
	 * Iterate over the samples and convert them into integer representation.
	 */
	for i, sample := range samples {

		/*
		 * Make sure that limits are not exceeded.
		 */
		if sample < -1.0 {
			sample = -1.0
		} else if sample > 1.0 {
			sample = 1.0
		}

		samples32[i] = float32(sample)
	}

	buf := createBuffer()
	err := binary.Write(buf, binary.LittleEndian, samples32)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to convert samples: %s", msg)
	} else {
		data := buf.Bytes()
		return data, nil
	}

}

/*
 * Convert bytes, encoded as 32-bit IEEE floating-point values, to samples.
 */
func bytesToSamplesIEEE32(data []byte) ([]float64, error) {
	numBytes := len(data)
	numBytes64 := uint64(numBytes)
	numSamples := numBytes64 >> 2
	samplesFloat32 := make([]float32, numSamples)
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, samplesFloat32)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to decode 32-bit IEEE floating-point data: %s", msg)
	} else {
		samplesFloat := make([]float64, numSamples)

		/*
		 * Convert samples to 64-bit floating-point representation.
		 */
		for i, sample := range samplesFloat32 {
			samplesFloat[i] = float64(sample)
		}

		return samplesFloat, nil
	}

}

/*
 * Convert samples to bytes, encoding them as 64-bit IEEE floating-point values.
 */
func samplesToBytesIEEE64(samples []float64) ([]byte, error) {
	buf := createBuffer()
	err := binary.Write(buf, binary.LittleEndian, samples)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to convert samples: %s", msg)
	} else {
		data := buf.Bytes()
		return data, nil
	}

}

/*
 * Convert bytes, encoded as 64-bit IEEE floating-point values, to samples.
 */
func bytesToSamplesIEEE64(data []byte) ([]float64, error) {
	numBytes := len(data)
	numBytes64 := uint64(numBytes)
	numSamples := numBytes64 >> 3
	samplesFloat64 := make([]float64, numSamples)
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.LittleEndian, samplesFloat64)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to decode 64-bit IEEE floating-point data: %s", msg)
	} else {
		return samplesFloat64, nil
	}

}

/*
 * Convert samples to bytes, given a sample format and bit depth.
 */
func samplesToBytes(samples []float64, sampleFormat uint16, bitDepth uint16) ([]byte, error) {

	/*
	 * Decide on the sample format.
	 */
	switch sampleFormat {
	case AUDIO_PCM:

		/*
		 * Decide on the bit depth.
		 */
		switch bitDepth {
		case 8:
			res, err := samplesToBytesLPCM8(samples)
			return res, err
		case 16:
			res, err := samplesToBytesLPCM16(samples)
			return res, err
		case 24:
			res, err := samplesToBytesLPCM24(samples)
			return res, err
		case 32:
			res, err := samplesToBytesLPCM32(samples)
			return res, err
		default:
			return nil, fmt.Errorf("Unsupported bit depth for audio in LPCM format: %d", bitDepth)
		}

	case AUDIO_IEEE_FLOAT:

		/*
		 * Decide on the bit depth.
		 */
		switch bitDepth {
		case 32:
			res, err := samplesToBytesIEEE32(samples)
			return res, err
		case 64:
			res, err := samplesToBytesIEEE64(samples)
			return res, err
		default:
			return nil, fmt.Errorf("Unsupported bit depth for audio in IEEE floating-point format: %d", bitDepth)
		}

	default:
		return nil, fmt.Errorf("Unknown sample format: %#04x", sampleFormat)
	}

}

/*
 * Convert bytes to samples, given a sample format and bit depth.
 */
func bytesToSamples(data []byte, sampleFormat uint16, bitDepth uint16) ([]float64, error) {

	/*
	 * Decide on the sample format.
	 */
	switch sampleFormat {
	case AUDIO_PCM:

		/*
		 * Decide on the bit depth.
		 */
		switch bitDepth {
		case 8:
			res, err := bytesToSamplesLPCM8(data)
			return res, err
		case 16:
			res, err := bytesToSamplesLPCM16(data)
			return res, err
		case 24:
			res, err := bytesToSamplesLPCM24(data)
			return res, err
		case 32:
			res, err := bytesToSamplesLPCM32(data)
			return res, err
		default:
			return nil, fmt.Errorf("Unsupported bit depth for audio in LPCM format: %d", bitDepth)
		}

	case AUDIO_IEEE_FLOAT:

		/*
		 * Decide on the bit depth.
		 */
		switch bitDepth {
		case 32:
			res, err := bytesToSamplesIEEE32(data)
			return res, err
		case 64:
			res, err := bytesToSamplesIEEE64(data)
			return res, err
		default:
			return nil, fmt.Errorf("Unsupported bit depth for audio in IEEE floating-point format: %d", bitDepth)
		}

	default:
		return nil, fmt.Errorf("Unknown sample format: %#04x", sampleFormat)
	}

}

/*
 * Returns the sample depth of this wave file in bits.
 */
func (this *fileStruct) BitDepth() uint16 {
	return this.bitDepth
}

/*
 * Returns the contents of this wave file as a byte slice.
 */
func (this *fileStruct) Bytes() ([]byte, error) {
	channelCount := len(this.channels)
	channelCount16 := uint16(channelCount)
	channelCount32 := uint32(channelCount)
	bitDepth := this.bitDepth
	sampleFormat := this.sampleFormat
	sampleRate := this.sampleRate
	sampleSize32 := uint32(bitDepth / BITS_PER_BYTE)
	sampleSize64 := uint64(sampleSize32)
	blockAlign := sampleSize32 * channelCount32
	blockAlign16 := uint16(blockAlign)
	byteRate := sampleRate * blockAlign
	samples := channelsToSamples(this.channels)
	numSamples := len(samples)
	data, err := samplesToBytes(samples, sampleFormat, bitDepth)

	/*
	 * Check if conversion was successful.
	 */
	if err != nil {
		return nil, err
	} else {
		idRIFF := uint32(ID_RIFF)
		numSamples32 := uint32(numSamples)
		numSamples64 := uint64(numSamples)
		dataBytes32 := sampleSize32 * numSamples32
		dataBytes64 := sampleSize64 * numSamples64
		riffSize64 := dataBytes64 + (MIN_TOTAL_HEADER_SIZE - MIN_CHUNK_HEADER_SIZE)
		riffSize32 := uint32(riffSize64)
		requiresRF64 := riffSize64 > math.MaxUint32

		/*
		 * If we write an RF64 file, replace RIFF chunk ID with 'RF64' and set 32-bit size to math.MaxUint32 (0xffffffff).
		 */
		if requiresRF64 {
			idRIFF = uint32(ID_RIFF64)
			riffSize32 = math.MaxUint32
		}

		/*
		 * Create RIFF header.
		 */
		hdrRiff := riffHeader{
			ChunkID:   idRIFF,
			ChunkSize: riffSize32,
			Format:    FORMAT_WAVE,
		}

		/*
		 * Create data size header.
		 */
		hdrDataSize := dataSizeHeader{
			ChunkID:     ID_DATASIZE,
			ChunkSize:   MIN_DATASIZE_CHUNK_SIZE,
			SizeRIFF:    riffSize64,
			SizeData:    dataBytes64,
			SampleCount: numSamples64,
			TableLength: 0,
		}

		/*
		 * Create format header.
		 */
		hdrFormat := formatHeader{
			ChunkID:      ID_FORMAT,
			ChunkSize:    MIN_CHUNK_SIZE_FORMAT,
			AudioFormat:  sampleFormat,
			ChannelCount: channelCount16,
			SampleRate:   sampleRate,
			ByteRate:     byteRate,
			BlockAlign:   blockAlign16,
			BitDepth:     bitDepth,
		}

		/*
		 * Create data header.
		 */
		hdrData := dataHeader{
			ChunkID:   ID_DATA,
			ChunkSize: dataBytes32,
		}

		buf := createBuffer()
		binary.Write(buf, binary.LittleEndian, hdrRiff)

		/*
		 * If we write an RF64 file, write mandatory data size chunk.
		 */
		if requiresRF64 {
			binary.Write(buf, binary.LittleEndian, hdrDataSize)
		}

		binary.Write(buf, binary.LittleEndian, hdrFormat)
		binary.Write(buf, binary.LittleEndian, hdrData)
		buf.Write(data)
		content := buf.Bytes()
		return content, nil
	}

}

/*
 * Returns a reference to the requested channel.
 */
func (this *fileStruct) Channel(id uint16) (Channel, error) {
	channelCount := this.ChannelCount()

	/*
	 * Check whether the requested channel is available in this wave file.
	 */
	if id >= channelCount {
		return nil, fmt.Errorf("No channel with id = %d in this wave file with channel count %d.", id, channelCount)
	} else {
		return this.channels[id], nil
	}

}

/*
 * Returns the number of channels available in this wave file.
 */
func (this *fileStruct) ChannelCount() uint16 {
	n := len(this.channels)
	n16 := uint16(n)
	return n16
}

/*
 * Returns the format code of the sample format of this wave file.
 */
func (this *fileStruct) SampleFormat() uint16 {
	return this.sampleFormat
}

/*
 * Returns the sample rate of this wave file in Hertz.
 */
func (this *fileStruct) SampleRate() uint32 {
	return this.sampleRate
}

/*
 * Creates an empty channel.
 */
func createChannel() Channel {
	channel := channelStruct{}
	return &channel
}

/*
 * Skips over a number of bytes in the file.
 */
func skipData(reader *bytes.Reader, numBytes uint64) error {
	max := uint64(math.MaxInt32)

	/*
	 * Check if we can seek this far.
	 */
	if numBytes > max {
		return fmt.Errorf("Cannot skip more than %d bytes.", max)
	} else {
		signedBytes := int64(numBytes)
		mode := io.SeekCurrent
		reader.Seek(signedBytes, mode)
		return nil
	}

}

/*
 * Look ahead to the next chunk.
 */
func lookaheadChunk(reader *bytes.Reader) (*chunkHeader, error) {
	hdrChunk := chunkHeader{}
	err := binary.Read(reader, binary.LittleEndian, &hdrChunk)

	/*
	 * Check if chunk header was read.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to read chunk header: %s", msg)
	} else {
		mode := io.SeekCurrent
		_, err = reader.Seek(-MIN_CHUNK_HEADER_SIZE, mode)
		return &hdrChunk, err
	}

}

/*
 * Skip over chunks until you find one with a certain ID.
 */
func skipToChunk(reader *bytes.Reader, chunkId uint32) error {
	abort := false

	/*
	 * Skip over chunks until we find the one we expect.
	 */
	for !abort {
		hdrChunk, err := lookaheadChunk(reader)

		/*
		 * Check if lookahead was successful.
		 */
		if err != nil {
			return err
		} else {
			id := hdrChunk.ChunkID

			/*
			 * If we found the right chunk, abort, otherwise skip over it.
			 */
			if id == chunkId {
				abort = true
			} else {
				size := hdrChunk.ChunkSize
				sizeLSB := size % 2

				/*
				 * If chunk size is not even, we have to read one
				 * additional byte of padding.
				 */
				if sizeLSB != 0 {
					size += 1
				}

				amount := uint64(size) + MIN_CHUNK_HEADER_SIZE
				err = skipData(reader, amount)

				/*
				 * Check if skipping failed.
				 */
				if err != nil {
					return err
				}

			}

		}

	}

	return nil
}

/*
 * Read RIFF header from file and validate it.
 */
func readHeaderRIFF(reader *bytes.Reader, totalSize uint64) (*riffHeader, error) {
	hdrRiff := riffHeader{}
	err := binary.Read(reader, binary.LittleEndian, &hdrRiff)

	/*
	 * Check if RIFF header was read.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to read RIFF header: %s", msg)
	} else {
		expectedRiffChunkSize64 := totalSize - 8
		expectedRiffChunkSize32 := uint32(expectedRiffChunkSize64)

		/*
		 * Check RIFF header for validity.
		 */
		if hdrRiff.ChunkID != ID_RIFF {
			return nil, fmt.Errorf("RIFF header contains invalid chunk id. Expected %#08x or %#08x, found %#08x.", ID_RIFF, ID_RIFF64, hdrRiff.ChunkID)
		} else if (expectedRiffChunkSize64 < math.MaxUint32 && hdrRiff.ChunkSize != expectedRiffChunkSize32) || (hdrRiff.ChunkID == ID_RIFF64 && hdrRiff.ChunkSize != math.MaxUint32) {
			return nil, fmt.Errorf("RIFF header contains invalid chunk size. Expected %#08x (or %#08x for 'RF64'), found %#08x.", expectedRiffChunkSize32, uint32(math.MaxUint32), hdrRiff.ChunkSize)
		} else if hdrRiff.Format != FORMAT_WAVE {
			return nil, fmt.Errorf("RIFF header contains invalid format. Expected %#08x, found %#08x.", FORMAT_WAVE, hdrRiff.Format)
		} else {
			return &hdrRiff, nil
		}

	}

}

/*
 * Read data size header from file and validate it.
 */
func readHeaderDataSize(reader *bytes.Reader, totalSize uint64) (*dataSizeHeader, error) {
	hdrDataSize := dataSizeHeader{}
	err := binary.Read(reader, binary.LittleEndian, &hdrDataSize)

	/*
	 * Check if data size header was read.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to read data size header: %s", msg)
	} else {
		expectedRiffChunkSize := totalSize - 8

		/*
		 * Check data size header for validity.
		 */
		if hdrDataSize.ChunkID != ID_DATASIZE {
			return nil, fmt.Errorf("Data size header contains invalid chunk id. Expected %#08x, found %#08x.", ID_DATASIZE, hdrDataSize.ChunkID)
		} else if hdrDataSize.ChunkSize < MIN_DATASIZE_CHUNK_SIZE {
			return nil, fmt.Errorf("Data size header has too small size. Expected at least %#08x, found %#08x.", MIN_DATASIZE_CHUNK_SIZE, hdrDataSize.ChunkSize)
		} else if hdrDataSize.SizeRIFF != expectedRiffChunkSize {
			return nil, fmt.Errorf("Unexpected RIFF chunk size in data size header. Expected %#08x, found %0#8x.", expectedRiffChunkSize, hdrDataSize.SizeRIFF)
		} else {
			return &hdrDataSize, nil
		}

	}

}

/*
 * Read format header from file and validate it.
 */
func readHeaderFormat(reader *bytes.Reader) (*formatHeader, error) {
	hdrFormat := formatHeader{}
	err := binary.Read(reader, binary.LittleEndian, &hdrFormat)

	/*
	 * Check if format header was read.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to read format header: %s", msg)
	} else {
		channelCount := hdrFormat.ChannelCount
		bitDepth := hdrFormat.BitDepth
		sampleRate := hdrFormat.SampleRate
		frameSize := channelCount * bitDepth
		expectedBlockAlign32 := uint32(frameSize / BITS_PER_BYTE)
		expectedBlockAlign16 := uint16(expectedBlockAlign32)
		expectedByteRate := expectedBlockAlign32 * sampleRate
		chunkSize := int64(hdrFormat.ChunkSize)
		numBytesSkip := chunkSize - MIN_CHUNK_SIZE_FORMAT

		/*
		 * Skip optional fields in the format header.
		 */
		if numBytesSkip > 0 {

			/*
			 * If this is even, we need to skip one more.
			 */
			if (numBytesSkip % 2) == 0 {
				numBytesSkip += 1
			}

			amount := uint64(numBytesSkip)
			skipData(reader, amount)
		}

		/*
		 * Check format header for validity.
		 */
		if hdrFormat.ChunkID != ID_FORMAT {
			return nil, fmt.Errorf("Format header contains invalid chunk id. Expected %#08x, found %#08x.", ID_FORMAT, hdrFormat.ChunkID)
		} else if hdrFormat.ChunkSize < MIN_CHUNK_SIZE_FORMAT {
			return nil, fmt.Errorf("Format header contains invalid chunk size. Expected at least %#08x, found %#08x.", MIN_CHUNK_SIZE_FORMAT, hdrFormat.ChunkSize)
		} else if hdrFormat.AudioFormat != AUDIO_PCM && hdrFormat.AudioFormat != AUDIO_IEEE_FLOAT {
			return nil, fmt.Errorf("Format header contains invalid audio format. Expected %#04x or %#04x, found %#04x.", AUDIO_PCM, AUDIO_IEEE_FLOAT, hdrFormat.AudioFormat)
		} else if hdrFormat.ByteRate != expectedByteRate {
			return nil, fmt.Errorf("Format header contains invalid byte rate. Expected %#08x, found %#08x.", expectedByteRate, hdrFormat.ByteRate)
		} else if hdrFormat.BlockAlign != expectedBlockAlign16 {
			return nil, fmt.Errorf("Format header contains invalid block align. Expected %#04x, found %#04x.", expectedBlockAlign16, hdrFormat.BlockAlign)
		} else if hdrFormat.AudioFormat == AUDIO_PCM && hdrFormat.BitDepth != 8 && hdrFormat.BitDepth != 16 && hdrFormat.BitDepth != 24 && hdrFormat.BitDepth != 32 {
			return nil, fmt.Errorf("Format header contains invalid bit depth for PCM format. Expected %#04x or %#04x or %#04x or %#04x, found %#04x.", 8, 16, 24, 32, hdrFormat.BitDepth)
		} else if hdrFormat.AudioFormat == AUDIO_IEEE_FLOAT && hdrFormat.BitDepth != 32 && hdrFormat.BitDepth != 64 {
			return nil, fmt.Errorf("Format header contains invalid bit depth for IEEE floating-point format. Expected %#04x or %#04x, found %#04x.", 32, 64, hdrFormat.BitDepth)
		} else {
			return &hdrFormat, nil
		}

	}

}

/*
 * Read data header from file and validate it.
 */
func readHeaderData(reader *bytes.Reader, totalSize uint64) (*dataHeader, error) {
	hdrData := dataHeader{}
	err := binary.Read(reader, binary.LittleEndian, &hdrData)

	/*
	 * Check if data header was read.
	 */
	if err != nil {
		msg := err.Error()
		return nil, fmt.Errorf("Failed to read data header: %s", msg)
	} else {
		maxDataLength := totalSize - MIN_TOTAL_HEADER_SIZE
		maxDataLength32 := uint32(maxDataLength)
		chunkId := hdrData.ChunkID
		chunkSize := hdrData.ChunkSize

		/*
		 * Check data header for validity.
		 */
		if chunkId != ID_DATA {
			return nil, fmt.Errorf("Data header contains invalid chunk id. Expected %#08x, found %#08x.", ID_DATA, chunkId)
		} else if (chunkSize > maxDataLength32) && (chunkSize != math.MaxUint32) {
			return nil, fmt.Errorf("Data header contains invalid chunk size. Expected at most %#08x (or %#08x), found %#08x.", maxDataLength32, uint32(math.MaxUint32), chunkSize)
		} else {
			return &hdrData, nil
		}

	}

}

/*
 * Create an empty wave file with the desired sample rate, sample format, bit depth and channel count.
 */
func CreateEmpty(sampleRate uint32, sampleFormat uint16, bitDepth uint16, channelCount uint16) (File, error) {

	/*
	 * Check if sample format is valid.
	 */
	if sampleFormat != AUDIO_PCM && sampleFormat != AUDIO_IEEE_FLOAT {
		return nil, fmt.Errorf("Unknown sample format: %#04x - Expected either %#04x or %#04x.", sampleFormat, AUDIO_PCM, AUDIO_IEEE_FLOAT)
	} else {

		/*
		 * Check if bit depth is valid for sample format.
		 */
		if sampleFormat == AUDIO_PCM && bitDepth != 8 && bitDepth != 16 && bitDepth != 24 && bitDepth != 32 {
			return nil, fmt.Errorf("Bit depth must be either %d or %d or %d or %d for audio in PCM format.", 8, 16, 24, 32)
		} else if sampleFormat == AUDIO_IEEE_FLOAT && bitDepth != 32 && bitDepth != 64 {
			return nil, fmt.Errorf("Bit depth must be either %d or %d for audio in IEEE floating-point format.", 32, 64)
		} else {
			channels := make([]Channel, channelCount)

			/*
			 * Create channels for this wave file.
			 */
			for i := uint16(0); i < channelCount; i++ {
				channels[i] = createChannel()
			}

			/*
			 * Create wave file structure.
			 */
			file := fileStruct{
				bitDepth:     bitDepth,
				sampleFormat: sampleFormat,
				sampleRate:   sampleRate,
				channels:     channels,
			}

			return &file, nil
		}

	}

}

/*
 * Creates a wave file from the contents of a byte buffer.
 */
func FromBuffer(buffer []byte) (File, error) {
	totalSize := len(buffer)
	totalSize64 := uint64(totalSize)
	reader := bytes.NewReader(buffer)
	hdrRiff, err := readHeaderRIFF(reader, totalSize64)

	/*
	 * Check if RIFF header was successfully read.
	 */
	if err != nil {
		return nil, err
	} else {
		hdrDataSize := &dataSizeHeader{}

		/*
		 * If this is an 'RF64' file, read data size header.
		 */
		if hdrRiff.ChunkID == ID_RIFF64 {
			hdrDataSize, err = readHeaderDataSize(reader, totalSize64)

			/*
			 * If data size header was successfully read, skip over optional table entries.
			 */
			if err != nil {
				msg := err.Error()
				return nil, fmt.Errorf("Failed to read data size chunk: %s", msg)
			} else {
				numEntries := hdrDataSize.TableLength
				numEntries64 := uint64(numEntries)
				bytesToSkip := LENGTH_DATASIZE_TABLE_ENTRIES * numEntries64
				err := skipData(reader, bytesToSkip)

				/*
				 * Check if we successfully skipped over the table entries.
				 */
				if err != nil {
					msg := err.Error()
					return nil, fmt.Errorf("Failed to skip over data size table entries: %s", msg)
				}

			}

		}

		hdrFormat, err := readHeaderFormat(reader)

		/*
		 * Check if format header was successfully read.
		 */
		if err != nil {
			return nil, err
		} else {
			bitDepth := hdrFormat.BitDepth
			sampleFormat := hdrFormat.AudioFormat
			err = skipToChunk(reader, ID_DATA)

			/*
			 * Check if we successfully arrived at the data chunk.
			 */
			if err != nil {
				msg := err.Error()
				return nil, fmt.Errorf("Failed to locate data chunk: %s", msg)
			} else {
				hdrData, err := readHeaderData(reader, totalSize64)
				chunkSize32 := hdrData.ChunkSize
				chunkSize64 := uint64(chunkSize32)

				/*
				 * If this is an 'RF64' file, take chunk size from data size header.
				 */
				if hdrRiff.ChunkID == ID_RIFF64 {
					chunkSize64 = hdrDataSize.SizeData
				}

				/*
				 * Check if data header was successfully read.
				 */
				if err != nil {
					return nil, err
				} else {
					sampleData := make([]byte, chunkSize64)
					_, err = reader.Read(sampleData)

					/*
					 * Check if sample data was read.
					 */
					if err != nil {
						msg := err.Error()
						return nil, fmt.Errorf("Failed to read sample data: %s", msg)
					} else {
						samples, err := bytesToSamples(sampleData, sampleFormat, bitDepth)

						/*
						 * Check if sample data was decoded.
						 */
						if err != nil {
							msg := err.Error()
							return nil, fmt.Errorf("Failed to decode sample data: %s", msg)
						} else {
							channelCount := hdrFormat.ChannelCount
							channels := samplesToChannels(samples, channelCount)

							/*
							 * Create a new data structure representing the contents of the wave file.
							 */
							file := fileStruct{
								bitDepth:     bitDepth,
								sampleFormat: sampleFormat,
								sampleRate:   hdrFormat.SampleRate,
								channels:     channels,
							}

							return &file, nil
						}

					}

				}

			}

		}

	}

}

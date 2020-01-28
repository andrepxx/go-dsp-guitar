package oversampling

import (
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/filter"
	"github.com/andrepxx/go-dsp-guitar/resample"
)

/*
 * Global constants.
 */
const (
	ATTENUATION_HALF_DECIBEL     = 0.9440608762859234
	LOOKAHEAD_SAMPLES_ONE_SIDE   = 4
	LOOKAHEAD_SAMPLES_BOTH_SIDES = 2 * LOOKAHEAD_SAMPLES_ONE_SIDE
)

/*
 * An oversampler / decimator increases or decreases the sample rate of a
 * signal by a constant factor.
 *
 * When oversampling the signal, band-limited interpolation is applied.
 *
 * When decimating (downsampling) the signal, a band-limiting filter is
 * applied to prevent aliasing.
 */
type OversamplerDecimator interface {
	Oversample(in []float64, out []float64) error
	Decimate(in []float64, out []float64) error
}

/*
 * Data structure implementing an oversampler / decimator.
 */
type oversamplerDecimatorStruct struct {
	factor               uint32
	antiAliasingFilter   filter.Filter
	attenuationFactor    float64
	bufferPreUpsampling  []float64
	bufferPostUpsampling []float64
	bufferPreDecimation  []float64
}

/*
 * Oversamples the signal in the input buffer by an oversampling factor.
 *
 * The resulting output is M = factor * N samples long when the input provided
 * is N samples long.
 *
 * The oversampling is stateful, since it requires both some lookahead on the
 * input (and therefore needs to introduce a delay of a few samples) and some
 * input samples from past invocations.
 */
func (this *oversamplerDecimatorStruct) Oversample(in []float64, out []float64) error {
	factor32 := this.factor
	factor := int(factor32)

	/*
	 * Check if signal has to be oversampled in time.
	 */
	if factor <= 1 {
		copy(out, in)
		return nil
	} else {
		numInputSamples := len(in)
		numInputSamples32 := uint32(numInputSamples)
		expectedNumOutputSamples := numInputSamples * factor
		expectedNumOutputSamples32 := uint32(expectedNumOutputSamples)
		numOutputSamples := len(out)
		numOutputSamples32 := uint32(numOutputSamples)

		/*
		 * Ensure that input and output buffer has the correct size.
		 */
		if numOutputSamples32 != expectedNumOutputSamples32 {
			template := "Error while oversampling: Expected output buffer of size %d (= %d * %d), but buffer has size %d."
			return fmt.Errorf(template, expectedNumOutputSamples32, numInputSamples32, factor, numOutputSamples32)
		} else {
			bufferPre := this.bufferPreUpsampling
			bufferPreSize := numInputSamples + LOOKAHEAD_SAMPLES_BOTH_SIDES

			/*
			 * Make sure the pre-upsampling buffer has the correct size.
			 */
			if len(bufferPre) != bufferPreSize {
				bufferPre = make([]float64, bufferPreSize)
				this.bufferPreUpsampling = bufferPre
			}

			tailStart := bufferPreSize - LOOKAHEAD_SAMPLES_BOTH_SIDES
			copy(bufferPre[0:LOOKAHEAD_SAMPLES_BOTH_SIDES], bufferPre[tailStart:bufferPreSize])
			copy(bufferPre[LOOKAHEAD_SAMPLES_BOTH_SIDES:bufferPreSize], in)
			bufferPost := this.bufferPostUpsampling
			bufferPostSize := ((bufferPreSize - 1) * factor) + 1

			/*
			 * Make sure the post-upsampling buffer has the correct size.
			 */
			if len(bufferPost) != bufferPostSize {
				bufferPost = make([]float64, bufferPostSize)
				this.bufferPostUpsampling = bufferPost
			}

			resample.Oversample(bufferPre, bufferPost, factor32)
			idxStart := LOOKAHEAD_SAMPLES_ONE_SIDE * factor
			idxEnd := idxStart + expectedNumOutputSamples
			copy(out, bufferPost[idxStart:idxEnd])
			return nil
		}

	}

}

/*
 * Decimates the signal in the input buffer by an oversampling factor.
 *
 * The resulting output is N = M / factor samples long when the input provided
 * is M samples long.
 *
 * The decimation is stateful, since it makes use of a stateful anti-aliasing
 * filter.
 */
func (this *oversamplerDecimatorStruct) Decimate(in []float64, out []float64) error {
	factor32 := this.factor
	factor := int(factor32)

	/*
	 * Check if signal has to be decimated in time.
	 */
	if factor <= 1 {
		copy(out, in)
		return nil
	} else {
		numInputSamples := len(in)
		buffer := this.bufferPreDecimation

		/*
		 * Ensure that the output buffer for the anti-aliasing filter
		 * is as long as the input buffer.
		 */
		if len(buffer) != numInputSamples {
			buffer = make([]float64, numInputSamples)
			this.bufferPreDecimation = buffer
		}

		flt := this.antiAliasingFilter
		err := flt.Process(in, buffer)

		/*
		 * Check if an error occured.
		 */
		if err != nil {
			msg := err.Error()
			return fmt.Errorf("Error while applying anti-aliasing filter for decimation: %s", msg)
		} else {
			bufferSize := len(buffer)

			/*
			 * Decimate the output by taking one sample and dropping N - 1
			 * samples, where N is the oversampling factor.
			 */
			for i := range out {
				idx := factor * i

				/*
				 * Check if we are still within the bounds of the buffer.
				 */
				if idx < bufferSize {
					out[i] = ATTENUATION_HALF_DECIBEL * buffer[idx]
				} else {
					out[i] = 0.0
				}

			}

			return nil
		}

	}

}

/*
 * Creates an oversampler / decimator with the requested oversampling factor.
 *
 * The oversampling factor can be either 1, 2 or 4.
 *
 * (An oversampler / decimator with an oversampling factor of one technically
 * just copies buffers when oversampling / decimating though.)
 */
func CreateOversamplerDecimator(factor uint32) OversamplerDecimator {

	/*
	 * The oversampling factor determines the coefficients for the
	 * anti-aliasing filter used before decimation (= downsampling).
	 */
	switch factor {
	case 1:

		/*
		 * An oversampler / decimator with an oversampling
		 * factor of 1.
		 */
		osd := oversamplerDecimatorStruct{
			factor:             1,
			antiAliasingFilter: nil,
			attenuationFactor:  1.0,
		}

		return &osd
	case 2:

		/*
		 * Anti-aliasing filter for decimation after 2-times
		 * oversampling.
		 *
		 * - Order: 77
		 * - Passband: 0 to 0.4 * fs
		 * - Ripple: less than +/- 0.5 dB
		 * - Stopband: from 0.5 * fs
		 * - Attenuation: more than 120 dB
		 *
		 * Where fs is the sample rate after decimation.
		 *
		 * Examples:
		 *
		 * Sample rate | Oversampled |  Passband | Stopband
		 * --------------------------------------------------
		 *    44.1 kHz |    88.2 kHz | 17.64 kHz | 22.05 kHz
		 *    48.0 kHz |    96.0 kHz | 19.20 kHz | 24.00 kHz
		 *    88.2 kHz |   176.4 kHz | 35.28 kHz | 44.10 kHz
		 *    96.0 kHz |   192.0 kHz | 38.40 kHz | 48.00 kHz
		 *   176.4 kHz |   352.8 kHz | 70.56 kHz | 88.20 kHz
		 *   192.0 kHz |   384.0 kHz | 76.80 kHz | 96.00 kHz
		 */
		coeffs := []float64{
			-0.00003492934784890941,
			-0.0003392120149044798,
			-0.0014714158707716568,
			-0.0038999530054612506,
			-0.006911343940235415,
			-0.008083007705624154,
			-0.005071878685632956,
			0.0010436432381931216,
			0.005133162788276673,
			0.0029441192777868346,
			-0.002966456827156262,
			-0.0049140011265012915,
			0.0002504710726990459,
			0.005677026805555885,
			0.0029644609183865525,
			-0.00502776088831408,
			-0.006372346916246025,
			0.0025325196219111606,
			0.009161633713407162,
			0.0019414998184190532,
			-0.010183325233642465,
			-0.007923283926229728,
			0.008288108595786142,
			0.014272756343260685,
			-0.0026415624384956015,
			-0.019218660155603737,
			-0.006932661398171457,
			0.020610554179589805,
			0.01985948065927744,
			-0.016017276726217215,
			-0.03475499073668333,
			0.0026332590387740207,
			0.049675468036826695,
			0.02427989874123878,
			-0.062421162066919444,
			-0.08025548852929663,
			0.07100233144132687,
			0.30929817925728875,
			0.4259645026667285,
			0.30929817925728875,
			0.07100233144132687,
			-0.08025548852929663,
			-0.062421162066919444,
			0.02427989874123878,
			0.049675468036826695,
			0.0026332590387740207,
			-0.03475499073668333,
			-0.016017276726217215,
			0.01985948065927744,
			0.020610554179589805,
			-0.006932661398171457,
			-0.019218660155603737,
			-0.0026415624384956015,
			0.014272756343260685,
			0.008288108595786142,
			-0.007923283926229728,
			-0.010183325233642465,
			0.0019414998184190532,
			0.009161633713407162,
			0.0025325196219111606,
			-0.006372346916246025,
			-0.00502776088831408,
			0.0029644609183865525,
			0.005677026805555885,
			0.0002504710726990459,
			-0.0049140011265012915,
			-0.002966456827156262,
			0.0029441192777868346,
			0.005133162788276673,
			0.0010436432381931216,
			-0.005071878685632956,
			-0.008083007705624154,
			-0.006911343940235415,
			-0.0038999530054612506,
			-0.0014714158707716568,
			-0.0003392120149044798,
			-0.00003492934784890941,
		}

		flt := filter.FromCoefficients(coeffs, 0, "Anti-aliasing filter for 2-times oversampling")

		/*
		 * An oversampler / decimator with an oversampling
		 * factor of 2.
		 */
		osd := oversamplerDecimatorStruct{
			factor:             2,
			antiAliasingFilter: flt,
			attenuationFactor:  ATTENUATION_HALF_DECIBEL,
		}

		return &osd
	case 4:

		/*
		 * Anti-aliasing filter for decimation after 4-times
		 * oversampling.
		 *
		 * - Order: 155
		 * - Passband: 0 to 0.4 * fs
		 * - Ripple: less than +/- 0.5 dB
		 * - Stopband: from 0.5 * fs
		 * - Attenuation: more than 120 dB
		 *
		 * Where fs is the sample rate after decimation.
		 *
		 * Examples:
		 *
		 * Sample rate | Oversampled |  Passband | Stopband
		 * --------------------------------------------------
		 *    44.1 kHz |   176.4 kHz | 17.64 kHz | 22.05 kHz
		 *    48.0 kHz |   192.0 kHz | 19.20 kHz | 24.00 kHz
		 *    88.2 kHz |   352.8 kHz | 35.28 kHz | 44.10 kHz
		 *    96.0 kHz |   384.0 kHz | 38.40 kHz | 48.00 kHz
		 *   176.4 kHz |   705.6 kHz | 70.56 kHz | 88.20 kHz
		 *   192.0 kHz |   768.0 kHz | 76.80 kHz | 96.00 kHz
		 */
		coeffs := []float64{
			-0.0000015021121037413662,
			-0.000014247930232761388,
			-0.00005540231978931542,
			-0.00015630206030270148,
			-0.0003584399777348438,
			-0.000705026640710802,
			-0.0012235616110217874,
			-0.001903660659808354,
			-0.00267726234331065,
			-0.003411584910561322,
			-0.003924344675899259,
			-0.004025045645232759,
			-0.0035762101728860417,
			-0.0025573008067263044,
			-0.0011066687495094368,
			0.0004821972926674136,
			0.0018221416962137336,
			0.002543137384860401,
			0.00242138962866836,
			0.0014804833709829318,
			0.000018970679160808765,
			-0.0014609367040901548,
			-0.0024134283298567795,
			-0.0024492254183795556,
			-0.0015017106353334075,
			0.00011357732553534381,
			0.0017859787871056622,
			0.0028277613883314974,
			0.0027499128054293363,
			0.0014872096489652376,
			-0.0005294306298690314,
			-0.002503516126880322,
			-0.003575536194224602,
			-0.003187701643215724,
			-0.00135969080293838,
			0.0012559883020625854,
			0.0035898853367056904,
			0.004580083649141936,
			0.0036424372986039993,
			0.0009834739417648997,
			-0.0024017684965412264,
			-0.005083213814497091,
			-0.005784183213696061,
			-0.00396945510709691,
			-0.00016979882374028044,
			0.004133247336931644,
			0.0070622101801412545,
			0.0071407896959270085,
			0.004004024888678994,
			-0.0013073464277146174,
			-0.006649326836701509,
			-0.009608632396950771,
			-0.008557172161681335,
			-0.00348067806403019,
			0.003816684112995466,
			0.010298649251403313,
			0.012908606169822902,
			0.00994335100177495,
			0.001997586597194608,
			-0.007994368492098312,
			-0.01575380865287906,
			-0.01738598716205188,
			-0.011180107332533986,
			0.0012948642543910843,
			0.015369959683332234,
			0.024832317669641207,
			0.024391753256398193,
			0.012155808794391754,
			-0.00890111973242649,
			-0.031201026812293402,
			-0.04467639062071552,
			-0.04014181689513159,
			-0.01278803733109894,
			0.035484245817926856,
			0.0958427608403164,
			0.15465705312784375,
			0.19739158968717124,
			0.21300669185029408,
			0.19739158968717124,
			0.15465705312784375,
			0.0958427608403164,
			0.035484245817926856,
			-0.01278803733109894,
			-0.04014181689513159,
			-0.04467639062071552,
			-0.031201026812293402,
			-0.00890111973242649,
			0.012155808794391754,
			0.024391753256398193,
			0.024832317669641207,
			0.015369959683332234,
			0.0012948642543910843,
			-0.011180107332533986,
			-0.01738598716205188,
			-0.01575380865287906,
			-0.007994368492098312,
			0.001997586597194608,
			0.00994335100177495,
			0.012908606169822902,
			0.010298649251403313,
			0.003816684112995466,
			-0.00348067806403019,
			-0.008557172161681335,
			-0.009608632396950771,
			-0.006649326836701509,
			-0.0013073464277146174,
			0.004004024888678994,
			0.0071407896959270085,
			0.0070622101801412545,
			0.004133247336931644,
			-0.00016979882374028044,
			-0.00396945510709691,
			-0.005784183213696061,
			-0.005083213814497091,
			-0.0024017684965412264,
			0.0009834739417648997,
			0.0036424372986039993,
			0.004580083649141936,
			0.0035898853367056904,
			0.0012559883020625854,
			-0.00135969080293838,
			-0.003187701643215724,
			-0.003575536194224602,
			-0.002503516126880322,
			-0.0005294306298690314,
			0.0014872096489652376,
			0.0027499128054293363,
			0.0028277613883314974,
			0.0017859787871056622,
			0.00011357732553534381,
			-0.0015017106353334075,
			-0.0024492254183795556,
			-0.0024134283298567795,
			-0.0014609367040901548,
			0.000018970679160808765,
			0.0014804833709829318,
			0.00242138962866836,
			0.002543137384860401,
			0.0018221416962137336,
			0.0004821972926674136,
			-0.0011066687495094368,
			-0.0025573008067263044,
			-0.0035762101728860417,
			-0.004025045645232759,
			-0.003924344675899259,
			-0.003411584910561322,
			-0.00267726234331065,
			-0.001903660659808354,
			-0.0012235616110217874,
			-0.000705026640710802,
			-0.0003584399777348438,
			-0.00015630206030270148,
			-0.00005540231978931542,
			-0.000014247930232761388,
			-0.0000015021121037413662,
		}

		flt := filter.FromCoefficients(coeffs, 0, "Anti-aliasing filter for 4-times oversampling")

		/*
		 * An oversampler / decimator with an oversampling
		 * factor of 4.
		 */
		osd := oversamplerDecimatorStruct{
			factor:             4,
			antiAliasingFilter: flt,
			attenuationFactor:  ATTENUATION_HALF_DECIBEL,
		}

		return &osd
	default:
		return nil
	}

}

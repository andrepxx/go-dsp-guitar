# go-dsp-guitar

This project implements a cross-platform multichannel multi-effects processor for electric guitars and other instruments, based upon concepts and algorithms originating from the field of circuit simulation.

The software takes the signals from N audio input channels, processes them and provides N + 3 audio output channels. The user may, for example, connect the signal from individual instruments to separate input channels of his / her sound card / audio interface. The input signals are then taken and put through dedicated signal chains for processing. The software provides one dedicated signal chain for each input. The output from the last processing element in each chain is then sent to one of the output channels, providing one output channel for each of the input channels. The remaining three output channels include a dedicated metronome, which creates a monophonic click track, as well as a pair of "master output" channels providing a stereo mixdown of all processed signals, say for monitoring purposes.

To manipulate the signal, the user may choose from a variety of highly customizable signal processing units, including the following.

- signal / function generator
- noise gate
- bandpass filter
- auto-wah / envelope follower
- (multi-)octaver
- excess (distortion by phase-modulation)
- fuzz (asymmetric hard or soft saturation)
- overdrive (symmetric soft saturation)
- distortion (symmetric hard saturation)
- tone stack (four-band equalizer)
- (multi-)chorus
- flanger (simple LFO-driven comb filter)
- phaser (complex LFO-driven comb filter)
- tremolo (amplitude modulation)
- ring modulator
- delay (echo)
- power amplifier and speaker simulation

In addition, the software provides ...

- a means to dynamically control the latency of the audio hardware / JACK server
- a highly sensitive, fully chromatic instrument tuner based on the auto-correlation function
- a room simulation (spatializer) to create a stereo mixdown from all (processed) instrument signals
- a metronome to generate a click track for the performing musician for synchronization
- sampled peak programme meters (SPPMs) for controlling the signal level of each input and output channel

... and much more.

The software itself runs in headless mode and is entirely controlled via a modern, web-based user interface, accessible either from the same machine or remotely over the network. It may operate either in real-time mode (default), where it takes signals from either the computer's audio hardware or other applications (e. g. a software synth) and delivers signals to either the computer's audio hardware or other applications (e. g. a DAW), via JACK, or in batch processing mode, where it reads signals from and writes generated output to audio files in either RIFF WAVE or RF64 format. It currently supports files in 8-bit, 16-bit, 24-bit and 32-bit linear PCM (LPCM), as well as 32-bit and 64-bit IEEE 754 floating-point format. Supported sample rates include 22.05 kHz, 32 kHz, 44.1 kHz, 48 kHz, 88.2 kHz, 96 kHz and 192 kHz. The simulation engine will adjust its internal time discretization to the selected sample rate. It will also use the highest precision available from the processor's floating-point implementation for all intermediate results. Only when the results are written to file or handed back to the JACK audio server, the (amplitude) resolution of the audio signal may be reduced, if required.

## Screenshots

![Screenshot 01](/doc/img/screenshot-01-thumb.png)

[View full resolution image](/doc/img/screenshot-01.png)

## Running the software (from a binary package)

Just download the binary (non-`src`) tarball from our *Releases* page, extract it somewhere, start JACK (e. g. via `qjackctl`), `cd` into the directory where you extracted *go-dsp-guitar* and run the `./dsp-*` executable which matches your target architecture. For example, on an x86-64 system running Linux, you may do the following.

```
cd go-dsp-guitar/
./dsp-linux-amd64
```

If you want to run the software in batch processing mode (without JACK) instead, replace the last line with the following.

```
./dsp-linux-amd64 -channels 1
```

Replace the number `1` with the actual number of input channels you want to process, then enter the sample rate (time discretization) you want the simulation engine to operate at.

No matter if you run the software in real-time (JACK-aware) or batch processing mode, you should finally get the following message in your terminal emulator / console.

```
Web interface ready: https://localhost:8443/
```

Point your browser to the following URL to fire up the web interface: https://localhost:8443/

You will find more documentation inside the web interface.

## Building the software from source locally

To download and build the software from source for your system, run the following commands in a shell (assuming that `~/go` is your `$GOPATH`).

```
cd ~/go/src/
go get -v github.com/andrepxx/go-dsp-guitar
cd github.com/andrepxx/go-dsp-guitar/
make keys
make
```

This will create an RSA key pair for the TLS connection between the user-interface and the actual signal processing service (`make keys`) and then build the software for your system (`make`). The executable is called `dsp`, but you may re-name it to match your architecture. For example, on an x86-64 system running on Linux, you may rename the executable as follows.

```
mv dsp dsp-linux-amd64
```

## Building the software from source for other architectures (cross-compilation)

In addition, you may cross-compile the software from source for other architectures. Currently, the following targets are supported for cross-compilation.

```
make dsp-linux-aarch64
make dsp-linux-amd64
make dsp-linux-arm
make dsp-win-amd64.exe
make dsp-win-i686.exe
```

In order to cross-compile the software, you will need a cross-compilation toolchain and a populated `sysroot` for your target architecture. You may find it by invoking your cross-compiler with the `-v` option. For example, the `sysroot` may be one of the following.

```
/usr/aarch64-linux-gnu/sys-root/
/usr/arm-linux-gnu/sys-root/
/usr/x86_64-linux-gnu/sys-root/
/usr/x86_64-w64-mingw32/sys-root/
/usr/i686-w64-mingw32/sys-root/
```

## Packaging the software for distribution

After you build either a binary for your system or cross-compiled binaries for different systems (or both), you can bundle your binaries, along with scripts and auxiliary data, into packages for distribution.

```
make dist
```

This will create a binary package under `dist/bin/go-dsp-guitar-vX.X.X.tar.gz`, as well as a source package under `dist/src/go-dsp-guitar-src-vX.X.X.tar.gz`. Rename these for proper semantic versioning.

## Other build targets

There are other build targets in the `Makefile`.

- `make clean`: Removes the `dist/` directory and the `dsp` executable built for your local system.
- `make clean-all`: Removes the `dist/` directory, as well as all `dsp` executables built for your local system and cross-compiled for other systems.
- `make fmt`: Format the source code. Run this build target immediately before committing source code to version control.
- `make test`: Run automated tests to ensure the software functions correctly on your system. You should also run this before committing source code to version control to ensure that there are no regressions.

## Build requirements

You may need the following packages in order to build the software on your system.

- `gcc-aarch64-linux-gnu`
- `gcc-arm-linux-gnu`
- `gcc-x86_64-linux`
- `git`
- `glibc-arm-linux-gnu`
- `glibc-devel.i686`
- `glibc-devel.x86_64`
- `glibc-headers.i686`
- `glibc-headers.x86_64`
- `golang-bin` (Fedora / RHEL)
- `golang-go` (Debian / Ubuntu)
- `jack-audio-connection-kit` (Fedora / RHEL)
- `jack-audio-connection-kit-devel` (Fedora / RHEL)
- `libjack-jackd2-dev` (Debian / Ubuntu)
- `mingw32-gcc`
- `mingw32-gcc-c++`
- `mingw32-pkg-config`
- `mingw64-gcc`
- `mingw64-gcc-c++`
- `mingw64-pkg-config`
- `openssl`
- `rsync`

## Q and A

**Q: To the project initiator: Why the hell did you create this software?**

**A (short):** I created this software because I wanted to do things that I was unable to do without it.

**A (long):** I'm a fanatic music lover and concert-goer. Learning to play electric guitar was one of my greatest dreams back in the day when I was like thirteen years old and has remained it ever since. I always wanted to establish a band, stand on a stage, play my instrument in front of a large audience, you get the idea. I've seen hundreds of bands live, and I'm very passionate about audio engineering in general since back in the day when I went to school. However, for a long time, electric guitar required a lot of expensive and large and heavy and especially LOUD equipment, stuff that one simply could not justify to operate inside the parents' home or later inside a small flat in the city. While this became less of an issue since decently-sounding (mostly solid-state) pre-amplifier pedals (and rack units) are available, I always felt the desire to create my own audio equipment, which would sound and behave exactly like I wanted. Being a computer scientist, not an electrical engineer, it was clear that, instead of designing and building an actual circuit for, say, a guitar effects pedal, I had to create a piece of software that would behave LIKE such a circuit. While we currently reach a point where we begin to see some decent computer-based audio effects on the horizon, they simply weren't there back in the day when I started creating this software. Also, most of these effects are still not designed with precision as a priority - at least not the kind of precision you'd expect if you have experience with (scientific) simulation software, like I did. They therefore lack the ultimate "realism", so to speak. Also, these effects are normally not THAT customizable that you could ever hope to recreate, say, the sound of a very specific and rare pedal, which was one of the things I always wanted to be able to do, but until now simply couldn't.

**Q: How long did it take you to build this thing?**

**A (short):** Overall it took about five years from the first concept to the first release. Our actual implementation in Golang took about three years until the first release.

**A (long):** The project started in October 2013 as a small tool, written in Python, which could perform some simple processing (especially non-linear distortion and filtering) on RIFF WAVE files. At this point, our vision was born, to take concepts and algorithms originating from the field of circuit simulation and create an effects processor based upon them. We weren't certain if we could ever hope to achieve sufficient performance for real-time processing with this approach, but if we couldn't, at least we could fall back to file-based processing. We first tried to combine Python (for control) with C (for the actual processing) in order to achieve the performance necessary to perform any kind of real-time processing. We tried several audio APIs and libraries, first "raw" ALSA, then PortAudio, and (very briefly) also rtAudio, but with each of these APIs, there was always a point where we hit a roadblock, so development began to stagnate. Almost two years later, in summer 2015, we found that Golang was finally mature enough and we might give this language a chance and try to implement our algorithms in it instead. In addition, we were also at some sort of tipping point, where we could hope that hardware, which is fast enough to run our algorithms in real-time at an acceptable latency, would begin to become commonplace. We also chose to transition to JACK as our audio API. All of these changes finally gave us the required performance and, after countless nights of coding, countless cups of coffee, a lot of (original) research including measurements on actual audio equipment and analysis, as well as countless chords strummed on the guitar for testing, we finally arrived at what we released in December 2018 as *go-dsp-guitar*. So all in all, it took us slightly more than five years from the first prototype to what we finally considered stable and "polished" enough to be released.

**Q: Which platforms can I run this software on?**

**A:** We currently provide binaries for Linux on x86-64 / amd64 (typically PCs), Linux on ARM (typically embedded devices, like a Raspberry Pi), Windows on x86-64 / amd64 (64-bit CPUs, basically all current machines) and Windows on i686 (32-bit CPUs, PCs from pre-2004 or very ressource-limited devices like netbooks from pre-2010). Note that even though the 32-bit variant of the software will typically run on a 64-bit CPU (but not the other way round), using the native (64-bit) variant on a 64-bit machine will be both faster and able to process larger files due to the larger address space of the process. We highly recommend Linux on x86-64 / amd64 for real-time use with JACK. Based upon our own testing, current ARM-based devices will not nearly be fast enough for real-time processing, and Linux currently performs better on x86-64 / amd64 for real-time use than Windows does. We still chose to support Windows as well due to its high market share on the desktop. When running this software on Windows, use the x86-64 / amd64 binary if possible. (It should run, unless your're on a very old machine.) And, of course, you can always use the file-based batch processing mode on slower machines.

**Q: Why do I run out of memory when batch-processing files?**

**A:** The batch processing mode currently requires a lot of (virtual) memory, especially when processing large files and / or many channels, since it has to load the entire files into memory, resample all audio material to the target sampling rate and, finally, extend (zero-pad) it to the same length. During this entire process, the entire high-resolution audio data has to reside in (virtual) memory. This is a technical limitation of how *go-dsp-guitar* currently operates when in this particular mode. If you want to process large files, but are unable to provide a lot of (virtual) memory, you might consider running *go-dsp-guitar* in real-time mode with a very high latency setting instead, then feed audio streams from a DAW through *go-dsp-guitar* and back into the DAW, where the result gets recorded. Real-time operation of *go-dsp-guitar* requires a certain amount of processing power from the CPU though.

**Q: Why don't you support macOS?**

**A:** We're well aware of the fact that macOS has a high market share among the creative folks. However, we currently neither have a device for building and testing nor do we know, which changes we'd have to make to our software so that it builds for macOS. Feel free to fork our project and try to port it to macOS though. When you're done, submit a pull request and we might merge your changes into mainline. (We still won't be able to provide binaries though.) Keep in mind that we will only accept changes which do not break functionality on our currently supported platforms.


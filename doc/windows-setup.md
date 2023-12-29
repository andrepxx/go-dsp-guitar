# Setting up go-dsp-guitar on Windows.

*go-dsp-guitar* uses *JACK Audio Connection Kit* as the way to access the system's audio devices. *JACK* a so called "sound server" and related API that comes originally from the POSIX / Unix world, but has been ported to Windows as well. It's the way Unix applications access the sound card for latency-sensitive operation, so you can think of it as the Unix equivalent to *ASIO*. However, it can actually do more, since it cannot only connect applications to audio devices, but also applications to one another. For example, you could forward the signal from a software synthesizer through an effects processor (like *go-dsp-guitar*) and then through your digital audio workstation (DAW), like *Ardour*, *Cubase*, *Presonus*, or whatever.

### Setting up JACK and running go-dsp-guitar in real-time mode

**First make sure that you have an ASIO-capable driver installed for the audio interface you want to use!**

You can download *JACK for Windows* from here.

https://jackaudio.org/downloads/

Use the 64-bit installer. (Basically all systems have a 64-bit processor now.)

After you installed *JACK*, open a program called *JACK Control* to configure the sound server.

Set the following options.

1. In the "Parameters" tab:

- Driver: portaudio
- Realtime: yes
- Interface: *[the interface you want to use]* (*)
- Sample rate: *[sample rate you want to use]*
- Frames / period: 4096 (**)

(*) Select the audio interface you want to use for real-time processing (for input and output) here. The interface selected here should use the ASIO driver.

(**) This will give the most stable operation. You can later adjust the frames / period setting from the *go-dsp-guitar* user interface while it is running.

2. In the "Advanced" tab:

- Input device: (default)
- Output device: (default)
- Max. Port: 128

If you did not find your interface (with ASIO driver) in the "Interface" list of the "Parameters" tab, instead set the following.

1. In the "Parameters" tab:

- Interface: (default)

2. In the "Advanced" tab:

- Output device: *[Your ASIO driver]*
- Input device: *[Your ASIO driver]*

Set all other options as outlined above.

Apply settings by clicking "Ok", then hit "Start" to start the *JACK* audio server. (*JACK* then runs as a background service on your system.)

After you did that, you can start *go-dsp-guitar* in real-time mode. (It's best to use a console window, for example *PowerShell* for that. Type "./dsp-win-amd64.exe" in *PowerShell* to start *go-dsp-guitar* in real-time mode.)

If the Windows firewall asks you, whether you want to grant *go-dsp-guitar* access to the network, you can generally deny it, if you only want to control the effects processor from the local machine, since local socket connections always appear to be allowed. If you let *go-dsp-guitar* access your network, you can also remote-control it from other computers in your network, for example for headless operation.

Even now, *go-dsp-guitar* will probably not have access to your audio hardware yet, because you still have to configure how to map the actual inputs and outputs that your audio hardware provides to the input and output channels of *go-dsp-guitar*. You can do that in the "Connections" window of the *JACK Control* application, where *go-dsp-guitar* should show up as a client when it's running. If you don't want to do this manually all the time, you can also configure these connections in the "Connections" section inside the "config/config.json" configuration file in the location where you extracted *go-dsp-guitar*.

To find what to put in that section, and how *JACK* and *go-dsp-guitar* are set up on Linux in general, how the connection setup works, etc., watch this video: https://www.youtube.com/watch?v=fxgwbZSU4_g

For specific setup of JACK and *go-dsp-guitar* on Windows, watch this video: 

### Alternatively: Running go-dsp-guitar in batch processing mode

Alternatively, if you want to process **audio files** instead of processing **real-time audio**, you can also run *go-dsp-guitar* in "batch processing mode" instead. To do this, type "./dsp-win-amd64.exe -channels 1" on the console. Replace the number *1* by the actual number of channels you want to process, then type in the sample rate (time discretization) you want the simulation to operate at. (This will be the sample rate of the output files *go-dsp-guitar* generates.) The sample rate is in *Hertz*, so type *96000* for *96 kHz*. After that, the web interface will be accessible via your browser, where you can configure the signal chain(s). (You will have to bypass a certificate warning in your browser to access it.)

After you are finished configuring, click on the "Process now" button in the web interface. The web interface will "freeze" claiming to be "Synchronizing ...". Return to the console window and answer the questions regarding sample format, bit depth, as well as both input and output paths. You can specify relative paths (relative to the current working directory in the shell / console) or absolute paths for input and output files. **Beware that paths you specify as output files will get overwritten without further questions.** Input files have to be in either *RIFF WAVE* or *RF64* format and use either 8 bit, 16 bit, 24 bit or 32 bit linear PCM ("lpcm") or, alternatively, 32 bit or 64 bit IEEE 754 floating-point ("float") format.

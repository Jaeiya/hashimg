# Hashimg

<p align="center">
<i>Hashimg removes duplicate images and renames them for faster future scans.</i>
</p>
<p align="center">
   <img src="https://github.com/Jaeiya/hashimg/blob/1c5b3435dfede011f2f28e0f5d3d2907e1928e8d/demo/hashimg_demo.gif" alt="demo">
</p>

## About

A quaint little hobby utility that I created for the hoard of random images I have in folders. It
reads all images within a folder (not sub-folders) and compares them to one another. It then deletes
the duplicates and renames the remaining ones to their hash.

The hash I'm using to compare the files is only 32 out of 64 characters. I didn't want the names to be
64 characters long as that's a bit unruly, but I also wanted a way to cache the hash in a convenient way.
Now, even though I've truncated the hash to 32 characters, the probability of an accidental collision on 1
billion images is `1 in 2^64` or `1 in 18.4 Quintillion`.

Those odds are what we can consider a "virtual impossibility" and therefore I'm not worried about
finding false positives; it's good enough for my use case and probably most peoples use cases.

### Build Instructions

You can download the source and build this for either Windows or Linux, using their respective build
scripts `build.bat` (for Windows) or `build.sh` (for Linux). They will both build a `hashimg.exe`
to the `/bin` folder.

If you want `hashimg.exe` to be accessible from any directory in your terminal, then you'll have to
move the program to a path that's already part of your `PATH` environment variable, or add a new
path to it that points to the location of `hashimg.exe`.

### Expectations

I work on this as I'm inspired to do so. I would not expect any regular updates but I do plan to add a
better UI using the following frameworks:

[Lip Gloss](https://github.com/charmbracelet/lipgloss) & [Bubble Tea](https://github.com/charmbracelet/bubbletea)

### Issues

If you notice any bugs, feel free to create an issue. I do use this on my own images, so I won't shy
away from critiques or suggestions that might make it better.

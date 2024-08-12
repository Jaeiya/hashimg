# Hashimg

<p align="center">
<i>Hashimg removes duplicate images and renames them for faster future scans.</i>
</p>
<p align="center">
   <img src="https://github.com/Jaeiya/hashimg/blob/1c5b3435dfede011f2f28e0f5d3d2907e1928e8d/demo/hashimg_demo.gif" alt="demo">
</p>

## About

A quaint little hobby utility that I created for the hoard of random images I have in folders. It
reads all images within a folder (**not** including sub-folders) and compares them to one another.
It then deletes the duplicates and renames the remaining ones to their hash.

The hash I'm using to compare the files is only 32 out of 64 characters. I didn't want the names to be
64 characters long as that's a bit unruly, but I also wanted a way to cache the hash in a convenient way.
Now, even though I've truncated the hash to 32 characters, the probability of an accidental collision on 1
billion images is `1 in 2^64` or `1 in 18.4 Quintillion`.

Those odds are what we can consider a "virtual impossibility" and therefore there should be no issue
running this program for your personal images. If you're a company storing significantly more than
1 billion images, then probably safe to assume this app is not for that use-case 😜

### Installing

You can download this app in multiple flavors (platforms) from the [Releases] page. It supports Windows,
Linux, and MacOS. If you're using a modern processor, you can use the `v2` and `v3` versions. You may
or may not get better performance when using those, but ultimately it will be negligible.

Once you download the version for your Operating system, you can either run it in whatever folder
you've extracted it in, but if you want it to be more useful, you'll probably want to set it up in your
`PATH` environment variable, which is talked about a little more below.

### Build Instructions

You can download the source and build this for either Windows or Linux, using their respective build
scripts `build.bat` (for Windows) or `build.sh` (for Linux). They will both build a `hashimg.exe`
to the `/bin` folder.

If you want `hashimg.exe` to be accessible from any directory in your terminal, then you'll have to
move the program to a path that's already part of your `PATH` environment variable, or add a new
path to it that points to the location of `hashimg.exe`. There are many tutorials on how to do this
so you should be able to quickly search for it and find what you're looking for.

### Expectations

As of right now, this program is feature complete. It has all the bells and whistles that I wanted
to add to it, and seems to be quite quick for what it does. I don't plan on adding any more features
but I'm always open to performance-critical tweaks.

### Issues

If you notice any bugs, feel free to create an issue. I do use this on my own images, so I won't shy
away from critiques or suggestions that might make it better.

### Shout-out

I'd like to thank the creators of [Lip Gloss] and [Bubble Tea] for making an incredibly easy framework
for creating useful TUIs. I didn't want to have to build all of it out myself, so thanks to the
[Charm] team, I didn't have to!

[Releases]: https://github.com/Jaeiya/hashimg/releases
[Lip Gloss]: https://github.com/charmbracelet/lipgloss
[Bubble Tea]: https://github.com/charmbracelet/bubbletea
[Charm]: https://charm.sh

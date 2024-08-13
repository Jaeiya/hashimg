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

Those odds are what we call a "virtual impossibility." Unless you're working with significantly more
than 1 billion images, you won't have to worry about collisions. In most cases, you'll probably have
folders named based on date, theme, or event...in which case if you ever **were** to accumulate
significantly more than a billion images, they likely wouldn't all be in a single folder.

### Accuracy

As long as an image file has exactly the same dimensions and data as another image, it will be found
to be a duplicate. This means that images which have the same _visual_ appearance, can't necessarily
be distinguished by the program. If you have two images that are _visually_ identical, but one is
larger than the other, there's no way for the program to know that they're identical because their
**data** is not identical.

So while it may be virtually impossible to get false positives (flagging images that are dupes, but are
not), it will **literally** be impossible to find duplicates among only the _visually_ identical, if
the data isn't also identical. If most of your images are duplicates simply because some are larger
than others, then this program is not the right tool; however if you have a lot of exact copies of
images in a folder, then this program is perfect.

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

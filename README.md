# Hash Image

A quaint little hobby utility that I created for the hoard of random images I have in folders. It
reads all images within a folder (not sub-folders) and compares them to one another. It then deletes
the duplicates and renames the remaining ones to their hash.

The hash I'm using to compare the files is only 32 out of 64 characters. I didn't want the names to be
64 characters long as that's a bit unruly, but I also wanted a way to cache the hash in a convenient way.
Now, even though I've truncated the hash to 32 characters, the probability of an accidental collision on 1
billion images is `1 in 2^64` or `1 in 18.4 Quintillion`.

Those odds are what we can consider a "virtual impossibility" and therefore I'm not worried about
finding false positives; it's good enough for my use case and probably most peoples use cases.

### Expectations

I work on this as I'm inspired to do so. I would not expect any regular updates but I do plan to add a
better UI using the following frameworks:

[Lip Gloss](https://github.com/charmbracelet/lipgloss) & [Bubble Tea](https://github.com/charmbracelet/bubbletea)

### Issues

If you notice any bugs, feel free to create an issue. I do use this on my own images, so I won't shy
away from critiques or suggestions that might make it better.

<p align="center">
<b>Hashimg,</b> <i>removes duplicate images and renames them for faster future scans.</i>
</p>

<p align="center">
   <img src="https://github.com/Jaeiya/hashimg/blob/1c5b3435dfede011f2f28e0f5d3d2907e1928e8d/demo/hashimg_demo.gif" alt="demo">
</p>

<p align="center">
   <a href="https://goreportcard.com/report/github.com/jaeiya/hashimg"><img src="https://goreportcard.com/badge/github.com/jaeiya/hashimg"></a>
   <a href="https://github.com/Jaeiya/hashimg/actions"><img src="https://img.shields.io/github/actions/workflow/status/jaeiya/hashimg/release.yml"></a>
   <a href="https://github.com/Jaeiya/hashimg/releases"><img src="https://img.shields.io/github/v/release/jaeiya/hashimg"></a>
   <a href="#"><img src="https://img.shields.io/github/go-mod/go-version/jaeiya/hashimg"></a>
   <a href="https://wakatime.com/@92eac300-9535-4747-a2e0-0cfb5d345c51/projects/csqjoyhxbl"><img src="https://wakatime.com/badge/user/92eac300-9535-4747-a2e0-0cfb5d345c51/project/bb183dcc-4615-42c1-95f8-2395c879c3e3.svg"></a>

</p>

## Hashimg

A quaint little hobby utility that I created for the hoard of random images I have in folders. It
reads all images within a folder (**not** including sub-folders) and compares them to one another.
It then deletes the duplicates and renames the remaining ones to their hash.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Installation](#installation)
  - [Binary Releases](#binary-releases)
  - [Go CLI](#using-go-cli)
- [FAQ](#faq)
  - [Does it find all duplicates?](#will-it-find-all-duplicate-images-no-matter-what)
  - [How likely are false-positives?](#how-likely-are-false-positives)
  - [Will it auto-delete my images?](#will-it-automatically-delete-my-images)
  - [What will my files end up looking like?](#what-will-my-files-look-like-after-its-done)
- [Future Updates](#future-updates)
- [Developer Instructions](#developer-instructions)
  - [Install Prerequisites](#install-prerequisites)
  - [Build Dev](#build-development-binaries)
  - [Build Snapshot](#build-snapshot-of-production-archives)
- [Feedback](#feedback)
- [Shout-Out](#shout-out)

## Installation

### Binary Releases

For Linux, MacOS (10.15+ Catalina), and Windows, you can download them from the [Releases Page](https://github.com/Jaeiya/hashimg/releases)

### Using Go CLI

```bash
go install github.com/Jaeiya/hashimg@latest
```

## FAQ

### Will it find all duplicate images no matter what?

No. The file data must be identical. Just because images _appear_ to be identical, does not mean
that they are. If those images have different resolutions or one is compressed more than another,
they will not be flagged as duplicates. An image is **only** considered a duplicate if it has
the **exact** same data as another image, including meta-data.

### How likely are false-positives?

Virtually impossible. In order for there to even be a reasonable possibility of false-positives,
you would need to have quadrillions of images. Those images would also have to all be in
a single folder. So the likelihood of an accidental deletion of a novel image, is virtually zero.

### Will it automatically delete my images?

Yes and No. There is a review option that will ask you to review the images before deletion takes
place. During this period, you can choose to keep the images that were detected as duplicates,
or to delete them. If you choose to keep them, the program aborts and your files are completely
untouched.

If you choose **not** to review the duplicate images, then all duplicates are automatically deleted
and your existing images will be renamed to their hash. Renaming the images to their hash allows
for faster future scans.

### What will my files look like after it's done?

```
=== Before ===
file1.png
file2.bmp
file3.webp

=== After ===
0x@c147efcfc2d7ea666a9e4f5187b115c9.png
0x@3377870dfeaaa7adf79a374d2702a3fd.bmp
0x@6f3fef6dc51c7996a74992b70d0c35f3.webp
```

The `0x@` is a unique identifier so that my program knows the file name is part of the calculated
hash of the file.

## Future Updates

The program at this point **is** feature complete. The review feature was actually something a friend
had mentioned, so it wasn't even planned, but I added it anyway. The only future feature that makes
sense at this point, is "similar image" detection. Right now only identical images are detected,
but obviously this is inferior to comparing them based on their _visual_ similarity.

I might look into [pHash] support, which is a hashing technique known as "perceptual hashing"
which quantifies the _features_ of an image to generate a hash that can be compared for similarities.
If you're interested in the more technical details, you can read about it [here](https://www.hackerfactor.com/blog/index.php?/archives/432-Looks-Like-It.html)

That all being said, I don't plan to work on this project for awhile unless I discover bugs or better
ways to do what the program is already doing.

[pHash]: https://github.com/corona10/goimagehash

## Developer Instructions

You'll need to have Go `1.22.x` or higher installed. If you're using `1.23.x` or higher - as of
`2024-08-25` - you will end up with significantly larger binaries. Not sure if that's a feature
or a bug but...that's how it is.

All of the following builds are technically production builds, in terms of how the binaries are
optimized. Even if you build using the `dev` target, you're still getting the most optimized
build possible. The cold start time can take a little bit, but subsequent builds are typically
1s or less.

The output of all builds is the `./dist` dir

### Install Prerequisites

If you don't have Go or Goreleaser installed, then the following make targets will attempt to
install them. The linux version is only partially tested and the MacOS isn't tested at all
since I don't have an Apple device.

If it fails for any reason and you do in fact have Go installed, then you can just execute
`make install-goreleaser`. Because it's using the Go CLI, there should be no issues.

```bash
make setup-windows
```

```bash
make setup-linux
```

```bash
make setup-macos
```

### Build development binaries

This will build 3 binaries to the `/dist` dir, one for each platform (linux, darwin, & windows),
all of which are `x86_64` compatible.

```bash
make dev
```

### Build snapshot of production archives

This is a preview of how the builds will look in production, however they do use a snapshot
version, which will not be used in production. They also build to the `/dist` dir.

```bash
make snapshot
```

## Feedback

If you notice any bugs, feel free to create an issue. I do use this on my own images, so I won't shy
away from critiques or suggestions that might make it better. It won't be my main focus, but I am
active enough that I will definitely respond.

## Shout-out

I'd like to thank the creators of [Lip Gloss] and [Bubble Tea] for making an incredibly easy framework
for creating useful TUIs. I didn't want to have to build all of it out myself, so thanks to the
[Charm] team, I didn't have to!

[Releases]: https://github.com/Jaeiya/hashimg/releases
[Lip Gloss]: https://github.com/charmbracelet/lipgloss
[Bubble Tea]: https://github.com/charmbracelet/bubbletea
[Charm]: https://charm.sh

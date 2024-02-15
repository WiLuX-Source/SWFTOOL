<h1 align="center">
SWFTOOL
</h1>

## Description

CLI tool that provides a set of commands for working with SWF (Shockwave Flash) files. It allows you to download the latest projector for your operating system, merge a projector with a movie, extract a movie from a projector. Supports Linux & Windows

*Linux merge **not works** with 64-bit projectors atm. Next release will fix it.*

## How To Use

1. Download the [SWFTOOL](https://github.com/WiLuX-Source/SWFTOOL/releases/download/2.0/swftool.exe).
2. Get file path of your projector & swf.
3. Use the app through cmd or drag on it.

## How To Build

- You can latest Go runtime to build via `go build`

## Command List

Command | Parameters | Description
--- | --- | ---
extract | movie (path) | Extract a movie from projector.
merge | projector (path), movie (path) | Merge projector with a movie.
download | none | Downloads the latest & compatible projector with your OS.
help | none | Displays the version & commands you have access to.

## NOTICE

- SWFTOOL was built top on [Magicswf](https://github.com/PopovEvgeniy/magicswf) & [Swfknife](https://github.com/PopovEvgeniy/swfknife) in 1.0
- Version 1.0 Code belongs to [PopovEvgeniy](https://github.com/PopovEvgeniy).
- This program made for ease of use.
- Special Thanks to [JrMasterModelBuilder](https://github.com/JrMasterModelBuilder) for his [Shockpkg](https://github.com/shockpkg) project.

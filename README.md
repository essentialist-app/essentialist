# Essentialist


Programs for [spaced repetition][1] using flashcards in [Markdown][2].

- **Essentialist**: Application for desktops (Windows, MacOS and Linux) and mobile (Android)
- **Flashdown**: Terminal user interface

The space repetition algorithm used is based on [SM-2][3].

Key features:

- **Privacy**: No network access, your data never leave your device.
- **Easy card creation**: Flash cards are plain text Markdown files.
- **Cross-platform**: Runs on Linux, MacOS, Windows and android.

[1]: https://en.wikipedia.org/wiki/Spaced_repetition
[2]: https://en.wikipedia.org/wiki/Markdown
[3]: https://en.wikipedia.org/wiki/SuperMemo#Description_of_SM-2_algorithm

See the [CONTRIBUTING.md](/.github/CONTRIBUTING.md) for how to report bugs and
submit pull request.

## Flash cards syntax

Each deck of cards is a plain text Markdown files with the extension `.md` (ex:
`sample.md`). You can put all your decks in the same directory.

Each card starts with a heading level 2 (line starting with `##`) defining the
question. The answer is the content following (until the next heading level 2).

You progress is stored in a hidden file `.<deck file>.db` (ex: `.sample.md.db`).

Example of a deck with 3 cards:

```markdown
## Question: what format is used?

Questions and answers are in **Markdown**.

## Are lists supported?

Yes, here is an example:

- one
- **two**
- three

## How to include a table in the answer?

Answer with a table.

|  A  |  B  |
| --- | --- |
| 124 | 456 |
```

## Essentialist (GUI)

A GUI version for desktops and mobile (Android, iOS support isn't tested).

![Screenshot](docs/essentialist-screenshot.png)

### Installation

[Download the latest version](https://github.com/essentialist-app/essentialist/releases)
or compile it with the following instructions:

<details><summary>Linux</summary>
<p>

In order to build Essentialist, install the following dependencies
(Debian/Ubuntu):

```
sudo apt-get install libxrandr-dev libxcursor-dev libglx-dev libxi-dev libgl-dev libxxf86vm-dev
```

The simplest way to build and install Essentialist with its desktop launcher
and icon on Linux is to use the Makefile like:

```
make install DESTDIR=$HOME/.local PREFIX=""
```

Essentialist [does not support](https://github.com/fyne-io/fyne/issues/5471)
Xorg and Wayland at the same time. By default, Xorg is enabled. If you want to
enable Wayland support, run:

```shell
make install DESTDIR=$HOME/.local PREFIX="" WAYLAND=true
```

You can also simply build the binary with:

```
go build ./cmd/essentialist
```

</p>
</details>

<details><summary>MacOS</summary>
<p>

```shell
go install fyne.io/tools/cmd/fyne@latest
cd cmd/essentialist
fyne package -os darwin
```

</p>
</details>

<details><summary>Windows</summary>
<p>

```shell
go build -x -o essentialist.exe ./cmd/essentialist
```

</p>
</details>

<details><summary>Android</summary>
<p>

1. Install the Android NDK from <https://developer.android.com/ndk/downloads>.
   Set the `ANDROID_NDK_HOME` variable to the directory where the NDK is located.

1. Build the Android APK with:

  ```shell
go install fyne.io/tools/cmd/fyne@latest
cd cmd/essentialist
fyne package -os android
  ```

1. Plug your phone over USB and install the APK with:

  ```shell
adb install Essentialist.apk
  ```

Use the local storage (of your Android device) to import flash cards. For
example, you can put them in an SD card and import them from the Essentialist
application.

</p>
</details>

## Flashdown (terminal version)

Flashdown is the terminal application.

To install it, clone this repo and run:

```shell
go install ./cmd/flashdown
```

Usage:

```shell
flashdown <deck_file> [<deck_file>]
```

![Screenshot](docs/flashdown-screenshot.png)

Similar project: <https://github.com/Yvee1/hascard>.

## Maintenance

### How to update dependencies

```shell
go get -u ./...
go mod tidy
go run github.com/dennwc/flatpak-go-mod@latest -out cmd/essentialist/flatpak/ .
go-licenses report ./... --template cmd/essentialist/licenses.tpl > cmd/essentialist/licenses.md
```

### How to make a new version

A new version can be created from the [Release workflow](https://github.com/essentialist-app/essentialist/actions/workflows/release.yml).

1. Go to the "Actions" tab in the GitHub repository.
2. Select the "Create Release" workflow.
3. Click on "Run workflow".
4. Enter the new version number (e.g., 1.2.3) in the "Version" input field.
5. Click on "Run workflow".

The workflow will automatically update the version number in the files, commit the changes, build the application for all platforms, and create a new Git tag. This will trigger the release creation workflow, which will create a new release on GitHub with the build artifacts.

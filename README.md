# AppFuk

Make macOS application bundles deployable.

For some reason Apple does not provide any tool to do this. Therefore there are plenty of tools out there to help with that. This is one of them.

## Usage
Since your application's main executable in your bundle may be a wrapper, it is often not enough to read the executable entry from the `Info.plist`. Also, you may have additional helper binaries located in your bundle. This tool expects you to point it directly to the executable inside your bundle.

```shell
$ appfuk ~/MyApp.app/Contents/MacOS/myapp
```

This copies non-system linked libraries and all its dependencies into your bundle's `Frameworks` directory and rewrites the link locations. Make sure your executables and dependencies are compiled with the `-headerpad_max_install_names` option.

Note that this process invalidades existing code signatures.

## Example

```shell
$ appfuk ezQuake.app/Contents/MacOS/ezquake-darwin-arm
/Users/fzwoch/code/ezquake-source/ezQuake.app/Contents/MacOS/ezquake-darwin-arm:
  [copy] libSDL2-2.0.0.dylib
  [copy] libpcre.1.dylib
  [copy] libpng16.16.dylib
  [copy] libjpeg.8.dylib
  [copy] libjansson.4.dylib
  [copy] libminizip.1.dylib
  [copy] libspeex.1.dylib
  [copy] libspeexdsp.1.dylib
  [copy] libfreetype.6.dylib
  [copy] libsndfile.1.dylib
  /opt/homebrew/opt/freetype/lib/libfreetype.6.dylib:
    [skip] libpng16.16.dylib
  /opt/homebrew/opt/libsndfile/lib/libsndfile.1.dylib:
    [copy] libogg.0.dylib
    [copy] libvorbisenc.2.dylib
    [copy] libFLAC.12.dylib
    [copy] libopus.0.dylib
    [copy] libmpg123.0.dylib
    [copy] libmp3lame.0.dylib
    [copy] libvorbis.0.dylib
    /opt/homebrew/opt/libvorbis/lib/libvorbisenc.2.dylib:
      [skip] libvorbis.0.dylib
      [skip] libogg.0.dylib
    /opt/homebrew/opt/flac/lib/libFLAC.12.dylib:
      [skip] libogg.0.dylib
    /opt/homebrew/opt/libvorbis/lib/libvorbis.0.dylib:
      [skip] libogg.0.dylib
```

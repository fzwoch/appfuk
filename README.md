# appfuk

Make macOS application bundles deployable.

For some reason Apple does not provide any tool to do this. Therefore there are plenty of tools out there to help with this. This is one of them.

## Usage
Since your application's main executable in your bundle may be a wrapper it is often not enough to read the executable entry from the `Info.plist`. Also, you may have additional helper binaries located in your bundle. This tool expects you to point it directly at the executable inside of your bundle.

```shell
$ appfuk ~/MyApp.app/Contents/MacOS/myapp
```

This copies non-system linked libraries and all its dependencies into your bundle's `Frameworks` directory and rewrites the link locations. Make sure your executable and dependencies are compiled with the `-headerpad_max_install_names` option.

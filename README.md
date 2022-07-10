# appheist

Download apk files for static analysis


## Build & Launch

```
go build appheist.go
mkdir files
./appheist -help

2022/07/10 00:12:45 AppHeist Downloader by wunderwuzzi23
Usage of ./appheist:
  -app string
        Which app?
  -developer string
        Developer account
  -mode string
        buildindex, download, listapps, listapps+ (default "buildindex")
  -pagestart int
        Specify the page to start enumerating (default 1)
  -skipvariants
        Only index first found variant for version
```


## Index apk files
`./appheist -developer "developername" -mode buildindex`

## Download all indexed apks
`./appheist -developer "developername" -mode download`


## Features

Provide a developer account or app name and download the relevant apk files.

By default variants are skipped.

### Flags:

```
  -app string
        Which app?
  -developer string
        Developer account
  -mode string
        buildindex, download, listapps, listapps+ (default "buildindex")
  -pagestart int
        Specify the page to start enumerating (default 1)
  -skipvariants
        Only index first found variant for version
```   

## Credits

`apkmirror.com` is awesome - consider becoming a premium member. I'm not affiliated with it.

# IMDb Simple Movies WebService
[![Build Status](https://travis-ci.org/thiekus/imdbsmws.svg?branch=master)](https://travis-ci.org/thiekus/imdbsmws)

IMDb Simple Movies WebService (imdbsmws) is Simple RESTful API Web Service Server that manages and retriving titles data from imdb.com

## Setup
* Build from sources or [download precompiled](https://github.com/thiekus/imdbsmws/releases) releases.
* For using precompiled binaries, just download version that match with your OS, then extract to someplace.
* Launching server by running ```imdbsmws``` (```imdbws32.exe``` or ```imdbws64.exe``` for precompiled Windows binary).
* After running, visit admin panel at ```http://localhost:33666/admin``` to manage server.
* For first time login, username is ```admin``` and password is ```admin```.
* For first time use, you'll need importing data from IMDb datasets.

## Import Data from IMDb
* To import, visit admin panel, then select ```Import``` menu.
* You can specify datasets, filters at your preference (if you don't have any idea, just leave it :p )
* Press ```Import``` button to run Importing process, please wait until finished.
* After successfull import, you should get movies list in ```Movie List``` menu.

## API Endpoints
```
/movies (Method GET) -> to get Movie List stored in your database.
/movies/{titleId} (Method GET) -> to get individual Movie information.
/movies (Method POST) -> to post new movie entry.
/movies/{titleId} (Method PUT) -> to edit existing movie entry.
/movies/{titleId} (Method DELETE) -> to delete existing movie entry.
```

## Build From Sources
You need [Go Compiler](https://golang.org/dl/) and [Mingw64 GCC](https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/) (or [TDM GCC](http://tdm-gcc.tdragon.net/download) for simpler to setup), as needed by SQLite. Make sure Go and GCC placed in ```PATH``` environment (TDM GCC Installer will do this automatically). Then run ```go build -i -v github.com/thiekus/imdbsmws``` to compile. For compilation in windows, just edit and use ```win32_build.bat``` or ```win64_build.bat``` scrtipt.

## Screenshots
![Screenshoot](https://github.com/thiekus/imdbsmws/raw/master/_screenshots/scshot1.png)
![Screenshoot](https://github.com/thiekus/imdbsmws/raw/master/_screenshots/scshot2.jpg)
![Screenshoot](https://github.com/thiekus/imdbsmws/raw/master/_screenshots/scshot3.png)
![Screenshoot](https://github.com/thiekus/imdbsmws/raw/master/_screenshots/scshot4.png)

## License
This application is Licensed under [MIT License](https://github.com/thiekus/imdbsmws/blob/master/LICENSE).

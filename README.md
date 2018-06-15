goFBPages
======================

A facebook page photo album tool that supports concurrency download. This tool help you to download those photos for your backup, all the photos still own by original creator.

Install
--------------

    go get -u -x github.com/hoanglv00/goCrawlerFacebook

Note, you need go to [faceook developer page](https://developers.facebook.com/tools/explorer?method=GET&path=me) to get latest token and update in environment variables.

     export FBTOKEN = "YOUR_TOKEN_HERE"

Usage
---------------------

    goCrawlerFacebook [options] 

All the photos will download to `USERS/Pictures/FBPages` and it will separate folder by page name and album name.

Options
---------------

- `-n` Facebook page name such as: [Diemmy9x](https://www.facebook.com/Diemmy9x), or input facebook id if you know such as 112743018776863 
- `-s` photos or videos
- `-c` number of workers. (concurrency), default workers is "2"


Examples
---------------

Download all photos from Scottie Pippen facebook pages with 10 workers.

  goFBPages -n=Diemmy9x -s=videos -c=10


TODOs
---------------

- Support specific album download.
- Support videos download.
- Support firend/self album download for backup.
- Support CLI to select which album you want to download.



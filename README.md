# Google Takeout Metadata Fix

When photos and videos are backed up from Google Photos using Google Takeout, the exif metadata properties are often not correct. The `DateTimeOriginal`, `FileCreateDate`, and `FileModifyDate` end up being the timestamp when the files were downloaded, and not the actual timestamp of when the photo/video was created. Google provides us with a JSON file for each photo/video which contains the original metadata (including the correct date/timestamp properties), which we can use to correct the metadata of the downloaded files.

This script intends to provide an easy way to fix all the metadata issues related the Google Takeout output.

This script assumes you use the following process to perform your Google Photos backup:

1. Download your Google Takeout output
2. Unzip it
3. Download your Google Photos albums separately. Copy/paste that folder in your main Google Takeout directory (where you unzipped all the files) and merge the folders. This ensures that photos which were shared by others in your shared albums are actually copied to your album folders.
4. Once this Google Takeout folder is ready, follow the instructions to run the script at the top level of this folder.

> Google Takeout only stores _your_ photos in your album folder and does not have photos shared by others in them. Even if you "Save" them in Google Photos, it only ends up in your "Photos from Year" folder and not your album folders.

## Technical Requirements

1. This script assumes you are a Android/Pixel user as it does some HEIC conversions
2. [Go](https://go.dev/) needs to be installed and be available on the path `go`
3. [exiftool](https://exiftool.org/) needs to be installed and be available on the path `exiftool`
4. This script has only been tested to work on Windows

## Running the script

1. Clone or download this repository
2. Open a Terminal/PowerShell at the root of the repo and run `go run main.go`
3. Enter the path to the folder where your Google takeout is, for example `D:\Google Photos`
4. The script will perform all the metadata updates and display the progress

## What metadata is being updated?

At a top level, this script does the following:

1. Renames all `.TS.mp4` files to just `.mp4`
2. Renames any JSON metadata files which end in `.TS.mp4.json` to end in just `.mp4.json` to match the files updated in step 1
3. Converts all HEIC files which have an associated HEIC.json file to JPG
4. Renames the JSON metadata file associated with the HEIC file in step 3 to `JPG.json` for example `filename.HEIC.json` would become `filename.jpg.json`
5. Updates the date/timestamp metadata of all photo and video files using the JSON metadata files

We do all of the above because:

1. exiftool sometimes has trouble parsing files which have "double" extensions such as .TS.mp4, so we can just convert them to .mp4 without any issues. The `.TS.mp4` files are created by Top Shot or motion photos.
2. Google Photos converts HEIC files to JPG in their backend. However, when we download the files through Google Takeout, it keeps the .HEIC extension. exiftool does not like that, so we need to fix the extension to be JPG. Note that if you manually download Photo Albums from Google Photos (bypassing Googke Takeout) then any HEIC download are true HEIC, so we need to make sure we don't convert these files to JPG, which is why we check for a JSON metadata file first.
3. Google provides us the metadata in the form of JSON files. We can fetch the metadata from these files and use exiftool to override the properties in the image/video files.

## Other useful commands

View the tags in the json sidecar file:

```
exiftool -g1 -a -s 20231125_225835.jpg.json
```

# MassetDL
Mass asset downloader for Roblox assets. Inspired by [Iris's Asset Downloader](https://github.com/IrisV3rm/AssetDownloader) with the appeal of being more efficient. 
# Usage
Open command prompt, `cd` to the directory you wish to save the various assets to, make sure the .exe is in the same folder, have an `assets.txt` file with the asset ID on each line for example:
```
123
456
789
```
Then simply type `massetdl -file`!

If you would like to have the program scrape everything for you (it just starts at ID 1000000 and increments) then type `massetdl -scrape`.
If you would like to scrape for a certain type, add a filter: `massetdl -scrape -filter shirts`.
# Note
Does not require you to specify the asset's type as it will find it on its own.
Executable md5sum: `fed09713a832e6f1f4e7fa8faa1f91af`

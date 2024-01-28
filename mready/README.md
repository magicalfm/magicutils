# mready

magical-ready

ZOOM PodTrack P4で録音した、マルチトラックのディレクトリを指定すると、 `MIC1.WAV` と `MIC2.WAV` をそれぞれホストの名前に修正してから、 `lufs-normalizer` を実行して、LUFS 26で正規化します。


## Usage

```shell
$ ./mready
使用法: ./mready [--ep EP] 音声ファイルが配置されているディレクトリ

$ ./mready --ep 14 '/Volumes/Extreme SSD/ZOOM PodTrak P4/Untitled/P4_Multitrack/2024_0122_2105'
Input file loudness: -26.137191355062452
Output file loudness: -26.000000000000004
Output file: output/ep14_1706430311/upamune_normalized.wav
Input file loudness: -31.242533254382515
Output file loudness: -26.024661156182912
Output file: output/ep14_1706430311/michiru_da_normalized.wav
処理が完了しました。出力ディレクトリ: output/ep14_1706430311 
```

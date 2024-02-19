# StreamMusic LRC API

用于 [「音流」自定义 API](https://aqzscn.cn/archives/stream-music-custom-api)，优先读取同路径下的 LRC/TXT 格式歌词，否则通过 https://github.com/Binaryify/NeteaseCloudMusicApi 匹配网易云音乐中的歌词。

## Usage

``` bash
cp docker-compose.example.yml docker-compose.yml  
docker-compose build && docker-compose up -d --force-recreate
```
# go-pdf

## Fonts

https://github.com/minoryorg/Noto-Sans-CJK-JP

## Color

黒色テキストを使う場合 `rgb(0,0,0)` では正しく判定されないので、 `rgb(1,1,1)` を使う

```json
"color": {
    "r": 1,
    "g": 1,
    "b": 1
}
```
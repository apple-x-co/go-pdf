# go-pdf

json to pdf

## How to use

### basic

```bash
go-pdf --in layout.json --out output.pdf --ttf fonts/TakaoPGothic.ttf

```

### show help

```bash
go-pdf --help
```

---

## Notices

### Fonts

https://github.com/minoryorg/Noto-Sans-CJK-JP

### Color

黒色テキストを使う場合 `rgb(0,0,0)` では正しく判定されないので、 `rgb(1,1,1)` を使う

```json
"color": {
    "r": 1,
    "g": 1,
    "b": 1
}
```
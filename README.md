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

## Restriction

* 一つの `liner_layout` に対して `elements` と `liner_layouts` を含めることはできない。
* PDF生成に利用しているライブラリの関係上、テキスト色を黒色以外から黒色テキストに戻す場合に `rgb(0,0,0)` では正しく判定されないので、黒色テキストは `rgb(1,1,1)` を使う。

### Fonts

https://github.com/minoryorg/Noto-Sans-CJK-JP
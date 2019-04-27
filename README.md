# go-pdf

json to pdf

## Feature

* Liner layout
* Style template
* Auto line break
* Auto page break
* Header / Body / Footer
* Image resizing
* Text break

## Specification

* 一つの `liner_layout` に対して `elements` と `liner_layouts` を含めることはできない。
* PDF生成に利用しているライブラリの関係上、テキスト色を黒色以外から黒色テキストに戻す場合に `rgb(0,0,0)` では正しく判定されないので、黒色テキストは `rgb(1,1,1)` を使う。
* `text.align` を使うには `width` の指定が必要
* `text.valign` を使うには `height` の指定が必要
* `text.wrap` を使うには、 `width` の指定が必要

## Fonts

https://github.com/minoryorg/Noto-Sans-CJK-JP

## How to use

### basic

```bash
go-pdf --in layout.json --out output.pdf --ttf fonts/TakaoPGothic.ttf

```

### show help

```bash
go-pdf --help
```

## Examples

```json
{
  "width": 595,
  "height": 842,
  "pages": [
    {
      "liner_layout": {
        "orientation": "horizontal",
        "elements": [
          {
            "type": "text",
            "attributes": {
              "text": "text"
            }
          },
          {
            "type": "text",
            "attributes": {
              "text": "text",
              "size": {
                "width": 100
              },
              "border": {
                "width": 1,
                "color": {
                  "r": 255,
                  "g": 0,
                  "b": 0
                }
              }
            }
          }
        ]
      }
    }
  ]
}
```

![example1](./example1.png) 

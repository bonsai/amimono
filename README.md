# amimono 🧶

Machine-knitting / 編み図を、**pixel art → pixel grid → stitch pattern → machine IR**へ変換するためのドメイン実装。

## Go API

標準ライブラリ中心のGo APIで、編み図画像をセル単位の `pixel-grid/v1` JSONへ変換できます。

```bash
go run ./cmd/amimono-api
```

### Endpoints

- `GET /healthz` — ヘルスチェック
- `GET /v1/palette` — 現在の7色パレット
- `GET /v1/pattern` — リポジトリ内の35×53サンプルJSON
- `POST /v1/convert?width=35&height=53` — PNG/JPEG画像をセル分割してJSON化

画像変換は `multipart/form-data` の `image` フィールド、または画像バイナリの直接POSTに対応します。

```bash
curl -X POST \
  -F image=@IMG_2325.png \
  'http://localhost:8080/v1/convert?width=35&height=53'
```

中央60%をサンプリングしてグリッド線の影響を抑え、パレット最近傍色から1〜7の色番号へ変換します。

## 今回追加したもの

添付された35×53マスの編み図画像をセル単位で読み取り、`pixel-grid/v1` JSONとして正規化しました。

- `patterns/35x53-stitch-chart.json` — 35目 × 53段のセルデータ
- `index.html` — ブラウザで確認できるインタラクティブ編み図ビューア
- `cmd/amimono-api/main.go` — Go画像セグメントAPI
- `go.mod` — Go module
- `Dockerfile` — distrolessコンテナ

## データモデル

```text
PNG / SVG / JSON
      ↓
pixel_grid
      ↓
stitch_pattern
      ↓
panel_pattern
      ↓
machine_ir
      ↓
machine_output
      ↓
simulation / production_data
```

今回のJSONは、画像をそのまま保存するのではなく、**「1セル = 1つの色番号」**として扱える中間表現です。

## パレット

| 番号 | 用途 |
|---|---|
| 1 | black |
| 2 | white |
| 3 | light-yellow |
| 4 | yellow |
| 5 | red |
| 6 | brown |
| 7 | gold |

> 色名・HEXは元画像の凡例を基準にした表示用定義です。実際の毛糸色はゲージ・糸・機械条件に合わせて別途 yarn mapping します。

## 次のステップ

1. 画像セグメント解析の精度評価・自動トリミング
2. OCRではなくセル背景色＋凡例から色番号を推定
3. JSON Schemaを固定
4. stitch patternへの変換
5. ゲージ補正
6. machine-independent IRへコンパイル
7. シミュレーション結果をevidenceとして保存

`ontology.yaml` の `pixel_art2pixel_grid` → `pixel_grid2stitch_pattern` の流れに接続するMVPです。

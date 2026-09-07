# amimono 🧶

Machine-knitting / 編み図を、**pixel art → pixel grid → stitch pattern → machine IR**へ変換するためのドメイン実装。

## 今回追加したもの

添付された35×53マスの編み図画像をセル単位で読み取り、`pixel-grid/v1` JSONとして正規化しました。

- `patterns/35x53-stitch-chart.json` — 35目 × 53段のセルデータ
- `index.html` — ブラウザで確認できるインタラクティブ編み図ビューア
- 行・列番号表示
- 色番号表示 / 非表示
- 拡大・縮小
- セル選択で「段 × 目」と色番号を確認
- 印刷用レイアウト

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

1. 画像セグメント解析をCLI化
2. OCRではなくセル背景色＋凡例から色番号を推定
3. JSON Schemaを固定
4. stitch patternへの変換
5. ゲージ補正
6. machine-independent IRへコンパイル
7. シミュレーション結果をevidenceとして保存

`ontology.yaml` の `pixel_art2pixel_grid` → `pixel_grid2stitch_pattern` の流れに接続するMVPです。

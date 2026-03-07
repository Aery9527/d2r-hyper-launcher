# CLI Refactor Plan

## 為什麼這是架構訊號

- 最近為了把「玩家輸入錯誤時不要立刻跳回主選單」這個 UX 規則做一致，必須同時修改主選單、區域選單、mod 選單、flag 選單、switcher 選單
- 這代表「CLI 錯誤輸入回饋」其實是跨多個 flow 的共通 concern，但原本沒有單一清楚落點
- `cmd/d2r-hyper-launcher/cli_feedback.go` 雖然成功把錯誤提示樣式收斂成 common helper，但也反過來暴露：menu flow、驗證與導航控制仍然大量散在 `cmd/d2r-hyper-launcher/main.go`

## 問題落點

### `cmd/d2r-hyper-launcher/main.go`

- 主選單 dispatch、區域選擇、mod 選擇、flag 設定、switcher 設定都集中在同一個大檔案
- `launchAccount()` 與 `launchAll()` 有相似的區域 / mod 選擇結構
- `setupAccountLaunchFlags()`、`configureFlagsByFlag()`、`configureFlagsByAccount()` 把 UI 顯示、輸入解析、確認流程與 domain 套用混在一起
- `isMenuNav()` / `printSubMenuNav()` 雖然統一了導航字元，但沒有統一整個 prompt → validate → feedback flow

### `cmd/d2r-hyper-launcher/cli_feedback.go`

- 目前只收斂了「錯誤怎麼顯示」
- 還沒有收斂「錯誤後怎麼重試 / 返回」與「哪些選單應 loop、哪些應 exit」這些更高層的互動規則

## 目前成本

- 新增一個小型 CLI UX 規則，也要同步修改很多 callsite
- 同類型輸入錯誤很容易因為漏改，變成不同選單行為不一致
- `main.go` 越來越像一個集中所有互動細節的巨型 orchestration 檔案
- 測試雖然可補，但驗證點散，後續擴充成本會持續上升

## 可行方案

### 方案 1：低風險，先收斂共用驗證與 selector helper

目標：不改整體流程，只減少散落邏輯。

可做的事：

- 把 region 選擇抽成 `selectRegion(scanner)` 類型 helper
- 把 mod 選擇、單值 index 驗證、range parsing 這類 CLI validator / selector 再往共用 helper 收斂
- 讓 `main.go` 保留 orchestration，但減少重複的 prompt + parse + validate pattern

適合原因：

- 風險低
- 可逐步做
- 不會一次打亂目前 CLI 操作習慣

### 方案 2：中風險，建立共用 prompt / selection pipeline

目標：把「顯示選項 → 讀輸入 → 判斷導航 → 驗證 → 錯誤回饋 → 成功返回結果」變成一套可重用流程。

可做的事：

- 建一層共用 CLI prompt helper
- 讓帳號選擇、flag 選擇、region 選擇、mod 選擇都走一致 API
- 把 retry / invalid input / nav handling 的一致性收進同一層

收益：

- 後續加 menu flow 時，不必重寫一套新的 scanner + if/return 組合
- 一致性更容易保證

風險：

- 需要調整多個現有 flow
- 若抽象設計過重，容易把簡單 menu 變難懂

### 方案 3：較大範圍，將 CLI flow 拆成 menu/domain 分層

目標：讓 `main.go` 不再同時扮演 menu renderer、validator、navigator、domain coordinator。

可做的事：

- 把 CLI flow 分成更清楚的檔案或 package，例如：
  - menu / prompt / feedback
  - multibox launch flow
  - account flag config flow
  - switcher config flow
- 讓 `main.go` 只負責主選單 dispatch

這是較完整的方向，但不建議立刻做，除非接下來還會持續擴充 CLI 功能。

## 建議順序

1. 先做方案 1
   - region / selector / validator 類 helper 收斂
   - 優先減少 `launchAccount()`、`launchAll()`、flag 設定 flow 的重複
2. 若後續還會新增更多 CLI flow，再進方案 2
   - 建 prompt / selection pipeline
3. 只有在 CLI 持續長大時，才進方案 3
   - 做更明確的 menu / domain 分層

## 暫時不做的事

- 不先重構主迴圈 dispatch
- 不碰 handle monitor、D2R 啟動核心、視窗重命名這些目前不是問題源頭的邏輯
- 不急著引入過重的 state machine / framework 式抽象

## 下一步

- 若要開始實作，建議從「抽出 region selector 與共用 selection helper」作為第一階段
- 若這輪要正式執行 refactor，先從低風險方案開始，不要一次跨到大範圍分層

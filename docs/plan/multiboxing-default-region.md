# 每帳號預設登入區域 server implementation plan

## 目標與玩家操作流程

這版規劃不只是把 region 選單多塞一個 Enter 快捷鍵，而是要明確補上一個可持久化的 per-account 預設登入區域流程：

1. 玩家先進入 CLI 的帳號預設 region 設定介面
2. 為每個 account 指派預設登入區域 server
3. 之後無論是啟動指定帳號，或用 `a` 批次啟動帳號，到了 region 選單都可以直接按 Enter
4. 若直接按 Enter，launcher 會改用該 account 已保存的 `DefaultRegion`
5. 若輸入 `1` / `2` / `3`，仍視為這次啟動的手動覆蓋，不改動帳號設定
6. 若按 Enter 時有任何目標帳號尚未設定 `DefaultRegion`，launcher 直接擋下並列出帳號，讓玩家改成手動選 region，或回去先補設定

也就是說，V1 的核心 UX 是：

- **先為每個 account 設好預設登入區域**
- **launch 時按 Enter 走帳號既有預設**
- **launch 時輸入 `1` / `2` / `3` 仍保留單次手動覆蓋**
- **若預設值缺漏，就明確擋下，不做 silent fallback**

## 已確認的關鍵技術事實

- 目前帳號資料模型在 [account.go](..\..\internal\multiboxing\account\account.go)；`accounts.csv` 現在是 `Email,Password,DisplayName,LaunchFlags,ToolFlags,GraphicsProfile,DefaultRegion` 七欄。
- 啟動切點在 [cli_launch.go](..\..\cmd\d2r-hyper-launcher\cli_launch.go) 很清楚：
  - 單帳號：`launchAccount()` 在 `selectLaunchMod()` 之後，會呼叫 `LaunchD2R()`
  - 批次：`launchAll()` 目前先選一次共用 region，再對每個 pending account 呼叫 `LaunchD2R()`
- 現有 region 定義集中在 [constants.go](..\..\internal\common\d2r\constants.go) 的 `Regions` / `FindRegion()`；目前 canonical name 是 `NA`、`EU`、`Asia`，address 則分別對應 Battle.net host。
- 現有 per-account 設定 UI 範本在 [cli_flags.go](..\..\cmd\d2r-hyper-launcher\cli_flags.go) 與 [cli_graphics_profiles.go](..\..\cmd\d2r-hyper-launcher\cli_graphics_profiles.go)；它們已經有成熟的 `runMenu` / `runMenuRead`、`b` / `h` / `q` 契約可沿用。
- schema 文字不只存在 [account.go](..\..\internal\multiboxing\account\account.go)，也散落在 [template.go](..\..\internal\multiboxing\account\template.go)、locale 的 first-run CSV 說明、[README.md](..\..\README.md)、[README.en.md](..\..\README.en.md)、[multiboxing-usage-guide.md](..\multiboxing-usage-guide.md)、[switcher-usage-guide.md](..\switcher-usage-guide.md) 與 [d2r-multiboxing skill](..\..\.claude\skills\d2r-multiboxing\SKILL.md)。

## 這次規劃鎖定的產品決策

### 1. `DefaultRegion` 存 region name，不存 address

新欄位暫定命名為 `DefaultRegion`，只存 canonical region name：

- `NA`
- `EU`
- `Asia`

不直接存 `us.actual.battle.net` 這種 address。這樣若未來 Battle.net address 有變，只需要更新 [constants.go](..\..\internal\common\d2r\constants.go) 的 region 映射，不必回頭 migrate 全部 `accounts.csv`。

### 2. Enter 是「用預設」，不是「猜一個合理 fallback」

當玩家在 region 選單直接按 Enter：

- 若目標帳號都有 `DefaultRegion`，就使用那些已保存的預設值
- 若有任何目標帳號尚未設定 `DefaultRegion`，就直接擋下並列出帳號

第一版明確不做：

- silent fallback 到 `NA`
- 自動沿用上一個手動選過的 region
- 幫未設定的帳號猜一個共用預設值

### 3. 手動 `1` / `2` / `3` 選擇仍保留單次覆蓋能力

這個功能不會取代現有的手動 region 選擇。

- 單帳號啟動時，輸入 `1` / `2` / `3` 仍是本次手動覆蓋
- `a` 批次啟動時，輸入 `1` / `2` / `3` 仍保留「全部目標帳號共用同一 region」的快速覆蓋能力
- 這些手動覆蓋都不會回寫到 `accounts.csv`

### 4. V1 先做「指派 / 清除」，不做更複雜的 region 管理器

因為 region 清單固定只有 3 個，所以第一版不需要像畫質設定檔那樣再建一層 profile store。

第一版優先完成：

- 指派預設登入區域給 account
- 清除 account 的預設登入區域
- 在 launch 時支援 Enter 使用既有預設值

先不把 scope 擴到：

- 自訂 Battle.net address
- 自訂更多 server 類型
- 依 mod / 情境切換不同預設 region

## 建議的資料與檔案設計

### 1. 帳號 schema

在 [account.go](..\..\internal\multiboxing\account\account.go) 的 `Account` 與 `accounts.csv` 增加新欄位：

- `DefaultRegion`

建議語意：

- 存 canonical region name
- 空字串表示未指派

CSV 目標格式：

`Email,Password,DisplayName,LaunchFlags,ToolFlags,GraphicsProfile,DefaultRegion`

向後相容要求：

- 舊 4 / 5 / 6 欄 CSV 仍可正常載入
- 載入舊資料時 `DefaultRegion` 自動視為空字串
- 寫回時統一升級成 7 欄

### 2. 載入時的資料清理

第一版建議比照 `LaunchFlags` 的 cleanup 思維處理無效 `DefaultRegion`：

- 若 `DefaultRegion` 不是合法的 region name，就當成空字串
- 清理後直接回寫 `accounts.csv`

這樣可以避免某份舊 CSV 或手動修改後的髒值，反覆卡在 launch path 上。

## 建議的 CLI UX

### 1. 帳號預設 region 設定入口

新增一個平行於 `g` 的主選單入口，語意類似：

- 「帳號預設登入區域」

畫面上要直接列出：

- 帳號列表
- 每個帳號目前的 default region 狀態

### 2. 設定 / 清除流程

這塊 UX 應明確對齊 [cli_flags.go](..\..\cmd\d2r-hyper-launcher\cli_flags.go) 與 [cli_graphics_profiles.go](..\..\cmd\d2r-hyper-launcher\cli_graphics_profiles.go) 的操作感，而不是另外發明一套互動模型。

外層至少提供：

- `1` 指派預設登入區域
- `2` 清除預設登入區域

其中「指派」模式建議提供兩種操作路徑：

- 依 region 選 account
- 依 account 選 region

這樣玩家可以：

- 先選某個 region，再一次套給多個帳號
- 或先選某個 account，再替它指定 region

「清除」模式則可以先做最核心版本：

- 依 account 清除目前指派

這塊 UI 要維持現有契約：

- 所有子選單都走 `runMenu` / `runMenuRead`
- `b` 返回
- `h` 回主選單
- `q` 離開

## 建議的 launch 行為

### 1. 單帳號 launch

在 [cli_launch.go](..\..\cmd\d2r-hyper-launcher\cli_launch.go) 的 `launchAccount()`：

1. 驗證 `D2R.exe`
2. 進入 region 選單
3. 若輸入 `1` / `2` / `3`，就以手動選擇的 region 啟動
4. 若直接按 Enter，則解析該帳號的 `DefaultRegion`
5. 若該帳號沒有 `DefaultRegion`，就提示錯誤並留在 region 選單
6. 再進入 `selectLaunchMod()` / `LaunchD2R()`

### 2. 批次 launch

在 `launchAll()`：

- 若輸入 `1` / `2` / `3`，就保留目前「全部 pending accounts 共用同一 region」的行為
- 若直接按 Enter，則每個 pending account 依自己的 `DefaultRegion` 啟動
- 若其中任何一個 pending account 沒有 `DefaultRegion`，就直接擋下並列出缺少設定的帳號，讓玩家改成手動選 region 或回去補設定

### 3. 實作切法建議

目前 [cli_launch.go](..\..\cmd\d2r-hyper-launcher\cli_launch.go) 的 `promptLaunchRegion()` 是回傳單一 `*d2r.Region`。

這次比較合理的切法是改成回傳兩種模式之一：

- 「本次手動 override region」
- 「改用 account 既有預設」

再由 `launchAccount()` / `launchAll()` 各自 resolve 最終 region，避免把 batch 的 per-account default 邏輯硬塞回目前的單一回傳值裡。

## 主要風險與限制

- `launchAll()` 現在的 region 選單是「選一次，全部共用」，改成支援 Enter 走 per-account defaults 後，batch path 的 region resolve 會比現在複雜。
- `DefaultRegion` 會讓 `accounts.csv` 從 6 欄變 7 欄，因此 starter template、first-run CSV 提示、README / docs / skill 只要漏掉一處就會出現說明與實作不一致。
- 若玩家手動編輯 `accounts.csv` 寫入無效 region name，而 loader 沒有清理，之後每次 launch 都會卡在同一份壞資料上。

## 建議的 implementation todos

- `default-region-schema`
  - 擴充 `Account` 與 `accounts.csv`，新增 `DefaultRegion`
  - 保持舊 4 / 5 / 6 欄 CSV 向後相容，存檔時升級為 7 欄
  - 同步 starter template 與 first-run CSV schema 說明
  - 加入無效 `DefaultRegion` 的 cleanup 與測試

- `default-region-management-ui`
  - 新增 player-facing CLI：
    - 指派預設登入區域
    - 清除預設登入區域
  - 提供依 region 選 account / 依 account 選 region 兩種路徑
  - 顯示 account 目前的 default region 狀態

- `default-region-launch-resolution`
  - 調整 region prompt，讓 Enter 代表「用預設」
  - 單帳號與 batch 各自 resolve 最終 region
  - 保留 `1` / `2` / `3` 單次覆蓋行為，不改寫帳號設定
  - 若 Enter 命中任何未設定 default 的帳號，就明確擋下並列出缺少設定的帳號

- `default-region-docs-tests`
  - 補齊 CSV 7 欄 / 6 欄 / 5 欄向後相容測試
  - 補齊 region assignment / clear / launch 行為測試
  - 更新 [README.md](..\..\README.md)、[README.en.md](..\..\README.en.md)、[multiboxing-usage-guide.md](..\multiboxing-usage-guide.md)、[switcher-usage-guide.md](..\switcher-usage-guide.md)
  - 同步更新 locale 的 schema 文案與 [d2r-multiboxing skill](..\..\.claude\skills\d2r-multiboxing\SKILL.md)

## 備註

- `DefaultRegion` 只是一個 launch-time 預設值，不會改變 Battle.net region 清單本身。
- 這個功能的價值在於減少重複選 region 的操作，而不是把手動 region 選擇拿掉。

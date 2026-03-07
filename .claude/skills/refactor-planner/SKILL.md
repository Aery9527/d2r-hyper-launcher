---
name: refactor-planner
description: "Create a reviewable refactor plan when code structure shows architecture-smell signals. Use this whenever the user asks whether the current design is a problem, wants a refactor or architecture plan, mentions duplicated logic, scattered same-scope behavior, low cohesion, leaky boundaries, or when implementing a task reveals you had to add the same or very similar code in multiple places. Finish the current task first, then analyze the structure and produce a proposal for review instead of silently performing a large refactor."
---

# Refactor Planning Workflow

這個 skill 的職責不是直接重構，而是把「結構性問題」整理成可 review 的改善方案。

它特別適合這種情境：

- 當前任務已經完成，但實作過程暴露出 duplication / low cohesion
- 使用者開始質疑目前結構是不是有問題
- 你意識到同一類邏輯散在很多地方，後續維護風險會很高

## 核心原則

1. **先完成當前任務**
2. **不要未經同意就直接展開大型 refactor**
3. **把架構問題變成一份可 review 的 plan**

這個 skill 的目標，是讓「交付需求」與「整理架構」變成兩段清楚的流程，而不是互相打架。

## 什麼情況算是架構訊號

以下任一情況，通常都表示值得啟動這個 skill：

### 1. 重複邏輯開始出現

- 為了完成一個需求，要在多個地方加入幾乎相同的 code
- 同一個 validation / mapping / guard / formatting / retry / fallback 被複製到不同入口
- 修一個 bug 時，必須靠「記得去好幾個地方一起改」才能保證一致

### 2. 同 scope 的 code 缺乏內聚力

- 一個明顯屬於同一個概念的行為，卻散落在多個檔案 / 函式 / menu flow
- 新功能每次都要橫跨很多入口才能接起來
- 邏輯邊界不清楚，導致同一件事有很多半重疊實作

### 3. 結構成本開始高於功能本身

- 加一個小功能，卻需要碰很多零散位置
- 測試難寫，因為行為散落到太多地方
- 文件容易過期，因為沒有一個明確的單一落點可對應

### 4. 使用者直接提出結構疑慮

- 「這樣的架構是不是有問題？」
- 「幫我規劃 refactor」
- 「這塊 code 太散了，整理一下」
- 「這是不是腐敗的程式碼？」

> 這不一定代表整個系統設計都失敗；通常表示「這一塊的抽象、邊界或模組劃分不夠好」。

## 腐敗 / 需要重整的常見範例

下面是值得觸發此 skill 的典型例子：

- **Shotgun surgery**：一個需求要改很多分散位置
- **Copy-paste growth**：新功能是靠重複貼上舊邏輯擴出來
- **Low cohesion**：同一個概念被拆散在很多不相鄰的地方
- **Leaky boundaries**：本該由單一模組負責的事，外面很多地方都在偷做
- **Inconsistent behavior risk**：同類行為很容易因漏改而不一致

如果你在任務中感受到「現在雖然能交付，但再這樣長下去會越來越難維護」，通常就是這個 skill 的時機。

## 執行順序

### 第一步：先交付當前任務

若使用者目前要的是 bug fix、feature、docs、commit、release 或其他具體交付：

1. 先把那個任務做完
2. 先做必要驗證
3. 再啟動這個 skill 產出 refactor plan

不要把「順便整理一下」擴張成未經允許的大重構。

### 第二步：蒐集證據

至少整理出：

- duplication 出現在哪些檔案 / 函式 / class / flow
- 哪些邏輯其實屬於同一個概念，卻分散在不同地方
- 這種分散造成哪些成本：
  - 容易漏改
  - 行為不一致
  - 測試難補
  - 文件難同步
  - 擴充成本高

### 第三步：必要時 fork agent

如果架構分析本身會很吃 context，或需要較完整的橫向閱讀：

- fork `explore` 或 `general-purpose` agent
- 讓子 agent 專注在：
  - duplication 地圖
  - 建議的共用抽象落點
  - 低風險 / 中風險 / 高風險方案

主 agent 保持目前交付結果清楚，不要把分析和實作混成一團。

## 預設輸出：plan，不直接實作

除非使用者明確說要動手 refactor，否則預設只產出 plan。

建議格式：

```markdown
## 為什麼這是架構訊號
- ...

## 問題落點
- 檔案 / 模組 A：...
- 檔案 / 模組 B：...

## 目前成本
- ...

## 可行方案
1. 低風險方案：...
2. 中風險方案：...
3. 較完整的重整方向：...

## 建議順序
1. ...
2. ...
3. ...

## 暫時不做的事
- ...
```

## 判斷與建議原則

- 以 **內聚力**、**單一責任**、**一致行為應該有單一落點** 為優先
- 不要為了抽象而抽象
- 若已有簡單共用 helper 就能明顯降低重複，先提低風險方案
- 若問題其實牽涉模組邊界，才提較大範圍重整
- 明確區分：
  - **現在一定要做的**
  - **適合下一輪做的**
  - **只是長期方向，不該現在動的**

## 回應使用者時要說清楚

1. 這是不是架構問題的訊號
2. 為什麼它不是單純一次性小改動
3. 目前已先完成哪個當前任務
4. 這份 refactor plan 想解決什麼
5. 這份 plan 目前只是 proposal，是否真的實作由使用者決定

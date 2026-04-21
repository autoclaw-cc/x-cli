package baidu

// filterOrganic removes "aladdin card" results (百科卡片、AI 回答卡片、相关搜索、
// 视频推荐等) and keeps only organic web results that have meaningful content.
//
// Baidu's SERP mixes organic results (tpl="www_index") with dozens of aladdin
// card types. Some carry useful info (百科 summaries), others are ads or
// empty recommendation stubs. The right filter depends on what the caller
// wants — so this is the decision the user should make.
//
// TODO(user): implement this filter. Suggested shape (5–10 lines):
//
//   1. Drop results whose Title is empty.
//   2. Decide: do you want to keep baidu-baike (tpl starts with "sg_kg_")?
//      They often have the cleanest summary of what a thing *is*.
//   3. Decide: drop or keep AI-generated answers (tpl="ai_agent_distribute")?
//      They can be useful but may hallucinate.
//   4. Always drop tpl="recommend_list" — it's the "people also searched"
//      card and has no per-result content.
//
// Consider these tpl values seen in the wild:
//   - "www_index"            → organic web result (keep)
//   - "sg_kg_entity_san"     → baidu baike entity card (usually keep)
//   - "ai_agent_distribute"  → AI answer card (your call)
//   - "recommend_list"       → related searches stub (drop)
//   - "se_com_default"       → fallback generic (usually keep)
//
// When you're done, set includeAll=true in cmd/root.go to bypass this
// filter if the user passes --all.
func filterOrganic(results []Result) []Result {
	// TODO(user): replace this pass-through with your filter logic.
	return results
}

package deepseek

type AnswerSnapshot struct {
	Count  int    `json:"count"`
	Latest string `json:"latest"`
}

type SubmitResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

const SnapshotScript = `
(() => {
  const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
  const txt=s=>(s||'').replace(/\s+/g,' ').trim();
  const body=txt(document.body?.innerText||'');
  const inputs=[...document.querySelectorAll('textarea,[contenteditable=true],input')]
    .filter(visible).slice(0,12).map(e=>({
      tag:e.tagName,
      role:e.getAttribute('role')||'',
      type:e.getAttribute('type')||'',
      aria:e.getAttribute('aria-label')||'',
      placeholder:e.getAttribute('placeholder')||'',
      text:txt(e.innerText||e.textContent||e.value).slice(0,120),
      disabled:!!e.disabled,
      cls:String(e.className).slice(0,120)
    }));
  const re=/DeepThink|R1|深度思考|联网|智能搜索|Search|搜索|上传|Attach|发送|Send|新对话|New chat|设置|Settings|模型|model|文件|识图|vision/i;
  const controls=[...document.querySelectorAll('button,[role=button],[role=radio],a,div,span')]
    .filter(e=>visible(e)&&re.test(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||'')))
    .slice(0,40).map(e=>({
      tag:e.tagName,
      role:e.getAttribute('role')||'',
      aria:e.getAttribute('aria-label')||'',
      text:txt(e.innerText||e.textContent).slice(0,120),
      disabled:!!e.disabled,
      cls:String(e.className).slice(0,120)
    }));
  return {href:location.href,title:document.title,bodyText:body.slice(0,600),hasPromptInput:inputs.length>0,inputs,controls};
})()
`

const AnswerSnapshotScript = `
(() => {
  const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
  const txt=s=>(s||'').replace(/\s+/g,' ').trim();
  const answers=[...document.querySelectorAll('.ds-assistant-message-main-content')]
    .filter(visible).map(e=>txt(e.innerText||e.textContent)).filter(Boolean);
  return {count:answers.length,latest:answers[answers.length-1]||''};
})()
`

const SubmitPromptScript = `
(() => {
  function deepseekSubmitPrompt() {
    const textarea=document.querySelector('textarea');
    if(!textarea) return {ok:false,error:'textarea_missing'};
    const tr=textarea.getBoundingClientRect();
    const candidates=[...document.querySelectorAll('[role=button],button')]
      .map(e=>({e,r:e.getBoundingClientRect(),text:(e.innerText||e.textContent||'').trim()}))
      .filter(x=>x.r.width>20&&x.r.height>20&&x.r.y>=tr.y&&x.r.x>tr.x+tr.width*0.65)
      .sort((a,b)=>b.r.x-a.r.x);
    const target=candidates[0];
    if(!target) return {ok:false,error:'send_button_missing'};
    target.e.click();
    return {ok:true};
  }
  return deepseekSubmitPrompt();
})()
`

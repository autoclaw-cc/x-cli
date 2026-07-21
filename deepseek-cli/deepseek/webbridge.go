package deepseek

import "fmt"

type AnswerSnapshot struct {
	Count  int    `json:"count"`
	Latest string `json:"latest"`
}

type SubmitResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type ModeResult struct {
	OK      bool     `json:"ok"`
	Error   string   `json:"error"`
	Enabled []string `json:"enabled"`
	Clicked []string `json:"clicked"`
}

type ModeReadyResult struct {
	Ready bool `json:"ready"`
}

type UploadTargetResult struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error"`
	Selector string `json:"selector"`
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

func SetModesScript(deepthink, search, vision bool) string {
	return fmt.Sprintf(`
(() => {
  function deepseekSetModes() {
    const desired={deepthink:%t,web_search:%t,vision:%t};
    const txt=s=>(s||'').replace(/\s+/g,' ').trim();
    const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
    const label=e=>txt(e.getAttribute('aria-label')||e.innerText||e.textContent);
    const toggles=[...document.querySelectorAll('[aria-pressed],button,[role=button]')]
      .filter(visible);
    const findToggle=patterns=>toggles.find(e=>patterns.some(re=>re.test(label(e))));
    const controls={
      deepthink:findToggle([/^(深度思考|DeepThink|Deep Think|R1)$/i]),
      web_search:findToggle([/^(智能搜索|联网搜索|联网|Web Search|Search)$/i]),
      vision:document.querySelector('[role=radio][data-model-type="vision"]')
    };
    const active=e=>{
      const pressed=e.getAttribute('aria-pressed')==='true';
      const checked=e.getAttribute('aria-checked')==='true';
      const state=(e.getAttribute('data-state')||'').toLowerCase();
      const cls=String(e.className||'').toLowerCase();
      return pressed||checked||/^(checked|selected|active)$/.test(state)||
        /(?:^|\s)ds-toggle-button--selected(?:\s|$)/.test(cls);
    };
    const enabled=[],clicked=[],missing=[];
    for (const [key,want] of Object.entries(desired)) {
      if(!want) continue;
      const target=controls[key];
      if(!target||!visible(target)){missing.push(key);continue;}
      if(!active(target)){target.click();clicked.push(key);}
      enabled.push(key);
    }
    if(missing.length) return {ok:false,error:'mode_control_missing:'+missing.join(','),enabled,clicked};
    return {ok:true,enabled,clicked};
  }
  return deepseekSetModes();
})()
`, deepthink, search, vision)
}

func ModeReadyScript(mode string) string {
	return fmt.Sprintf(`
(() => {
  function deepseekModeReady() {
    const mode=%q;
    if(mode==='vision') {
      const target=document.querySelector('[role=radio][data-model-type="vision"]');
      return {ready:target?.getAttribute('aria-checked')==='true'};
    }
    const wanted=mode==='deepthink'?/^(深度思考|DeepThink|Deep Think|R1)$/i:/^(智能搜索|联网搜索|联网|Web Search|Search)$/i;
    const clean=s=>(s||'').replace(/\s+/g,' ').trim();
    const target=[...document.querySelectorAll('[aria-pressed]')]
      .find(e=>wanted.test(clean(e.getAttribute('aria-label')||e.innerText||e.textContent)));
    return {ready:target?.getAttribute('aria-pressed')==='true'};
  }
  return deepseekModeReady();
})()
`, mode)
}

const PrepareUploadScript = `
(async () => {
  function deepseekPrepareUpload() {
    const txt=s=>(s||'').replace(/\s+/g,' ').trim();
    const visible=e=>!!(e&&(e.offsetWidth||e.offsetHeight||e.getClientRects().length));
    const selector='input[type=file]';
    let input=document.querySelector(selector);
    if(input) return {ok:true,selector};
    const control=[...document.querySelectorAll('button,[role=button],div,span')]
      .filter(visible)
      .find(e=>/上传|文件|Attach|Upload|File/i.test(txt(e.innerText||e.textContent)+' '+(e.getAttribute('aria-label')||'')));
    if(control) (control.closest('button,[role=button]')||control).click();
    return {ok:!!document.querySelector(selector),selector,error:document.querySelector(selector)?'':'file_input_missing'};
  }
  let result=deepseekPrepareUpload();
  if(result.ok) return result;
  await new Promise(resolve=>setTimeout(resolve,250));
  return deepseekPrepareUpload();
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
    const disabled=target.e.disabled||target.e.getAttribute('aria-disabled')==='true'||
      /disabled|loading/i.test(String(target.e.className||''));
    if(disabled) return {ok:false,error:'send_button_not_ready'};
    target.e.click();
    return {ok:true};
  }
  return deepseekSubmitPrompt();
})()
`
